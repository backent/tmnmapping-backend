package motherbrand_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	serviceMotherBrand "github.com/malikabdulaziz/tmn-backend/services/motherbrand"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webMotherBrand "github.com/malikabdulaziz/tmn-backend/web/motherbrand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newMotherBrandService(db *sql.DB, repo *mocks.MockRepositoryMotherBrand) serviceMotherBrand.ServiceMotherBrandInterface {
	return serviceMotherBrand.NewServiceMotherBrandImpl(db, repo)
}

func newMotherBrandModel(id int, name string) models.MotherBrand {
	return models.MotherBrand{Id: id, Name: name}
}

// --- Create ---

func TestMotherBrandCreate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryMotherBrand{}
	svc := newMotherBrandService(db, repo)

	created := newMotherBrandModel(1, "Nike")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.MotherBrand) bool { return c.Name == "Nike" }),
	).Return(created, nil)

	request := webMotherBrand.CreateMotherBrandRequest{Name: "Nike"}
	response := svc.Create(context.Background(), request)

	assert.Equal(t, 1, response.Id)
	assert.Equal(t, "Nike", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindById ---

func TestMotherBrandFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryMotherBrand{}
	svc := newMotherBrandService(db, repo)

	motherBrand := newMotherBrandModel(3, "Adidas")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).
		Return(motherBrand, nil)

	response := svc.FindById(context.Background(), 3)

	assert.Equal(t, 3, response.Id)
	assert.Equal(t, "Adidas", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestMotherBrandFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryMotherBrand{}
	svc := newMotherBrandService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.MotherBrand{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "mother brand not found"},
		func() { svc.FindById(context.Background(), 404) },
	)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Update ---

func TestMotherBrandUpdate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryMotherBrand{}
	svc := newMotherBrandService(db, repo)

	existing := newMotherBrandModel(5, "Old Brand")
	updated := newMotherBrandModel(5, "New Brand")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.MotherBrand) bool { return c.Name == "New Brand" && c.Id == 5 }),
	).Return(updated, nil)

	request := webMotherBrand.UpdateMotherBrandRequest{Name: "New Brand"}
	response := svc.Update(context.Background(), request, 5)

	assert.Equal(t, 5, response.Id)
	assert.Equal(t, "New Brand", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Delete ---

func TestMotherBrandDelete_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryMotherBrand{}
	svc := newMotherBrandService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(newMotherBrandModel(2, "ToDelete"), nil)
	repo.On("Delete", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(nil)

	assert.NotPanics(t, func() { svc.Delete(context.Background(), 2) })

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestMotherBrandDelete_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryMotherBrand{}
	svc := newMotherBrandService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.MotherBrand{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "mother brand not found"},
		func() { svc.Delete(context.Background(), 999) },
	)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Import ---

func TestMotherBrandImport_CSV_NewItem(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryMotherBrand{}
	svc := newMotherBrandService(db, repo)

	csvData := "Name\nNike\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindByName", mock.Anything, mock.AnythingOfType("*sql.Tx"), "Nike").
		Return(models.MotherBrand{}, sql.ErrNoRows)

	created := newMotherBrandModel(1, "Nike")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.MotherBrand) bool { return c.Name == "Nike" }),
	).Return(created, nil)

	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, "Nike", responses[0].Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestMotherBrandImport_CSV_ExistingItem(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryMotherBrand{}
	svc := newMotherBrandService(db, repo)

	csvData := "Name\nNike\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	existing := newMotherBrandModel(7, "Nike")
	repo.On("FindByName", mock.Anything, mock.AnythingOfType("*sql.Tx"), "Nike").
		Return(existing, nil)

	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, 7, responses[0].Id)
	assert.Equal(t, "Nike", responses[0].Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
