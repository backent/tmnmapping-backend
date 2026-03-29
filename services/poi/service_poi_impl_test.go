package poi_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	servicePOI "github.com/malikabdulaziz/tmn-backend/services/poi"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webPOI "github.com/malikabdulaziz/tmn-backend/web/poi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newPOIService(db *sql.DB, repoPOI *mocks.MockRepositoryPOI, repoPOIPoint *mocks.MockRepositoryPOIPoint) servicePOI.ServicePOIInterface {
	return servicePOI.NewServicePOIImpl(db, repoPOI, repoPOIPoint)
}

func newPOIModel(id int, brand, color string) models.POI {
	return models.POI{
		Id:    id,
		Brand: brand,
		Color: color,
		Points: []models.POIPoint{
			{Id: 1, POIName: "Place A", Address: "Addr A", Latitude: -6.2, Longitude: 106.8, POIs: []models.POIRef{}},
		},
	}
}

// --- Create ---

func TestPOICreate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	svc := newPOIService(db, repoPOI, repoPOIPoint)

	created := newPOIModel(1, "Landmark", "#FF0000")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	// Validate point ID exists
	repoPOIPoint.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 1).
		Return(models.POIPoint{Id: 1, POIName: "Place A", POIs: []models.POIRef{}}, nil)

	repoPOI.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(p models.POI) bool { return p.Brand == "Landmark" && p.Color == "#FF0000" }),
		[]int{1},
	).Return(created, nil)

	request := webPOI.CreatePOIRequest{
		Brand:    "Landmark",
		Color:    "#FF0000",
		PointIds: []int{1},
	}

	response := svc.Create(context.Background(), request)

	assert.Equal(t, 1, response.Id)
	assert.Equal(t, "Landmark", response.Brand)
	assert.Equal(t, "#FF0000", response.Color)
	assert.Len(t, response.Points, 1)
	assert.Equal(t, "Place A", response.Points[0].POIName)

	repoPOI.AssertExpectations(t)
	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPOICreate_InvalidPointId(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	svc := newPOIService(db, repoPOI, repoPOIPoint)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPOIPoint.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.POIPoint{}, sql.ErrNoRows)

	request := webPOI.CreatePOIRequest{
		Brand:    "Test",
		Color:    "#000",
		PointIds: []int{999},
	}

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "POI point not found"},
		func() { svc.Create(context.Background(), request) },
	)

	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindById ---

func TestPOIFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	svc := newPOIService(db, repoPOI, repoPOIPoint)

	poi := newPOIModel(5, "Central Park POI", "#00FF00")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOI.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return(poi, nil)

	response := svc.FindById(context.Background(), 5)

	assert.Equal(t, 5, response.Id)
	assert.Equal(t, "Central Park POI", response.Brand)
	assert.Equal(t, "#00FF00", response.Color)

	repoPOI.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPOIFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	svc := newPOIService(db, repoPOI, repoPOIPoint)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPOI.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.POI{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "POI not found"},
		func() { svc.FindById(context.Background(), 999) },
	)

	repoPOI.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Update ---

func TestPOIUpdate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	svc := newPOIService(db, repoPOI, repoPOIPoint)

	existing := newPOIModel(3, "OldBrand", "#000000")
	updated := newPOIModel(3, "NewBrand", "#FFFFFF")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOI.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).
		Return(existing, nil)
	repoPOIPoint.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 10).
		Return(models.POIPoint{Id: 10, POIName: "New Place", POIs: []models.POIRef{}}, nil)
	repoPOI.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(p models.POI) bool { return p.Id == 3 && p.Brand == "NewBrand" && p.Color == "#FFFFFF" }),
		[]int{10},
	).Return(updated, nil)

	request := webPOI.UpdatePOIRequest{
		Brand:    "NewBrand",
		Color:    "#FFFFFF",
		PointIds: []int{10},
	}

	response := svc.Update(context.Background(), request, 3)

	assert.Equal(t, 3, response.Id)
	assert.Equal(t, "NewBrand", response.Brand)
	assert.Equal(t, "#FFFFFF", response.Color)

	repoPOI.AssertExpectations(t)
	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPOIUpdate_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	svc := newPOIService(db, repoPOI, repoPOIPoint)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPOI.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.POI{}, sql.ErrNoRows)

	request := webPOI.UpdatePOIRequest{
		Brand: "X", Color: "#FFF", PointIds: []int{1},
	}

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "POI not found"},
		func() { svc.Update(context.Background(), request, 404) },
	)

	repoPOI.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Delete ---

func TestPOIDelete_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	svc := newPOIService(db, repoPOI, repoPOIPoint)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOI.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(newPOIModel(2, "ToDelete", "#111"), nil)
	repoPOI.On("Delete", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(nil)

	assert.NotPanics(t, func() { svc.Delete(context.Background(), 2) })

	repoPOI.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPOIDelete_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	svc := newPOIService(db, repoPOI, repoPOIPoint)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPOI.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.POI{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "POI not found"},
		func() { svc.Delete(context.Background(), 999) },
	)

	repoPOI.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindAll ---

func TestPOIFindAll_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	svc := newPOIService(db, repoPOI, repoPOIPoint)

	poiList := []models.POI{
		newPOIModel(1, "POI A", "#AAA"),
		newPOIModel(2, "POI B", "#BBB"),
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOI.On("FindAll", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(poiList, nil)

	repoPOI.On("CountAll", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.Anything).
		Return(2, nil)

	var req webPOI.POIRequestFindAll
	req.SetTake(10)
	req.SetSkip(0)

	responses, total := svc.FindAll(context.Background(), req)

	assert.Equal(t, 2, total)
	assert.Len(t, responses, 2)
	assert.Equal(t, "POI A", responses[0].Brand)
	assert.Equal(t, "POI B", responses[1].Brand)

	repoPOI.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
