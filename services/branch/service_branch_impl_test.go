package branch_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	serviceBranch "github.com/malikabdulaziz/tmn-backend/services/branch"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webBranch "github.com/malikabdulaziz/tmn-backend/web/branch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newBranchService(db *sql.DB, repo *mocks.MockRepositoryBranch) serviceBranch.ServiceBranchInterface {
	return serviceBranch.NewServiceBranchImpl(db, repo)
}

func newBranchModel(id int, name string) models.Branch {
	return models.Branch{Id: id, Name: name}
}

// --- Create ---

func TestBranchCreate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryBranch{}
	svc := newBranchService(db, repo)

	created := newBranchModel(1, "Jakarta Branch")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.Branch) bool { return c.Name == "Jakarta Branch" }),
	).Return(created, nil)

	request := webBranch.CreateBranchRequest{Name: "Jakarta Branch"}
	response := svc.Create(context.Background(), request)

	assert.Equal(t, 1, response.Id)
	assert.Equal(t, "Jakarta Branch", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- FindById ---

func TestBranchFindById_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryBranch{}
	svc := newBranchService(db, repo)

	branch := newBranchModel(3, "Surabaya Branch")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).
		Return(branch, nil)

	response := svc.FindById(context.Background(), 3)

	assert.Equal(t, 3, response.Id)
	assert.Equal(t, "Surabaya Branch", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestBranchFindById_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryBranch{}
	svc := newBranchService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.Branch{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "branch not found"},
		func() { svc.FindById(context.Background(), 404) },
	)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Update ---

func TestBranchUpdate_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryBranch{}
	svc := newBranchService(db, repo)

	existing := newBranchModel(5, "Old Branch")
	updated := newBranchModel(5, "New Branch")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).
		Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.Branch) bool { return c.Name == "New Branch" && c.Id == 5 }),
	).Return(updated, nil)

	request := webBranch.UpdateBranchRequest{Name: "New Branch"}
	response := svc.Update(context.Background(), request, 5)

	assert.Equal(t, 5, response.Id)
	assert.Equal(t, "New Branch", response.Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Delete ---

func TestBranchDelete_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryBranch{}
	svc := newBranchService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(newBranchModel(2, "ToDelete"), nil)
	repo.On("Delete", mock.Anything, mock.AnythingOfType("*sql.Tx"), 2).
		Return(nil)

	assert.NotPanics(t, func() { svc.Delete(context.Background(), 2) })

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestBranchDelete_NotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryBranch{}
	svc := newBranchService(db, repo)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repo.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 999).
		Return(models.Branch{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NotFoundError{Error: "branch not found"},
		func() { svc.Delete(context.Background(), 999) },
	)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- Import ---

func TestBranchImport_CSV_NewItem(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryBranch{}
	svc := newBranchService(db, repo)

	csvData := "Name\nJakarta Branch\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("FindByName", mock.Anything, mock.AnythingOfType("*sql.Tx"), "Jakarta Branch").
		Return(models.Branch{}, sql.ErrNoRows)

	created := newBranchModel(1, "Jakarta Branch")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(c models.Branch) bool { return c.Name == "Jakarta Branch" }),
	).Return(created, nil)

	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, "Jakarta Branch", responses[0].Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestBranchImport_CSV_ExistingItem(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryBranch{}
	svc := newBranchService(db, repo)

	csvData := "Name\nJakarta Branch\n"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	existing := newBranchModel(7, "Jakarta Branch")
	repo.On("FindByName", mock.Anything, mock.AnythingOfType("*sql.Tx"), "Jakarta Branch").
		Return(existing, nil)

	responses := svc.Import(context.Background(), []byte(csvData), "csv")

	assert.Len(t, responses, 1)
	assert.Equal(t, 7, responses[0].Id)
	assert.Equal(t, "Jakarta Branch", responses[0].Name)

	repo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
