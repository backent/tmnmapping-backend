package quotation_test

import (
	"context"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	webQuotation "github.com/malikabdulaziz/tmn-backend/web/quotation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Selecting buildings in bulk is the point of these tests: a seller can now filter
// and take every match at once, so a selection of hundreds has to cost the same
// number of round trips as a selection of one.

func buildingAt(id int, name string, audience, impression int) models.Building {
	return models.Building{
		Id: id, Name: name, IrisCode: "IRIS-" + name, BuildingType: "Apartment",
		Citytown: "Jakarta", Audience: audience, Impression: impression,
	}
}

func previewRequest(ids ...int) webQuotation.PricingPreviewRequest {
	return webQuotation.PricingPreviewRequest{
		Discount: 50,
		Placement: &webQuotation.SelectionRequest{
			Mode: models.SelectionModeBuilding, BuildingIds: ids,
			TvcDurationSeconds: 15, Weeks: 4, Spots: 180,
		},
	}
}

// The whole reason for the rewrite: two queries, whatever the selection size.
func TestBuildSelection_ReadsEveryBuildingInOneRoundTrip(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, true)

	ids := make([]int, 0, 300)
	buildings := make([]models.Building, 0, 300)
	prices := map[int]int64{}
	for id := 1; id <= 300; id++ {
		ids = append(ids, id)
		buildings = append(buildings, buildingAt(id, "B", 10, 100))
		prices[id] = 1_000_000
	}

	d.building.On("FindByIds", mock.Anything, mock.Anything, ids).Return(buildings, nil).Once()
	d.buildingPrice.On("FindPricesByBuildingIds", mock.Anything, mock.Anything, ids).Return(prices, nil).Once()
	d.user.On("FindByRole", mock.Anything, mock.Anything, models.RoleHeadOfSales).
		Return([]models.User{{Id: 999, Name: "Head"}}, nil)
	d.user.On("FindById", mock.Anything, mock.Anything, 999).Return(models.User{Id: 999, Name: "Head"}, nil)

	response := svc.PreviewPricing(context.Background(), previewRequest(ids...), salesActor(111))

	// 300 buildings at 1,000,000/week for 4 weeks.
	assert.Equal(t, int64(300*1_000_000*4), response.Pricing.PlacementGross)
	assert.Equal(t, 300, response.Sections[0].ScreenCount)

	// Once() on both lookups is the assertion that matters: the old loop called
	// FindById and FindByBuildingId once per building, so this would have been 600.
	d.building.AssertNumberOfCalls(t, "FindByIds", 1)
	d.buildingPrice.AssertNumberOfCalls(t, "FindPricesByBuildingIds", 1)
	d.buildingPrice.AssertNotCalled(t, "FindByBuildingId", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

// A bulk "select all filtered" makes a repeated id easy to send by accident. Charging
// for it twice would be the worst possible way to find out.
func TestBuildSelection_ChargesADuplicateBuildingOnce(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, true)

	unique := []int{7, 8}
	d.building.On("FindByIds", mock.Anything, mock.Anything, unique).
		Return([]models.Building{buildingAt(7, "Seven", 10, 100), buildingAt(8, "Eight", 20, 200)}, nil).Once()
	d.buildingPrice.On("FindPricesByBuildingIds", mock.Anything, mock.Anything, unique).
		Return(map[int]int64{7: 1_000_000, 8: 2_000_000}, nil).Once()
	d.user.On("FindByRole", mock.Anything, mock.Anything, models.RoleHeadOfSales).
		Return([]models.User{{Id: 999, Name: "Head"}}, nil)
	d.user.On("FindById", mock.Anything, mock.Anything, 999).Return(models.User{Id: 999, Name: "Head"}, nil)

	response := svc.PreviewPricing(context.Background(), previewRequest(7, 8, 7, 8, 7), salesActor(111))

	assert.Equal(t, int64((1_000_000+2_000_000)*4), response.Pricing.PlacementGross)
	assert.Equal(t, 2, response.Sections[0].ScreenCount)
	assert.Len(t, response.Sections[0].Items, 2)
	assert.Equal(t, int64(30), response.Sections[0].Traffic, "audience counted once per building")
	assertMock()
}

// The order the seller picked is the order the document prints.
func TestBuildSelection_KeepsTheOrderTheClientSent(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, true)

	ids := []int{3, 1, 2}
	// Deliberately returned in a different order: an IN (...) query makes no promise.
	d.building.On("FindByIds", mock.Anything, mock.Anything, ids).Return([]models.Building{
		buildingAt(1, "One", 1, 1), buildingAt(2, "Two", 1, 1), buildingAt(3, "Three", 1, 1),
	}, nil).Once()
	d.buildingPrice.On("FindPricesByBuildingIds", mock.Anything, mock.Anything, ids).
		Return(map[int]int64{1: 1_000, 2: 1_000, 3: 1_000}, nil).Once()
	d.user.On("FindByRole", mock.Anything, mock.Anything, models.RoleHeadOfSales).
		Return([]models.User{{Id: 999, Name: "Head"}}, nil)
	d.user.On("FindById", mock.Anything, mock.Anything, 999).Return(models.User{Id: 999, Name: "Head"}, nil)

	response := svc.PreviewPricing(context.Background(), previewRequest(ids...), salesActor(111))

	assert.Equal(t, []string{"Three", "One", "Two"}, []string{
		response.Sections[0].Items[0].BuildingName,
		response.Sections[0].Items[1].BuildingName,
		response.Sections[0].Items[2].BuildingName,
	})
	assertMock()
}

// A missing row used to come back as sql.ErrNoRows from FindById. The bulk query just
// omits it, so absence has to be checked explicitly.
func TestBuildSelection_RefusesAnIdThatNoLongerExists(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, false)

	ids := []int{1, 404}
	d.building.On("FindByIds", mock.Anything, mock.Anything, ids).
		Return([]models.Building{buildingAt(1, "One", 1, 1)}, nil).Once()
	d.buildingPrice.On("FindPricesByBuildingIds", mock.Anything, mock.Anything, ids).
		Return(map[int]int64{1: 1_000}, nil).Once()

	assert.PanicsWithValue(t, exceptions.NewBadRequestError("building not found"), func() {
		svc.PreviewPricing(context.Background(), previewRequest(ids...), salesActor(111))
	})

	assertMock()
}

// An unpriced building must still be named, so the seller knows which one to fix.
func TestBuildSelection_NamesTheUnpricedBuilding(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, false)

	ids := []int{1, 2}
	d.building.On("FindByIds", mock.Anything, mock.Anything, ids).Return([]models.Building{
		buildingAt(1, "One", 1, 1), buildingAt(2, "Menara Kosong", 1, 1),
	}, nil).Once()
	d.buildingPrice.On("FindPricesByBuildingIds", mock.Anything, mock.Anything, ids).
		Return(map[int]int64{1: 1_000}, nil).Once()

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("\"Menara Kosong\" has no price yet, so it cannot be quoted"),
		func() { svc.PreviewPricing(context.Background(), previewRequest(ids...), salesActor(111)) })

	assertMock()
}

func TestBuildSelection_RefusesAnEmptySelection(t *testing.T) {
	svc, _, assertMock := newQuotationService(t, false)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("a building selection needs at least one building"),
		func() { svc.PreviewPricing(context.Background(), previewRequest(), salesActor(111)) })

	assertMock()
}
