package buildingrestriction_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	serviceRestriction "github.com/malikabdulaziz/tmn-backend/services/buildingrestriction"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webRestriction "github.com/malikabdulaziz/tmn-backend/web/buildingrestriction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newRestrictionService(
	db *sql.DB,
	repoRestriction *mocks.MockRepositoryBuildingRestriction,
	repoBuilding *mocks.MockRepositoryBuilding,
) serviceRestriction.ServiceBuildingRestrictionInterface {
	return serviceRestriction.NewServiceBuildingRestrictionImpl(db, repoRestriction, repoBuilding)
}

func newRestrictionModel(id int, name string, buildingRefs ...models.BuildingRef) models.BuildingRestriction {
	return models.BuildingRestriction{Id: id, Name: name, Buildings: buildingRefs}
}

// --- Create ---

func TestRestrictionCreate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoRestriction := &mocks.MockRepositoryBuildingRestriction{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newRestrictionService(db, repoRestriction, repoBuilding)

	bldg := testutil.NewBuilding(5, "Office A")
	created := newRestrictionModel(1, "Zone 1", models.BuildingRef{Id: 5, Name: "Office A"})

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return(bldg, nil)
	repoRestriction.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(r models.BuildingRestriction) bool { return r.Name == "Zone 1" }),
		[]int{5},
	).Return(created, nil)

	request := webRestriction.CreateBuildingRestrictionRequest{
		Name:        "Zone 1",
		BuildingIds: []int{5},
	}

	response := svc.Create(context.Background(), request)

	assert.Equal(t, 1, response.Id)
	assert.Equal(t, "Zone 1", response.Name)
	assert.Len(t, response.Buildings, 1)
	assert.Equal(t, 5, response.Buildings[0].Id)

	repoRestriction.AssertExpectations(t)
	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestRestrictionCreate_InvalidBuilding verifies that a non-existent building ID
// causes a BadRequestError panic before the restriction is created.
func TestRestrictionCreate_InvalidBuilding(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoRestriction := &mocks.MockRepositoryBuildingRestriction{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newRestrictionService(db, repoRestriction, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.Building{}, sql.ErrNoRows)

	request := webRestriction.CreateBuildingRestrictionRequest{
		Name:        "Bad Zone",
		BuildingIds: []int{999},
	}

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "building not found"},
		func() { svc.Create(context.Background(), request) },
	)

	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestRestrictionCreate_MultipleBuildings verifies that ALL building IDs are
// validated and all must exist.
func TestRestrictionCreate_MultipleBuildings_FirstInvalid(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoRestriction := &mocks.MockRepositoryBuildingRestriction{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newRestrictionService(db, repoRestriction, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	// First building does not exist — validation stops here.
	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 1).
		Return(models.Building{}, sql.ErrNoRows)

	request := webRestriction.CreateBuildingRestrictionRequest{
		Name:        "Multi Zone",
		BuildingIds: []int{1, 2},
	}

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "building not found"},
		func() { svc.Create(context.Background(), request) },
	)

	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindById ---

func TestRestrictionFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoRestriction := &mocks.MockRepositoryBuildingRestriction{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newRestrictionService(db, repoRestriction, repoBuilding)

	restriction := newRestrictionModel(4, "North Zone", models.BuildingRef{Id: 1, Name: "HQ"})

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoRestriction.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 4).
		Return(restriction, nil)

	response := svc.FindById(context.Background(), 4)

	assert.Equal(t, 4, response.Id)
	assert.Equal(t, "North Zone", response.Name)
	assert.Len(t, response.Buildings, 1)

	repoRestriction.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestRestrictionFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoRestriction := &mocks.MockRepositoryBuildingRestriction{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newRestrictionService(db, repoRestriction, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoRestriction.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.BuildingRestriction{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "building restriction not found"},
		func() { svc.FindById(context.Background(), 404) },
	)

	repoRestriction.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Update ---

func TestRestrictionUpdate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoRestriction := &mocks.MockRepositoryBuildingRestriction{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newRestrictionService(db, repoRestriction, repoBuilding)

	existing := newRestrictionModel(6, "OldZone")
	bldg := testutil.NewBuilding(15, "New Building")
	updated := newRestrictionModel(6, "NewZone", models.BuildingRef{Id: 15, Name: "New Building"})

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoRestriction.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 6).
		Return(existing, nil)
	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 15).
		Return(bldg, nil)
	repoRestriction.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(r models.BuildingRestriction) bool { return r.Name == "NewZone" }),
		[]int{15},
	).Return(updated, nil)

	request := webRestriction.UpdateBuildingRestrictionRequest{
		Name:        "NewZone",
		BuildingIds: []int{15},
	}

	response := svc.Update(context.Background(), request, 6)

	assert.Equal(t, 6, response.Id)
	assert.Equal(t, "NewZone", response.Name)
	assert.Equal(t, "New Building", response.Buildings[0].Name)

	repoRestriction.AssertExpectations(t)
	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestRestrictionUpdate_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoRestriction := &mocks.MockRepositoryBuildingRestriction{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newRestrictionService(db, repoRestriction, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoRestriction.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.BuildingRestriction{}, sql.ErrNoRows)

	request := webRestriction.UpdateBuildingRestrictionRequest{Name: "X", BuildingIds: []int{1}}

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "building restriction not found"},
		func() { svc.Update(context.Background(), request, 404) },
	)

	repoRestriction.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Delete ---

func TestRestrictionDelete_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoRestriction := &mocks.MockRepositoryBuildingRestriction{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newRestrictionService(db, repoRestriction, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoRestriction.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(newRestrictionModel(2, "ToDelete"), nil)
	repoRestriction.On("Delete", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(nil)

	assert.NotPanics(t, func() { svc.Delete(context.Background(), 2) })

	repoRestriction.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestRestrictionDelete_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoRestriction := &mocks.MockRepositoryBuildingRestriction{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newRestrictionService(db, repoRestriction, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoRestriction.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.BuildingRestriction{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "building restriction not found"},
		func() { svc.Delete(context.Background(), 999) },
	)

	repoRestriction.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
