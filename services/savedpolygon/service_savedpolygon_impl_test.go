package savedpolygon_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	servicePolygon "github.com/malikabdulaziz/tmn-backend/services/savedpolygon"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webPolygon "github.com/malikabdulaziz/tmn-backend/web/savedpolygon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newPolygonService(db *sql.DB, repoPolygon *mocks.MockRepositorySavedPolygon) servicePolygon.ServiceSavedPolygonInterface {
	return servicePolygon.NewServiceSavedPolygonImpl(db, repoPolygon)
}

// threePoints returns a minimal valid 3-point polygon request.
func threePoints() []webPolygon.SavedPolygonPointRequest {
	return []webPolygon.SavedPolygonPointRequest{
		{Lat: -6.1, Lng: 106.7},
		{Lat: -6.2, Lng: 106.8},
		{Lat: -6.3, Lng: 106.9},
	}
}

func newPolygonModel(id int, name string) models.SavedPolygon {
	return models.SavedPolygon{
		Id:   id,
		Name: name,
		Points: []models.SavedPolygonPoint{
			{Ord: 0, Lat: -6.1, Lng: 106.7},
			{Ord: 1, Lat: -6.2, Lng: 106.8},
			{Ord: 2, Lat: -6.3, Lng: 106.9},
		},
	}
}

// --- Create ---

func TestPolygonCreate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	created := newPolygonModel(1, "My Area")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPolygon.On("Create",
		mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(p models.SavedPolygon) bool { return p.Name == "My Area" }),
		mock.Anything,
	).Return(created, nil)

	request := webPolygon.CreateSavedPolygonRequest{
		Name:   "My Area",
		Points: threePoints(),
	}

	response := svc.Create(context.Background(), request)

	assert.Equal(t, 1, response.Id)
	assert.Equal(t, "My Area", response.Name)
	assert.Len(t, response.Points, 3)

	repoPolygon.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestPolygonCreate_TooFewPoints verifies that fewer than 3 points panics before
// the transaction even touches the repository.
func TestPolygonCreate_TooFewPoints(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	request := webPolygon.CreateSavedPolygonRequest{
		Name: "Small",
		Points: []webPolygon.SavedPolygonPointRequest{
			{Lat: -6.1, Lng: 106.7},
			{Lat: -6.2, Lng: 106.8},
		},
	}

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "polygon must have at least 3 points"},
		func() { svc.Create(context.Background(), request) },
	)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestPolygonCreate_InvalidLat verifies that an out-of-range latitude panics.
func TestPolygonCreate_InvalidLat(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	request := webPolygon.CreateSavedPolygonRequest{
		Name: "BadLat",
		Points: []webPolygon.SavedPolygonPointRequest{
			{Lat: -91.0, Lng: 106.7}, // invalid: below -90
			{Lat: -6.2, Lng: 106.8},
			{Lat: -6.3, Lng: 106.9},
		},
	}

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "invalid lat at point 1"},
		func() { svc.Create(context.Background(), request) },
	)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestPolygonCreate_InvalidLng verifies that an out-of-range longitude panics.
func TestPolygonCreate_InvalidLng(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	request := webPolygon.CreateSavedPolygonRequest{
		Name: "BadLng",
		Points: []webPolygon.SavedPolygonPointRequest{
			{Lat: -6.1, Lng: 181.0}, // invalid: above 180
			{Lat: -6.2, Lng: 106.8},
			{Lat: -6.3, Lng: 106.9},
		},
	}

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "invalid lng at point 1"},
		func() { svc.Create(context.Background(), request) },
	)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindById ---

func TestPolygonFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	polygon := newPolygonModel(8, "Jakarta Area")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPolygon.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 8).
		Return(polygon, nil)

	response := svc.FindById(context.Background(), 8)

	assert.Equal(t, 8, response.Id)
	assert.Equal(t, "Jakarta Area", response.Name)
	assert.Len(t, response.Points, 3)

	repoPolygon.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPolygonFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPolygon.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.SavedPolygon{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "saved polygon not found"},
		func() { svc.FindById(context.Background(), 999) },
	)

	repoPolygon.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Update ---

func TestPolygonUpdate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	existing := newPolygonModel(4, "OldArea")
	updated := newPolygonModel(4, "NewArea")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPolygon.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 4).
		Return(existing, nil)
	repoPolygon.On("Update",
		mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(p models.SavedPolygon) bool { return p.Name == "NewArea" }),
		mock.Anything,
	).Return(updated, nil)

	request := webPolygon.UpdateSavedPolygonRequest{
		Name:   "NewArea",
		Points: threePoints(),
	}

	response := svc.Update(context.Background(), request, 4)

	assert.Equal(t, 4, response.Id)
	assert.Equal(t, "NewArea", response.Name)

	repoPolygon.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPolygonUpdate_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	request := webPolygon.UpdateSavedPolygonRequest{
		Name:   "X",
		Points: threePoints(),
	}

	// validatePoints runs before FindById so we need FindById to return not found
	// but validatePoints passes here (3 valid points), so the panic should be NotFoundError.
	repoPolygon.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.SavedPolygon{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "saved polygon not found"},
		func() { svc.Update(context.Background(), request, 404) },
	)

	repoPolygon.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestPolygonUpdate_InvalidPoints verifies that point validation runs before
// the repository is consulted.
func TestPolygonUpdate_TooFewPoints(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	request := webPolygon.UpdateSavedPolygonRequest{
		Name: "X",
		Points: []webPolygon.SavedPolygonPointRequest{
			{Lat: -6.1, Lng: 106.7},
			{Lat: -6.2, Lng: 106.8},
		},
	}

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "polygon must have at least 3 points"},
		func() { svc.Update(context.Background(), request, 1) },
	)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Delete ---

func TestPolygonDelete_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPolygon.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 9).
		Return(newPolygonModel(9, "ToDelete"), nil)
	repoPolygon.On("Delete", mock.Anything, mock.AnythingOfType("*sql.Tx"), 9).
		Return(nil)

	assert.NotPanics(t, func() { svc.Delete(context.Background(), 9) })

	repoPolygon.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPolygonDelete_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPolygon := &mocks.MockRepositorySavedPolygon{}
	svc := newPolygonService(db, repoPolygon)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPolygon.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.SavedPolygon{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "saved polygon not found"},
		func() { svc.Delete(context.Background(), 999) },
	)

	repoPolygon.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
