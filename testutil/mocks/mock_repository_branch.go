package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryBranch implements repositories/branch.RepositoryBranchInterface
type MockRepositoryBranch struct {
	mock.Mock
}

func (m *MockRepositoryBranch) Create(ctx context.Context, tx *sql.Tx, branch models.Branch) (models.Branch, error) {
	args := m.Called(ctx, tx, branch)
	return args.Get(0).(models.Branch), args.Error(1)
}

func (m *MockRepositoryBranch) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.Branch, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search)
	return args.Get(0).([]models.Branch), args.Error(1)
}

func (m *MockRepositoryBranch) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryBranch) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Branch, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.Branch), args.Error(1)
}

func (m *MockRepositoryBranch) Update(ctx context.Context, tx *sql.Tx, branch models.Branch) (models.Branch, error) {
	args := m.Called(ctx, tx, branch)
	return args.Get(0).(models.Branch), args.Error(1)
}

func (m *MockRepositoryBranch) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositoryBranch) FindByName(ctx context.Context, tx *sql.Tx, name string) (models.Branch, error) {
	args := m.Called(ctx, tx, name)
	return args.Get(0).(models.Branch), args.Error(1)
}

func (m *MockRepositoryBranch) FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.Branch, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]models.Branch), args.Error(1)
}
