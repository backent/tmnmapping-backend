package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositorySubCategory implements repositories/subcategory.RepositorySubCategoryInterface
type MockRepositorySubCategory struct {
	mock.Mock
}

func (m *MockRepositorySubCategory) Create(ctx context.Context, tx *sql.Tx, subCategory models.SubCategory) (models.SubCategory, error) {
	args := m.Called(ctx, tx, subCategory)
	return args.Get(0).(models.SubCategory), args.Error(1)
}

func (m *MockRepositorySubCategory) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.SubCategory, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search)
	return args.Get(0).([]models.SubCategory), args.Error(1)
}

func (m *MockRepositorySubCategory) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositorySubCategory) FindById(ctx context.Context, tx *sql.Tx, id int) (models.SubCategory, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.SubCategory), args.Error(1)
}

func (m *MockRepositorySubCategory) Update(ctx context.Context, tx *sql.Tx, subCategory models.SubCategory) (models.SubCategory, error) {
	args := m.Called(ctx, tx, subCategory)
	return args.Get(0).(models.SubCategory), args.Error(1)
}

func (m *MockRepositorySubCategory) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositorySubCategory) FindByName(ctx context.Context, tx *sql.Tx, name string) (models.SubCategory, error) {
	args := m.Called(ctx, tx, name)
	return args.Get(0).(models.SubCategory), args.Error(1)
}

func (m *MockRepositorySubCategory) FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.SubCategory, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]models.SubCategory), args.Error(1)
}
