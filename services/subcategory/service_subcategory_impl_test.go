package subcategory_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	serviceSubCategory "github.com/malikabdulaziz/tmn-backend/services/subcategory"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webSubCategory "github.com/malikabdulaziz/tmn-backend/web/subcategory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newSubCategoryService(db *sql.DB, repo *mocks.MockRepositorySubCategory) serviceSubCategory.ServiceSubCategoryInterface {
	return serviceSubCategory.NewServiceSubCategoryImpl(db, repo)
}

func newSubCategoryModel(id int, name string) models.SubCategory {
	return models.SubCategory{Id: id, Name: name}
}

// --- Create ---

func TestSubCategoryCreate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositorySubCategory{}
	svc := newSubCategoryService(db, repo)

	created := newSubCategoryModel(1, "Fast Food")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.SubCategory) bool { return c.Name == "Fast Food" }),
	).Return(created, nil)

	request := webSubCategory.CreateSubCategoryRequest{Name: "Fast Food"}
	response := svc.Create(context.Background(), request)

	assert.Equal(t, 1, response.Id)
	assert.Equal(t, "Fast Food", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindById ---

func TestSubCategoryFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositorySubCategory{}
	svc := newSubCategoryService(db, repo)

	subCategory := newSubCategoryModel(3, "Coffee Shop")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).
		Return(subCategory, nil)

	response := svc.FindById(context.Background(), 3)

	assert.Equal(t, 3, response.Id)
	assert.Equal(t, "Coffee Shop", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSubCategoryFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositorySubCategory{}
	svc := newSubCategoryService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.SubCategory{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "sub category not found"},
		func() { svc.FindById(context.Background(), 404) },
	)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Update ---

func TestSubCategoryUpdate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositorySubCategory{}
	svc := newSubCategoryService(db, repo)

	existing := newSubCategoryModel(5, "Old Name")
	updated := newSubCategoryModel(5, "New Name")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.SubCategory) bool { return c.Name == "New Name" && c.Id == 5 }),
	).Return(updated, nil)

	request := webSubCategory.UpdateSubCategoryRequest{Name: "New Name"}
	response := svc.Update(context.Background(), request, 5)

	assert.Equal(t, 5, response.Id)
	assert.Equal(t, "New Name", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Delete ---

func TestSubCategoryDelete_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositorySubCategory{}
	svc := newSubCategoryService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(newSubCategoryModel(2, "ToDelete"), nil)
	repo.On("Delete", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(nil)

	assert.NotPanics(t, func() { svc.Delete(context.Background(), 2) })

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSubCategoryDelete_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositorySubCategory{}
	svc := newSubCategoryService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.SubCategory{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "sub category not found"},
		func() { svc.Delete(context.Background(), 999) },
	)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Import ---

func TestSubCategoryImport_CSV_NewItem(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositorySubCategory{}
	svc := newSubCategoryService(db, repo)

	csvData := "Name\nFast Food\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindByName", mock.Anything, mock.AnythingOfType("*sql.Tx"), "Fast Food").
		Return(models.SubCategory{}, sql.ErrNoRows)

	created := newSubCategoryModel(1, "Fast Food")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.SubCategory) bool { return c.Name == "Fast Food" }),
	).Return(created, nil)

	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, "Fast Food", responses[0].Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSubCategoryImport_CSV_ExistingItem(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositorySubCategory{}
	svc := newSubCategoryService(db, repo)

	csvData := "Name\nFast Food\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	existing := newSubCategoryModel(7, "Fast Food")
	repo.On("FindByName", mock.Anything, mock.AnythingOfType("*sql.Tx"), "Fast Food").
		Return(existing, nil)

	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, 7, responses[0].Id)
	assert.Equal(t, "Fast Food", responses[0].Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
