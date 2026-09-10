package buildingprice_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	service "github.com/malikabdulaziz/tmn-backend/services/buildingprice"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newService(t *testing.T) (service.ServiceBuildingPriceInterface, *mocks.MockRepositoryBuildingPrice, *mocks.MockRepositoryBuilding) {
	t.Helper()

	db, sqlMock := testutil.NewMockDB(t)
	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	prices := &mocks.MockRepositoryBuildingPrice{}
	buildings := &mocks.MockRepositoryBuilding{}

	return service.NewServiceBuildingPriceImpl(db, prices, buildings), prices, buildings
}

func knownBuildings(b *mocks.MockRepositoryBuilding, codes map[string]int) {
	for code, id := range codes {
		b.On("FindByIrisCode", mock.Anything, mock.Anything, code).Return(models.Building{Id: id}, nil)
	}
}

const threeRows = "IRIS Building ID,Price per Week (IDR)\nB1,1000\nB2,2500\nB3,3000\n"

// An upload is previewed before it is applied. The preview must count every row the
// same way the apply will -- and write nothing.
func TestImport_DryRunCountsWithoutWriting(t *testing.T) {
	svc, prices, buildings := newService(t)
	prices.On("FindAllPrices", mock.Anything, mock.Anything).Return(map[int]int64{1: 1000, 2: 2000}, nil)
	knownBuildings(buildings, map[string]int{"B1": 1, "B2": 2, "B3": 3})

	result := svc.Import(context.Background(), []byte(threeRows), "csv", true)

	assert.True(t, result.DryRun)
	assert.False(t, result.Imported)
	assert.Empty(t, result.Errors)
	assert.Equal(t, 1, result.Created, "B3 has no price yet")
	assert.Equal(t, 1, result.Updated, "B2 changes 2,000 -> 2,500")
	assert.Equal(t, 1, result.Unchanged, "B1 already costs 1,000")
	prices.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// Applying writes only what changes. An unchanged row is not rewritten, so its
// updated_at still means "when this price last changed".
func TestImport_ApplyWritesOnlyChangedRows(t *testing.T) {
	svc, prices, buildings := newService(t)
	prices.On("FindAllPrices", mock.Anything, mock.Anything).Return(map[int]int64{1: 1000, 2: 2000}, nil)
	knownBuildings(buildings, map[string]int{"B1": 1, "B2": 2, "B3": 3})
	prices.On("Upsert", mock.Anything, mock.Anything, 2, int64(2500)).Return(nil)
	prices.On("Upsert", mock.Anything, mock.Anything, 3, int64(3000)).Return(nil)

	result := svc.Import(context.Background(), []byte(threeRows), "csv", false)

	assert.True(t, result.Imported)
	assert.False(t, result.DryRun)
	prices.AssertNumberOfCalls(t, "Upsert", 2)
	prices.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, 1, mock.Anything)
}

// The business's own rate card workbook names the price column "Round Up" and has
// a "Pick Building Rate" beside it. It must import as-is.
func TestImport_AcceptsTheRateCardWorkbookAsItIs(t *testing.T) {
	svc, prices, buildings := newService(t)
	prices.On("FindAllPrices", mock.Anything, mock.Anything).Return(map[int]int64{}, nil)
	knownBuildings(buildings, map[string]int{"B000006": 6})

	workbook := "Building Type,IRIS Building ID,Building Name,Pick Building Rate,Round Up\n" +
		"Apartment,B000006,Kubikahomy Apartment - Tower,1083600,1100000\n"
	result := svc.Import(context.Background(), []byte(workbook), "csv", true)

	assert.Empty(t, result.Errors)
	assert.Equal(t, 1, result.Created)
}

// One bad row refuses the whole file, so a price list is never half-applied.
func TestImport_RefusesTheFileOnAnUnknownBuildingAndSkipsZeroPrices(t *testing.T) {
	svc, prices, buildings := newService(t)
	prices.On("FindAllPrices", mock.Anything, mock.Anything).Return(map[int]int64{}, nil)
	buildings.On("FindByIrisCode", mock.Anything, mock.Anything, "NOPE").Return(models.Building{}, sql.ErrNoRows)

	result := svc.Import(context.Background(),
		[]byte("IRIS Building ID,Price per Week (IDR)\nB1,0\nNOPE,5000\n"), "csv", false)

	assert.False(t, result.Imported)
	assert.Equal(t, 1, result.Skipped, "a zero price is skipped, not sold for nothing")
	assert.Len(t, result.Errors, 1)
	prices.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
