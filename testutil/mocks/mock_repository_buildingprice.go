package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryBuildingPrice implements repositories/buildingprice.RepositoryBuildingPriceInterface
type MockRepositoryBuildingPrice struct {
	mock.Mock
}

func (m *MockRepositoryBuildingPrice) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, search string) ([]models.BuildingPrice, error) {
	args := m.Called(ctx, tx, take, skip, search)
	return args.Get(0).([]models.BuildingPrice), args.Error(1)
}

func (m *MockRepositoryBuildingPrice) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryBuildingPrice) FindByBuildingId(ctx context.Context, tx *sql.Tx, buildingId int) (models.BuildingPrice, error) {
	args := m.Called(ctx, tx, buildingId)
	return args.Get(0).(models.BuildingPrice), args.Error(1)
}

func (m *MockRepositoryBuildingPrice) FindPricesByBuildingIds(ctx context.Context, tx *sql.Tx, ids []int) (map[int]int64, error) {
	args := m.Called(ctx, tx, ids)
	return args.Get(0).(map[int]int64), args.Error(1)
}

func (m *MockRepositoryBuildingPrice) FindAllPrices(ctx context.Context, tx *sql.Tx) (map[int]int64, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(map[int]int64), args.Error(1)
}

func (m *MockRepositoryBuildingPrice) Upsert(ctx context.Context, tx *sql.Tx, buildingId int, price int64) error {
	args := m.Called(ctx, tx, buildingId, price)
	return args.Error(0)
}

func (m *MockRepositoryBuildingPrice) Delete(ctx context.Context, tx *sql.Tx, buildingId int) error {
	args := m.Called(ctx, tx, buildingId)
	return args.Error(0)
}
