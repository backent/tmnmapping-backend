package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryCategory implements repositories/category.RepositoryCategoryInterface
type MockRepositoryCategory struct {
	mock.Mock
}

func (m *MockRepositoryCategory) Create(ctx context.Context, tx *sql.Tx, category models.Category) (models.Category, error) {
	args := m.Called(ctx, tx, category)
	return args.Get(0).(models.Category), args.Error(1)
}

func (m *MockRepositoryCategory) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.Category, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search)
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockRepositoryCategory) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryCategory) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Category, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.Category), args.Error(1)
}

func (m *MockRepositoryCategory) Update(ctx context.Context, tx *sql.Tx, category models.Category) (models.Category, error) {
	args := m.Called(ctx, tx, category)
	return args.Get(0).(models.Category), args.Error(1)
}

func (m *MockRepositoryCategory) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositoryCategory) FindByName(ctx context.Context, tx *sql.Tx, name string) (models.Category, error) {
	args := m.Called(ctx, tx, name)
	return args.Get(0).(models.Category), args.Error(1)
}

func (m *MockRepositoryCategory) FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.Category, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]models.Category), args.Error(1)
}
