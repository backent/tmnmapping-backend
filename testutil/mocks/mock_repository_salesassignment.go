package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositorySalesAssignment implements
// repositories/salesassignment.RepositorySalesAssignmentInterface
type MockRepositorySalesAssignment struct {
	mock.Mock
}

func (m *MockRepositorySalesAssignment) Create(ctx context.Context, tx *sql.Tx, a models.SalesAssignment) (models.SalesAssignment, error) {
	args := m.Called(ctx, tx, a)
	return args.Get(0).(models.SalesAssignment), args.Error(1)
}

func (m *MockRepositorySalesAssignment) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.SalesAssignment, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search)
	return args.Get(0).([]models.SalesAssignment), args.Error(1)
}

func (m *MockRepositorySalesAssignment) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositorySalesAssignment) FindById(ctx context.Context, tx *sql.Tx, id int) (models.SalesAssignment, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.SalesAssignment), args.Error(1)
}

func (m *MockRepositorySalesAssignment) FindByCustomerAndBrand(ctx context.Context, tx *sql.Tx, customerId int, brandId int) (models.SalesAssignment, error) {
	args := m.Called(ctx, tx, customerId, brandId)
	return args.Get(0).(models.SalesAssignment), args.Error(1)
}

func (m *MockRepositorySalesAssignment) Update(ctx context.Context, tx *sql.Tx, a models.SalesAssignment) (models.SalesAssignment, error) {
	args := m.Called(ctx, tx, a)
	return args.Get(0).(models.SalesAssignment), args.Error(1)
}

func (m *MockRepositorySalesAssignment) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositorySalesAssignment) CountBySalesUser(ctx context.Context, tx *sql.Tx, salesUserId int) (int, error) {
	args := m.Called(ctx, tx, salesUserId)
	return args.Int(0), args.Error(1)
}
