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

func newPOIService(db *sql.DB, repoPOI *mocks.MockRepositoryPOI) servicePOI.ServicePOIInterface {
	return servicePOI.NewServicePOIImpl(db, repoPOI)
}

func newPOIModel(id int, name, color string) models.POI {
	return models.POI{
		Id:    id,
		Name:  name,
		Color: color,
		Points: []models.POIPoint{
			{Id: 1, PlaceName: "Place A", Address: "Addr A", Latitude: -6.2, Longitude: 106.8},
		},
	}
}

// --- Create ---

func TestPOICreate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	svc := newPOIService(db, repoPOI)

	created := newPOIModel(1, "Landmark", "#FF0000")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOI.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.MatchedBy(func(p models.POI) bool {
		return p.Name == "Landmark" && p.Color == "#FF0000" && len(p.Points) == 1
	})).Return(created, nil)

	request := webPOI.CreatePOIRequest{
		Name:  "Landmark",
		Color: "#FF0000",
		Points: []webPOI.POIPointRequest{
			{PlaceName: "Place A", Address: "Addr A", Latitude: -6.2, Longitude: 106.8},
		},
	}

	response := svc.Create(context.Background(), request)

	assert.Equal(t, 1, response.Id)
	assert.Equal(t, "Landmark", response.Name)
	assert.Equal(t, "#FF0000", response.Color)
	assert.Len(t, response.Points, 1)
	assert.Equal(t, "Place A", response.Points[0].PlaceName)

	repoPOI.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindById ---

func TestPOIFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	svc := newPOIService(db, repoPOI)

	poi := newPOIModel(5, "Central Park POI", "#00FF00")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOI.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return(poi, nil)

	response := svc.FindById(context.Background(), 5)

	assert.Equal(t, 5, response.Id)
	assert.Equal(t, "Central Park POI", response.Name)
	assert.Equal(t, "#00FF00", response.Color)

	repoPOI.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPOIFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	svc := newPOIService(db, repoPOI)

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
	svc := newPOIService(db, repoPOI)

	existing := newPOIModel(3, "OldName", "#000000")
	updated := newPOIModel(3, "NewName", "#FFFFFF")
	updated.Points = []models.POIPoint{
		{PlaceName: "New Place", Address: "New Addr", Latitude: -7.0, Longitude: 107.0},
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOI.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).
		Return(existing, nil)
	repoPOI.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.MatchedBy(func(p models.POI) bool {
		return p.Id == 3 && p.Name == "NewName" && p.Color == "#FFFFFF"
	})).Return(updated, nil)

	request := webPOI.UpdatePOIRequest{
		Name:  "NewName",
		Color: "#FFFFFF",
		Points: []webPOI.POIPointRequest{
			{PlaceName: "New Place", Address: "New Addr", Latitude: -7.0, Longitude: 107.0},
		},
	}

	response := svc.Update(context.Background(), request, 3)

	assert.Equal(t, 3, response.Id)
	assert.Equal(t, "NewName", response.Name)
	assert.Equal(t, "#FFFFFF", response.Color)
	assert.Equal(t, "New Place", response.Points[0].PlaceName)

	repoPOI.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPOIUpdate_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOI := &mocks.MockRepositoryPOI{}
	svc := newPOIService(db, repoPOI)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPOI.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.POI{}, sql.ErrNoRows)

	request := webPOI.UpdatePOIRequest{
		Name: "X", Color: "#FFF",
		Points: []webPOI.POIPointRequest{{PlaceName: "P", Address: "A", Latitude: 1, Longitude: 1}},
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
	svc := newPOIService(db, repoPOI)

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
	svc := newPOIService(db, repoPOI)

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
	svc := newPOIService(db, repoPOI)

	poiList := []models.POI{
		newPOIModel(1, "POI A", "#AAA"),
		newPOIModel(2, "POI B", "#BBB"),
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOI.On("FindAll", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(poiList, nil)

	repoPOI.On("CountAll", mock.Anything, mock.AnythingOfType("*sql.Tx")).
		Return(2, nil)

	var req webPOI.POIRequestFindAll
	req.SetTake(10)
	req.SetSkip(0)

	responses, total := svc.FindAll(context.Background(), req)

	assert.Equal(t, 2, total)
	assert.Len(t, responses, 2)
	assert.Equal(t, "POI A", responses[0].Name)
	assert.Equal(t, "POI B", responses[1].Name)

	repoPOI.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
