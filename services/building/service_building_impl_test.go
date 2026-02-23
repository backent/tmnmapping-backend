package building_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	serviceBuilding "github.com/malikabdulaziz/tmn-backend/services/building"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webBuilding "github.com/malikabdulaziz/tmn-backend/web/building"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newBuildingService wires up ServiceBuildingImpl for tests.
// ERPClient is nil — SyncFromERP tests are out of scope (ERPClient is a concrete type).
func newBuildingService(
	db *sql.DB,
	repoBuilding *mocks.MockRepositoryBuilding,
	repoPOI *mocks.MockRepositoryPOI,
) serviceBuilding.ServiceBuildingInterface {
	logger := logrus.New()
	logger.SetOutput(io.Discard) // suppress log output during tests
	return serviceBuilding.NewServiceBuildingImpl(db, repoBuilding, repoPOI, nil, logger)
}

// --- FindById ---

// TestFindById_HappyPath verifies that a found building is returned as a
// correctly-mapped BuildingResponse.
func TestFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoBuilding := &mocks.MockRepositoryBuilding{}
	repoPOI := &mocks.MockRepositoryPOI{}
	svc := newBuildingService(db, repoBuilding, repoPOI)

	building := testutil.NewBuilding(10, "Grand Indonesia")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 10).
		Return(building, nil)

	response := svc.FindById(context.Background(), 10)

	assert.Equal(t, 10, response.Id)
	assert.Equal(t, "Grand Indonesia", response.Name)
	assert.Equal(t, "Office", response.BuildingType)
	assert.Equal(t, "A", response.GradeResource)
	assert.InDelta(t, -6.2, response.Latitude, 0.001)
	assert.InDelta(t, 106.8, response.Longitude, 0.001)

	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestFindById_NotFound verifies that sql.ErrNoRows produces a NotFoundError panic.
func TestFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoBuilding := &mocks.MockRepositoryBuilding{}
	repoPOI := &mocks.MockRepositoryPOI{}
	svc := newBuildingService(db, repoBuilding, repoPOI)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(testutil.NewBuilding(0, ""), sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "building not found"},
		func() { svc.FindById(context.Background(), 999) },
	)

	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestFindById_RepositoryError verifies that an unexpected repository error propagates as a panic.
func TestFindById_RepositoryError(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoBuilding := &mocks.MockRepositoryBuilding{}
	repoPOI := &mocks.MockRepositoryPOI{}
	svc := newBuildingService(db, repoBuilding, repoPOI)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return(testutil.NewBuilding(0, ""), errors.New("connection reset by peer"))

	assert.Panics(t, func() { svc.FindById(context.Background(), 5) })

	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Update ---

// TestUpdate_HappyPath verifies that existing building fields are updated
// and the updated response is returned.
func TestUpdate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoBuilding := &mocks.MockRepositoryBuilding{}
	repoPOI := &mocks.MockRepositoryPOI{}
	svc := newBuildingService(db, repoBuilding, repoPOI)

	existing := testutil.NewBuilding(7, "Central Park")
	existing.Sellable = "not_sell"
	existing.Connectivity = "manual"
	existing.ResourceType = ""

	updated := existing
	updated.Sellable = "sell"
	updated.Connectivity = "online"
	updated.ResourceType = "LCD"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 7).
		Return(existing, nil)

	// The service modifies existingBuilding in place before passing to Update.
	// Use mock.MatchedBy to verify the correct field values are sent to the repository.
	repoBuilding.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.MatchedBy(func(b models.Building) bool {
		return b.Sellable == "sell" && b.Connectivity == "online" && b.ResourceType == "LCD"
	})).Return(updated, nil)

	request := webBuilding.UpdateBuildingRequest{
		Sellable:     "sell",
		Connectivity: "online",
		ResourceType: "LCD",
	}

	response := svc.Update(context.Background(), request, 7)

	assert.Equal(t, 7, response.Id)
	assert.Equal(t, "sell", response.Sellable)
	assert.Equal(t, "online", response.Connectivity)
	assert.Equal(t, "LCD", response.ResourceType)

	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestUpdate_NotFound verifies that updating a non-existent building panics with NotFoundError.
func TestUpdate_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoBuilding := &mocks.MockRepositoryBuilding{}
	repoPOI := &mocks.MockRepositoryPOI{}
	svc := newBuildingService(db, repoBuilding, repoPOI)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(testutil.NewBuilding(0, ""), sql.ErrNoRows)

	request := webBuilding.UpdateBuildingRequest{Sellable: "sell"}

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "building not found"},
		func() { svc.Update(context.Background(), request, 404) },
	)

	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindAll ---

// TestFindAll_HappyPath verifies that a list of buildings is returned with the correct count.
func TestFindAll_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoBuilding := &mocks.MockRepositoryBuilding{}
	repoPOI := &mocks.MockRepositoryPOI{}
	svc := newBuildingService(db, repoBuilding, repoPOI)

	buildingList := []models.Building{
		testutil.NewBuilding(1, "Building A"),
		testutil.NewBuilding(2, "Building B"),
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	// Use mock.Anything for all filter parameters — the test verifies transformation, not filter logic.
	repoBuilding.On("FindAll",
		mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(buildingList, nil)

	repoBuilding.On("CountAll",
		mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(2, nil)

	var request webBuilding.BuildingRequestFindAll
	request.SetTake(10)
	request.SetSkip(0)

	responses, total := svc.FindAll(context.Background(), request)

	assert.Equal(t, 2, total)
	assert.Len(t, responses, 2)
	assert.Equal(t, 1, responses[0].Id)
	assert.Equal(t, "Building A", responses[0].Name)
	assert.Equal(t, 2, responses[1].Id)
	assert.Equal(t, "Building B", responses[1].Name)

	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
