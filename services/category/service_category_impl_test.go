package category_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	serviceCategory "github.com/malikabdulaziz/tmn-backend/services/category"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webCategory "github.com/malikabdulaziz/tmn-backend/web/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newCategoryService(db *sql.DB, repo *mocks.MockRepositoryCategory) serviceCategory.ServiceCategoryInterface {
	return serviceCategory.NewServiceCategoryImpl(db, repo)
}

func newCategoryModel(id int, name string) models.Category {
	return models.Category{Id: id, Name: name}
}

// --- Create ---

func TestCategoryCreate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryCategory{}
	svc := newCategoryService(db, repo)

	created := newCategoryModel(1, "Retail")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.Category) bool { return c.Name == "Retail" }),
	).Return(created, nil)

	request := webCategory.CreateCategoryRequest{Name: "Retail"}
	response := svc.Create(context.Background(), request)

	assert.Equal(t, 1, response.Id)
	assert.Equal(t, "Retail", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindById ---

func TestCategoryFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryCategory{}
	svc := newCategoryService(db, repo)

	category := newCategoryModel(3, "Food & Beverage")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).
		Return(category, nil)

	response := svc.FindById(context.Background(), 3)

	assert.Equal(t, 3, response.Id)
	assert.Equal(t, "Food & Beverage", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCategoryFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryCategory{}
	svc := newCategoryService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.Category{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "category not found"},
		func() { svc.FindById(context.Background(), 404) },
	)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Update ---

func TestCategoryUpdate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryCategory{}
	svc := newCategoryService(db, repo)

	existing := newCategoryModel(5, "Old Name")
	updated := newCategoryModel(5, "New Name")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.Category) bool { return c.Name == "New Name" && c.Id == 5 }),
	).Return(updated, nil)

	request := webCategory.UpdateCategoryRequest{Name: "New Name"}
	response := svc.Update(context.Background(), request, 5)

	assert.Equal(t, 5, response.Id)
	assert.Equal(t, "New Name", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Delete ---

func TestCategoryDelete_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryCategory{}
	svc := newCategoryService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(newCategoryModel(2, "ToDelete"), nil)
	repo.On("Delete", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(nil)

	assert.NotPanics(t, func() { svc.Delete(context.Background(), 2) })

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCategoryDelete_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryCategory{}
	svc := newCategoryService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.Category{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "category not found"},
		func() { svc.Delete(context.Background(), 999) },
	)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Import ---

func TestCategoryImport_CSV_NewItem(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryCategory{}
	svc := newCategoryService(db, repo)

	csvData := "Name\nRetail\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindByName", mock.Anything, mock.AnythingOfType("*sql.Tx"), "Retail").
		Return(models.Category{}, sql.ErrNoRows)

	created := newCategoryModel(1, "Retail")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.Category) bool { return c.Name == "Retail" }),
	).Return(created, nil)

	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, "Retail", responses[0].Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCategoryImport_CSV_ExistingItem(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryCategory{}
	svc := newCategoryService(db, repo)

	csvData := "Name\nRetail\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	existing := newCategoryModel(7, "Retail")
	repo.On("FindByName", mock.Anything, mock.AnythingOfType("*sql.Tx"), "Retail").
		Return(existing, nil)

	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, 7, responses[0].Id)
	assert.Equal(t, "Retail", responses[0].Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
