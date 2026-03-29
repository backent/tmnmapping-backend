package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryMotherBrand implements repositories/motherbrand.RepositoryMotherBrandInterface
type MockRepositoryMotherBrand struct {
	mock.Mock
}

func (m *MockRepositoryMotherBrand) Create(ctx context.Context, tx *sql.Tx, motherBrand models.MotherBrand) (models.MotherBrand, error) {
	args := m.Called(ctx, tx, motherBrand)
	return args.Get(0).(models.MotherBrand), args.Error(1)
}

func (m *MockRepositoryMotherBrand) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.MotherBrand, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search)
	return args.Get(0).([]models.MotherBrand), args.Error(1)
}

func (m *MockRepositoryMotherBrand) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryMotherBrand) FindById(ctx context.Context, tx *sql.Tx, id int) (models.MotherBrand, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.MotherBrand), args.Error(1)
}

func (m *MockRepositoryMotherBrand) Update(ctx context.Context, tx *sql.Tx, motherBrand models.MotherBrand) (models.MotherBrand, error) {
	args := m.Called(ctx, tx, motherBrand)
	return args.Get(0).(models.MotherBrand), args.Error(1)
}

func (m *MockRepositoryMotherBrand) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositoryMotherBrand) FindByName(ctx context.Context, tx *sql.Tx, name string) (models.MotherBrand, error) {
	args := m.Called(ctx, tx, name)
	return args.Get(0).(models.MotherBrand), args.Error(1)
}

func (m *MockRepositoryMotherBrand) FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.MotherBrand, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]models.MotherBrand), args.Error(1)
}
