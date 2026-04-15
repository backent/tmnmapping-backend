package salespackage_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	serviceSalesPackage "github.com/malikabdulaziz/tmn-backend/services/salespackage"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webSalesPackage "github.com/malikabdulaziz/tmn-backend/web/salespackage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newSalesPackageService(
	db *sql.DB,
	repoPkg *mocks.MockRepositorySalesPackage,
	repoBuilding *mocks.MockRepositoryBuilding,
) serviceSalesPackage.ServiceSalesPackageInterface {
	return serviceSalesPackage.NewServiceSalesPackageImpl(db, repoPkg, repoBuilding)
}

func newSalesPackageModel(id int, name string, buildingRefs ...models.BuildingRef) models.SalesPackage {
	return models.SalesPackage{Id: id, Name: name, Buildings: buildingRefs}
}

// --- Create ---

func TestSalesPackageCreate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	bldg := testutil.NewBuilding(10, "Tower A")
	createdRef := models.BuildingRef{
		Id:           10,
		Name:         "Tower A",
		ProjectName:  "Alpha Project",
		Subdistrict:  "Menteng",
		Citytown:     "Jakarta Pusat",
		Province:     "DKI Jakarta",
		BuildingType: "Office",
	}
	created := newSalesPackageModel(1, "Package Alpha", createdRef)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	// validateBuildingIdsErr calls FindById for each building ID
	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 10).
		Return(bldg, nil)

	repoPkg.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(p models.SalesPackage) bool { return p.Name == "Package Alpha" }),
		[]int{10},
	).Return(created, nil)

	request := webSalesPackage.CreateSalesPackageRequest{
		Name:        "Package Alpha",
		BuildingIds: []int{10},
	}

	response := svc.Create(context.Background(), request)

	assert.Equal(t, 1, response.Id)
	assert.Equal(t, "Package Alpha", response.Name)
	assert.Len(t, response.Buildings, 1)
	assert.Equal(t, 10, response.Buildings[0].Id)
	assert.Equal(t, "Tower A", response.Buildings[0].Name)
	assert.Equal(t, "Alpha Project", response.Buildings[0].ProjectName)
	assert.Equal(t, "Menteng", response.Buildings[0].Subdistrict)
	assert.Equal(t, "Jakarta Pusat", response.Buildings[0].Citytown)
	assert.Equal(t, "DKI Jakarta", response.Buildings[0].Province)
	assert.Equal(t, "Office", response.Buildings[0].BuildingType)

	repoPkg.AssertExpectations(t)
	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestSalesPackageCreate_InvalidBuilding verifies that a non-existent building ID
// causes a BadRequestError panic before the package is created.
func TestSalesPackageCreate_InvalidBuilding(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.Building{}, sql.ErrNoRows)

	request := webSalesPackage.CreateSalesPackageRequest{
		Name:        "Bad Package",
		BuildingIds: []int{999},
	}

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "building not found"},
		func() { svc.Create(context.Background(), request) },
	)

	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindById ---

func TestSalesPackageFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	pkg := newSalesPackageModel(7, "Premium Package", models.BuildingRef{Id: 1, Name: "HQ"})

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPkg.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 7).
		Return(pkg, nil)

	response := svc.FindById(context.Background(), 7)

	assert.Equal(t, 7, response.Id)
	assert.Equal(t, "Premium Package", response.Name)
	assert.Len(t, response.Buildings, 1)

	repoPkg.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSalesPackageFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPkg.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.SalesPackage{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "sales package not found"},
		func() { svc.FindById(context.Background(), 404) },
	)

	repoPkg.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Update ---

func TestSalesPackageUpdate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	existing := newSalesPackageModel(5, "OldName")
	bldg := testutil.NewBuilding(20, "Tower B")
	updated := newSalesPackageModel(5, "NewName", models.BuildingRef{Id: 20, Name: "Tower B"})

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPkg.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return(existing, nil)
	repoBuilding.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 20).
		Return(bldg, nil)
	repoPkg.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(p models.SalesPackage) bool { return p.Name == "NewName" }),
		[]int{20},
	).Return(updated, nil)

	request := webSalesPackage.UpdateSalesPackageRequest{
		Name:        "NewName",
		BuildingIds: []int{20},
	}

	response := svc.Update(context.Background(), request, 5)

	assert.Equal(t, 5, response.Id)
	assert.Equal(t, "NewName", response.Name)
	assert.Equal(t, "Tower B", response.Buildings[0].Name)

	repoPkg.AssertExpectations(t)
	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSalesPackageUpdate_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPkg.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.SalesPackage{}, sql.ErrNoRows)

	request := webSalesPackage.UpdateSalesPackageRequest{Name: "X", BuildingIds: []int{1}}

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "sales package not found"},
		func() { svc.Update(context.Background(), request, 404) },
	)

	repoPkg.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Delete ---

func TestSalesPackageDelete_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPkg.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).
		Return(newSalesPackageModel(3, "ToDelete"), nil)
	repoPkg.On("Delete", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).
		Return(nil)

	assert.NotPanics(t, func() { svc.Delete(context.Background(), 3) })

	repoPkg.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSalesPackageDelete_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoPkg.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.SalesPackage{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "sales package not found"},
		func() { svc.Delete(context.Background(), 999) },
	)

	repoPkg.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Export ---

func TestSalesPackageExport_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	packages := []models.SalesPackage{
		{Id: 1, Name: "Package A", Buildings: []models.BuildingRef{{Id: 10, Name: "Tower A"}}},
		{Id: 2, Name: "Package B", Buildings: []models.BuildingRef{{Id: 20, Name: "Tower B"}, {Id: 30, Name: "Tower C"}}},
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoPkg.On("FindAllFlat", mock.Anything, mock.AnythingOfType("*sql.Tx"), "").
		Return(packages, nil)

	excelBytes, err := svc.Export(context.Background(), "")

	assert.NoError(t, err)
	assert.NotEmpty(t, excelBytes)
	// Verify it starts with XLSX magic bytes (PK zip header)
	assert.Equal(t, byte(0x50), excelBytes[0])
	assert.Equal(t, byte(0x4B), excelBytes[1])

	repoPkg.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Import ---

func TestSalesPackageImport_UnsupportedFileType(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "Unsupported file type. Use xlsx or csv."},
		func() { svc.Import(context.Background(), []byte("data"), "txt") },
	)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSalesPackageImport_CSV_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoPkg := &mocks.MockRepositorySalesPackage{}
	repoBuilding := &mocks.MockRepositoryBuilding{}
	svc := newSalesPackageService(db, repoPkg, repoBuilding)

	csvData := "Name,Building Name\nPackage X,Tower A\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoBuilding.On("FindAllDropdown", mock.Anything, mock.AnythingOfType("*sql.Tx")).
		Return([]models.Building{testutil.NewBuilding(10, "Tower A")}, nil)

	repoPkg.On("FindByNames", mock.Anything, mock.AnythingOfType("*sql.Tx"), []string{"Package X"}).
		Return([]models.SalesPackage{}, nil)

	created := models.SalesPackage{
		Id:        1,
		Name:      "Package X",
		Buildings: []models.BuildingRef{{Id: 10, Name: "Tower A"}},
	}
	repoPkg.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(p models.SalesPackage) bool { return p.Name == "Package X" }),
		[]int{10},
	).Return(created, nil)

	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, "Package X", responses[0].Name)
	assert.Len(t, responses[0].Buildings, 1)

	repoPkg.AssertExpectations(t)
	repoBuilding.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
