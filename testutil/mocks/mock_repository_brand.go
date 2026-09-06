package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryBrand implements repositories/brand.RepositoryBrandInterface.
// This is the advertiser brand, not MotherBrand (the POI/competitor concept).
type MockRepositoryBrand struct {
	mock.Mock
}

func (m *MockRepositoryBrand) Create(ctx context.Context, tx *sql.Tx, b models.Brand) (models.Brand, error) {
	args := m.Called(ctx, tx, b)
	return args.Get(0).(models.Brand), args.Error(1)
}

func (m *MockRepositoryBrand) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string, customerId int) ([]models.Brand, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search, customerId)
	return args.Get(0).([]models.Brand), args.Error(1)
}

func (m *MockRepositoryBrand) CountAll(ctx context.Context, tx *sql.Tx, search string, customerId int) (int, error) {
	args := m.Called(ctx, tx, search, customerId)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryBrand) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Brand, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.Brand), args.Error(1)
}

func (m *MockRepositoryBrand) FindByCode(ctx context.Context, tx *sql.Tx, code string) (models.Brand, error) {
	args := m.Called(ctx, tx, code)
	return args.Get(0).(models.Brand), args.Error(1)
}

func (m *MockRepositoryBrand) Update(ctx context.Context, tx *sql.Tx, b models.Brand) (models.Brand, error) {
	args := m.Called(ctx, tx, b)
	return args.Get(0).(models.Brand), args.Error(1)
}

func (m *MockRepositoryBrand) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}
