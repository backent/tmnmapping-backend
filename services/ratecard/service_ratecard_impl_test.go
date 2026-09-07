package ratecard_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	service "github.com/malikabdulaziz/tmn-backend/services/ratecard"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webRateCard "github.com/malikabdulaziz/tmn-backend/web/ratecard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type deps struct {
	rateCard *mocks.MockRepositoryRateCard
	building *mocks.MockRepositoryBuilding
	pkg      *mocks.MockRepositorySalesPackage
}

func newService(t *testing.T, commits bool) (service.ServiceRateCardInterface, deps, func()) {
	t.Helper()

	db, sqlMock := testutil.NewMockDB(t)
	d := deps{
		rateCard: &mocks.MockRepositoryRateCard{},
		building: &mocks.MockRepositoryBuilding{},
		pkg:      &mocks.MockRepositorySalesPackage{},
	}

	sqlMock.ExpectBegin()
	if commits {
		sqlMock.ExpectCommit()
	} else {
		sqlMock.ExpectRollback()
	}

	svc := service.NewServiceRateCardImpl(db, d.rateCard, d.building, d.pkg)

	return svc, d, func() { assert.NoError(t, sqlMock.ExpectationsWereMet()) }
}

func draft(id int) models.RateCardVersion {
	return models.RateCardVersion{Id: id, VersionCode: "RC-2026-01", Status: models.RateCardStatusDraft, Currency: "IDR"}
}

func csv(lines ...string) []byte {
	out := ""
	for _, line := range lines {
		out += line + "\n"
	}

	return []byte(out)
}

// ---------------------------------------------------------------------------
// The editable-only rule
// ---------------------------------------------------------------------------

// Publishing freezes a version for good: approved quotations reference it and must
// never re-price. Every write goes through the same gate.
func TestWritesAreRefusedOnAPublishedVersion(t *testing.T) {
	for _, status := range []string{models.RateCardStatusCurrent, models.RateCardStatusHistorical} {
		t.Run(status, func(t *testing.T) {
			svc, d, assertMock := newService(t, false)

			version := draft(1)
			version.Status = status
			d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 1).Return(version, nil)

			assert.PanicsWithValue(t,
				exceptions.NewBadRequestError("this rate card is "+status+
					" and can no longer be changed; create a new draft instead"),
				func() {
					svc.UpsertBuildingPrice(context.Background(), 1,
						webRateCard.UpsertBuildingPriceRequest{BuildingId: 5, PriceIdrPer4Weeks: 1000})
				})

			d.rateCard.AssertNotCalled(t, "UpsertBuildingPrice", mock.Anything, mock.Anything, mock.Anything)
			assertMock()
		})
	}
}

func TestDeleteVersion_RefusedOncePublished(t *testing.T) {
	svc, d, assertMock := newService(t, false)

	version := draft(1)
	version.Status = models.RateCardStatusCurrent
	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 1).Return(version, nil)

	assert.Panics(t, func() { svc.DeleteVersion(context.Background(), 1) })
	d.rateCard.AssertNotCalled(t, "DeleteVersion", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

// ---------------------------------------------------------------------------
// Publishing
// ---------------------------------------------------------------------------

func TestPublish_DemotesCurrentAndFreezesPackageComposition(t *testing.T) {
	svc, d, assertMock := newService(t, true)

	version := draft(2)
	version.BuildingPriceCount = 3
	version.PackagePriceCount = 1

	published := version
	published.Status = models.RateCardStatusCurrent

	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 2).Return(version, nil).Once()
	d.rateCard.On("FindPackagePrices", mock.Anything, mock.Anything, 2, 100000, 0, "").
		Return([]models.RateCardPackagePrice{{SalesPackageId: 9, SalesPackageName: "Jakarta CBD"}}, nil)
	d.rateCard.On("FindLivePackageBuildingIds", mock.Anything, mock.Anything, 9).Return([]int{11, 12}, nil)
	d.rateCard.On("ReplacePackageBuildings", mock.Anything, mock.Anything, 2, 9, []int{11, 12}).Return(nil)
	d.rateCard.On("DemoteCurrentVersion", mock.Anything, mock.Anything).Return(nil)
	d.rateCard.On("SetVersionStatus", mock.Anything, mock.Anything, 2, models.RateCardStatusCurrent, 7).Return(nil)
	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 2).Return(published, nil)

	resp := svc.PublishVersion(context.Background(), 2, 7)

	assert.Equal(t, models.RateCardStatusCurrent, resp.Status)
	assert.False(t, resp.IsEditable)
	d.rateCard.AssertExpectations(t)
	assertMock()
}

// An empty rate card would leave the quotation wizard unable to price anything.
func TestPublish_RefusesAVersionWithNoPrices(t *testing.T) {
	svc, d, assertMock := newService(t, false)

	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 2).Return(draft(2), nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("this rate card has no prices; add some before publishing"),
		func() { svc.PublishVersion(context.Background(), 2, 7) })

	d.rateCard.AssertNotCalled(t, "DemoteCurrentVersion", mock.Anything, mock.Anything)
	assertMock()
}

// Pricing a package that contains nothing would produce a quotation line with no
// buildings behind it.
func TestPublish_RefusesAPricedPackageWithNoBuildings(t *testing.T) {
	svc, d, assertMock := newService(t, false)

	version := draft(2)
	version.PackagePriceCount = 1

	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 2).Return(version, nil)
	d.rateCard.On("FindPackagePrices", mock.Anything, mock.Anything, 2, 100000, 0, "").
		Return([]models.RateCardPackagePrice{{SalesPackageId: 9, SalesPackageName: "Empty Package"}}, nil)
	d.rateCard.On("FindLivePackageBuildingIds", mock.Anything, mock.Anything, 9).Return([]int{}, nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("sales package \"Empty Package\" is priced but contains no buildings"),
		func() { svc.PublishVersion(context.Background(), 2, 7) })

	d.rateCard.AssertNotCalled(t, "DemoteCurrentVersion", mock.Anything, mock.Anything)
	assertMock()
}

// ---------------------------------------------------------------------------
// Versions
// ---------------------------------------------------------------------------

func TestCreateVersion_CopiesPricesFromAnExistingVersion(t *testing.T) {
	svc, d, assertMock := newService(t, true)

	d.rateCard.On("FindVersionByCode", mock.Anything, mock.Anything, "RC-2027").
		Return(models.RateCardVersion{}, sql.ErrNoRows)
	d.rateCard.On("CreateVersion", mock.Anything, mock.Anything, mock.AnythingOfType("models.RateCardVersion")).
		Return(models.RateCardVersion{Id: 5}, nil)
	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 3).
		Return(models.RateCardVersion{Id: 3, Status: models.RateCardStatusCurrent}, nil)
	d.rateCard.On("CopyPrices", mock.Anything, mock.Anything, 3, 5).Return(nil)
	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 5).Return(draft(5), nil)

	resp := svc.CreateVersion(context.Background(), webRateCard.CreateVersionRequest{
		VersionCode: "RC-2027", CopyFromVersionId: 3,
	})

	assert.Equal(t, 5, resp.Id)
	assert.True(t, resp.IsEditable)
	d.rateCard.AssertExpectations(t)
	assertMock()
}

func TestCreateVersion_RejectsDuplicateCode(t *testing.T) {
	svc, d, assertMock := newService(t, false)

	d.rateCard.On("FindVersionByCode", mock.Anything, mock.Anything, "RC-2026-01").
		Return(models.RateCardVersion{Id: 1}, nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("that version code is already in use"),
		func() {
			svc.CreateVersion(context.Background(), webRateCard.CreateVersionRequest{VersionCode: "RC-2026-01"})
		})

	assertMock()
}

func TestFindCurrentVersion_WhenNothingPublished(t *testing.T) {
	svc, d, assertMock := newService(t, false)

	d.rateCard.On("FindCurrentVersion", mock.Anything, mock.Anything).
		Return(models.RateCardVersion{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NewNotFoundError("no rate card has been published yet"),
		func() { svc.FindCurrentVersion(context.Background()) })

	assertMock()
}

// ---------------------------------------------------------------------------
// Import
// ---------------------------------------------------------------------------

func TestBuildingPriceTemplate_RoundTripsThroughTheImporter(t *testing.T) {
	svc, _, _ := newService(t, true)

	fileBytes, err := svc.BuildingPriceTemplate()
	require.NoError(t, err)

	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, "xlsx")
	require.NoError(t, err)

	colMap := spreadsheets.MapHeaderColumns(rows[0], service.BuildingPriceColumns)
	assert.Empty(t, spreadsheets.MissingRequiredColumns(colMap, service.BuildingPriceColumns))
}

func TestImportBuildingPrices_HappyPath(t *testing.T) {
	svc, d, assertMock := newService(t, true)

	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 1).Return(draft(1), nil)
	d.building.On("FindByExternalId", mock.Anything, mock.Anything, "BLDG-001").
		Return(models.Building{Id: 11}, nil)
	d.rateCard.On("CountBuildingPrices", mock.Anything, mock.Anything, 1, "").Return(0, nil).Once()

	var captured models.RateCardBuildingPrice
	d.rateCard.On("UpsertBuildingPrice", mock.Anything, mock.Anything, mock.AnythingOfType("models.RateCardBuildingPrice")).
		Run(func(args mock.Arguments) { captured = args.Get(2).(models.RateCardBuildingPrice) }).
		Return(nil)
	d.rateCard.On("CountBuildingPrices", mock.Anything, mock.Anything, 1, "").Return(1, nil).Once()

	result := svc.ImportBuildingPrices(context.Background(), 1, csv(
		"External Building ID,Building Name,Price per 4 Weeks (IDR)",
		"BLDG-001,Menara BCA,92000000",
	), "csv")

	assert.True(t, result.Imported)
	assert.Equal(t, 1, result.Created)
	assert.Equal(t, int64(92000000), captured.PriceIdrPer4Weeks)
	assertMock()
}

// People paste numbers straight out of Excel, commas and all.
func TestImportBuildingPrices_AcceptsFormattedNumbers(t *testing.T) {
	svc, d, assertMock := newService(t, true)

	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 1).Return(draft(1), nil)
	d.building.On("FindByExternalId", mock.Anything, mock.Anything, "BLDG-001").
		Return(models.Building{Id: 11}, nil)
	d.rateCard.On("CountBuildingPrices", mock.Anything, mock.Anything, 1, "").Return(0, nil).Once()

	var captured models.RateCardBuildingPrice
	d.rateCard.On("UpsertBuildingPrice", mock.Anything, mock.Anything, mock.AnythingOfType("models.RateCardBuildingPrice")).
		Run(func(args mock.Arguments) { captured = args.Get(2).(models.RateCardBuildingPrice) }).
		Return(nil)
	d.rateCard.On("CountBuildingPrices", mock.Anything, mock.Anything, 1, "").Return(1, nil).Once()

	result := svc.ImportBuildingPrices(context.Background(), 1, csv(
		"External Building ID,Price per 4 Weeks (IDR)",
		"BLDG-001,\"92,000,000.00\"",
	), "csv")

	assert.True(t, result.Imported)
	assert.Equal(t, int64(92000000), captured.PriceIdrPer4Weeks)
	assertMock()
}

// Rupiah has no minor unit, so a real fraction is a mistake rather than something
// to round silently.
func TestImportBuildingPrices_RejectsFractionalRupiah(t *testing.T) {
	svc, d, assertMock := newService(t, true)

	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 1).Return(draft(1), nil)

	result := svc.ImportBuildingPrices(context.Background(), 1, csv(
		"External Building ID,Price per 4 Weeks (IDR)",
		"BLDG-001,92000000.55",
	), "csv")

	assert.False(t, result.Imported)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "whole number of rupiah")
	d.rateCard.AssertNotCalled(t, "UpsertBuildingPrice", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

func TestImportBuildingPrices_RejectsUnknownBuilding(t *testing.T) {
	svc, d, assertMock := newService(t, true)

	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 1).Return(draft(1), nil)
	d.building.On("FindByExternalId", mock.Anything, mock.Anything, "BLDG-999").
		Return(models.Building{}, sql.ErrNoRows)

	result := svc.ImportBuildingPrices(context.Background(), 1, csv(
		"External Building ID,Price per 4 Weeks (IDR)",
		"BLDG-999,1000",
	), "csv")

	assert.False(t, result.Imported)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "No building with this ID")
	assertMock()
}

func TestImportBuildingPrices_RejectsTheSameBuildingTwice(t *testing.T) {
	svc, d, assertMock := newService(t, true)

	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 1).Return(draft(1), nil)
	d.building.On("FindByExternalId", mock.Anything, mock.Anything, "BLDG-001").
		Return(models.Building{Id: 11}, nil)

	result := svc.ImportBuildingPrices(context.Background(), 1, csv(
		"External Building ID,Price per 4 Weeks (IDR)",
		"BLDG-001,1000",
		"BLDG-001,2000",
	), "csv")

	assert.False(t, result.Imported)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "priced twice")
	assertMock()
}

func TestImportPackagePrices_RejectsUnknownPackage(t *testing.T) {
	svc, d, assertMock := newService(t, true)

	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 1).Return(draft(1), nil)
	d.pkg.On("FindByNames", mock.Anything, mock.Anything, []string{"Nope"}).
		Return([]models.SalesPackage{}, nil)

	result := svc.ImportPackagePrices(context.Background(), 1, csv(
		"Sales Package,Price per 4 Weeks (IDR)",
		"Nope,1000",
	), "csv")

	assert.False(t, result.Imported)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "No sales package with this name")
	assertMock()
}

// An upload into a published version must be refused before any row is considered.
func TestImportBuildingPrices_RefusedOnAPublishedVersion(t *testing.T) {
	svc, d, assertMock := newService(t, false)

	version := draft(1)
	version.Status = models.RateCardStatusCurrent
	d.rateCard.On("FindVersionById", mock.Anything, mock.Anything, 1).Return(version, nil)

	assert.Panics(t, func() {
		svc.ImportBuildingPrices(context.Background(), 1, csv(
			"External Building ID,Price per 4 Weeks (IDR)",
			"BLDG-001,1000",
		), "csv")
	})

	d.rateCard.AssertNotCalled(t, "UpsertBuildingPrice", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}
