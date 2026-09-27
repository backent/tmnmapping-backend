package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryCustomer implements repositories/customer.RepositoryCustomerInterface
type MockRepositoryCustomer struct {
	mock.Mock
}

func (m *MockRepositoryCustomer) Create(ctx context.Context, tx *sql.Tx, c models.Customer) (models.Customer, error) {
	args := m.Called(ctx, tx, c)
	return args.Get(0).(models.Customer), args.Error(1)
}

func (m *MockRepositoryCustomer) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.Customer, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search)
	return args.Get(0).([]models.Customer), args.Error(1)
}

func (m *MockRepositoryCustomer) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryCustomer) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Customer, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.Customer), args.Error(1)
}

func (m *MockRepositoryCustomer) FindByCode(ctx context.Context, tx *sql.Tx, code string) (models.Customer, error) {
	args := m.Called(ctx, tx, code)
	return args.Get(0).(models.Customer), args.Error(1)
}

func (m *MockRepositoryCustomer) Update(ctx context.Context, tx *sql.Tx, c models.Customer) (models.Customer, error) {
	args := m.Called(ctx, tx, c)
	return args.Get(0).(models.Customer), args.Error(1)
}

func (m *MockRepositoryCustomer) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositoryCustomer) CountBrands(ctx context.Context, tx *sql.Tx, customerId int) (int, error) {
	args := m.Called(ctx, tx, customerId)
	return args.Int(0), args.Error(1)
}
