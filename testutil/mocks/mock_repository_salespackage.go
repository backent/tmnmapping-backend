package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositorySalesPackage implements repositories/salespackage.RepositorySalesPackageInterface
type MockRepositorySalesPackage struct {
	mock.Mock
}

func (m *MockRepositorySalesPackage) Create(ctx context.Context, tx *sql.Tx, pkg models.SalesPackage, buildingIds []int) (models.SalesPackage, error) {
	args := m.Called(ctx, tx, pkg, buildingIds)
	return args.Get(0).(models.SalesPackage), args.Error(1)
}

func (m *MockRepositorySalesPackage) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.SalesPackage, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection)
	return args.Get(0).([]models.SalesPackage), args.Error(1)
}

func (m *MockRepositorySalesPackage) CountAll(ctx context.Context, tx *sql.Tx) (int, error) {
	args := m.Called(ctx, tx)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositorySalesPackage) FindById(ctx context.Context, tx *sql.Tx, id int) (models.SalesPackage, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.SalesPackage), args.Error(1)
}

func (m *MockRepositorySalesPackage) Update(ctx context.Context, tx *sql.Tx, pkg models.SalesPackage, buildingIds []int) (models.SalesPackage, error) {
	args := m.Called(ctx, tx, pkg, buildingIds)
	return args.Get(0).(models.SalesPackage), args.Error(1)
}

func (m *MockRepositorySalesPackage) DeleteBuildingLinksBySalesPackageId(ctx context.Context, tx *sql.Tx, salesPackageId int) error {
	args := m.Called(ctx, tx, salesPackageId)
	return args.Error(0)
}

func (m *MockRepositorySalesPackage) CreateBuildingLink(ctx context.Context, tx *sql.Tx, salesPackageId int, buildingId int) error {
	args := m.Called(ctx, tx, salesPackageId, buildingId)
	return args.Error(0)
}

func (m *MockRepositorySalesPackage) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}
