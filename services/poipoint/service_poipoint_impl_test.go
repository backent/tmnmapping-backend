package poipoint_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	servicePOIPoint "github.com/malikabdulaziz/tmn-backend/services/poipoint"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webPOIPoint "github.com/malikabdulaziz/tmn-backend/web/poipoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newPOIPointService(
	db *sql.DB,
	repoPOIPoint *mocks.MockRepositoryPOIPoint,
	repoCategory *mocks.MockRepositoryCategory,
	repoSubCategory *mocks.MockRepositorySubCategory,
	repoMotherBrand *mocks.MockRepositoryMotherBrand,
	repoBranch *mocks.MockRepositoryBranch,
) servicePOIPoint.ServicePOIPointInterface {
	return servicePOIPoint.NewServicePOIPointImpl(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)
}

func newPOIPointModel(id int, poiName string, poiRefs ...models.POIRef) models.POIPoint {
	if poiRefs == nil {
		poiRefs = []models.POIRef{}
	}
	return models.POIPoint{Id: id, POIName: poiName, POIs: poiRefs}
}

func newMetaMocks() (*mocks.MockRepositoryCategory, *mocks.MockRepositorySubCategory, *mocks.MockRepositoryMotherBrand, *mocks.MockRepositoryBranch) {
	return &mocks.MockRepositoryCategory{}, &mocks.MockRepositorySubCategory{}, &mocks.MockRepositoryMotherBrand{}, &mocks.MockRepositoryBranch{}
}

// --- Create ---

func TestPOIPointCreate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	created := newPOIPointModel(1, "Store A")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOIPoint.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(p models.POIPoint) bool { return p.POIName == "Store A" }),
	).Return(created, nil)

	request := webPOIPoint.CreatePOIPointRequest{
		POIName:   "Store A",
		Address:   "123 Main St",
		Latitude:  -6.2,
		Longitude: 106.8,
	}

	response := svc.Create(context.Background(), request)

	assert.Equal(t, 1, response.Id)
	assert.Equal(t, "Store A", response.POIName)

	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindById ---

func TestPOIPointFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	point := newPOIPointModel(4, "Office B", models.POIRef{Id: 1, Brand: "BrandX"})

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOIPoint.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 4).
		Return(point, nil)

	response := svc.FindById(context.Background(), 4)

	assert.Equal(t, 4, response.Id)
	assert.Equal(t, "Office B", response.POIName)
	assert.Len(t, response.POIs, 1)
	assert.Equal(t, "BrandX", response.POIs[0].Brand)

	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPOIPointFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPOIPoint.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.POIPoint{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "POI point not found"},
		func() { svc.FindById(context.Background(), 404) },
	)

	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Update ---

func TestPOIPointUpdate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	existing := newPOIPointModel(6, "OldName")
	updated := newPOIPointModel(6, "NewName")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOIPoint.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 6).
		Return(existing, nil)
	repoPOIPoint.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(p models.POIPoint) bool { return p.POIName == "NewName" }),
	).Return(updated, nil)

	request := webPOIPoint.UpdatePOIPointRequest{
		POIName:   "NewName",
		Address:   "456 Oak Ave",
		Latitude:  -6.3,
		Longitude: 106.9,
	}

	response := svc.Update(context.Background(), request, 6)

	assert.Equal(t, 6, response.Id)
	assert.Equal(t, "NewName", response.POIName)

	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Delete ---

func TestPOIPointDelete_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOIPoint.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(newPOIPointModel(2, "ToDelete"), nil)
	repoPOIPoint.On("Delete", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(nil)

	assert.NotPanics(t, func() { svc.Delete(context.Background(), 2) })

	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPOIPointDelete_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPOIPoint.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.POIPoint{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "POI point not found"},
		func() { svc.Delete(context.Background(), 999) },
	)

	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- GetPointUsage ---

func TestPOIPointGetPointUsage_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOIPoint.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return(newPOIPointModel(5, "Some Point"), nil)
	repoPOIPoint.On("FindPOIRefsByPointId", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return([]models.POIRef{{Id: 1, Brand: "BrandA"}, {Id: 2, Brand: "BrandB"}}, nil)

	response := svc.GetPointUsage(context.Background(), 5)

	assert.Len(t, response.POIs, 2)
	assert.Equal(t, "BrandA", response.POIs[0].Brand)
	assert.Equal(t, "BrandB", response.POIs[1].Brand)

	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Export ---

func TestPOIPointExport_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	points := []models.POIPoint{
		{Id: 1, POIName: "Store A", Address: "123 Main", Latitude: -6.2, Longitude: 106.8, POIs: []models.POIRef{{Id: 1, Brand: "BrandX"}}},
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPOIPoint.On("FindAllFlat", mock.Anything, mock.AnythingOfType("*sql.Tx"), "").
		Return(points, nil)

	excelBytes, err := svc.Export(context.Background(), "")

	assert.NoError(t, err)
	assert.NotEmpty(t, excelBytes)
	assert.Equal(t, byte(0x50), excelBytes[0]) // PK zip header
	assert.Equal(t, byte(0x4B), excelBytes[1])

	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Import ---

func TestPOIPointImport_UnsupportedFileType(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "Unsupported file type. Use xlsx or csv."},
		func() { svc.Import(context.Background(), []byte("data"), "txt") },
	)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPOIPointImport_CSV_HappyPath_NewPoint(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	csvData := "POI Name,Address,Coordinate\nStore A,123 Main,\"-6.2, 106.8\"\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	// Point does not exist yet
	repoPOIPoint.On("FindByNameAndAddress", mock.Anything, mock.AnythingOfType("*sql.Tx"), "Store A", "123 Main").
		Return(models.POIPoint{}, sql.ErrNoRows)

	created := models.POIPoint{Id: 1, POIName: "Store A", Address: "123 Main", POIs: []models.POIRef{}}
	repoPOIPoint.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(p models.POIPoint) bool { return p.POIName == "Store A" }),
	).Return(created, nil)

	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, "Store A", responses[0].POIName)

	repoPOIPoint.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPOIPointImport_CSV_ExistingPoint(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPOIPoint := &mocks.MockRepositoryPOIPoint{}
	repoCategory, repoSubCategory, repoMotherBrand, repoBranch := newMetaMocks()
	svc := newPOIPointService(db, repoPOIPoint, repoCategory, repoSubCategory, repoMotherBrand, repoBranch)

	csvData := "POI Name,Address,Coordinate\nStore A,123 Main,\"-6.2, 106.8\"\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	// Point already exists
	existing := models.POIPoint{Id: 42, POIName: "Store A", Address: "123 Main", POIs: []models.POIRef{}}
	repoPOIPoint.On("FindByNameAndAddress", mock.Anything, mock.AnythingOfType("*sql.Tx"), "Store A", "123 Main").
		Return(existing, nil)

	// Create should NOT be called
	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, 42, responses[0].Id)
	assert.Equal(t, "Store A", responses[0].POIName)

	repoPOIPoint.AssertExpectations(t)
	repoPOIPoint.AssertNotCalled(t, "Create")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
