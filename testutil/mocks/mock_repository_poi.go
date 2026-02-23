package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryPOI implements repositories/poi.RepositoryPOIInterface
type MockRepositoryPOI struct {
	mock.Mock
}

func (m *MockRepositoryPOI) Create(ctx context.Context, tx *sql.Tx, poi models.POI) (models.POI, error) {
	args := m.Called(ctx, tx, poi)
	return args.Get(0).(models.POI), args.Error(1)
}

func (m *MockRepositoryPOI) CreatePoint(ctx context.Context, tx *sql.Tx, point models.POIPoint) (models.POIPoint, error) {
	args := m.Called(ctx, tx, point)
	return args.Get(0).(models.POIPoint), args.Error(1)
}

func (m *MockRepositoryPOI) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.POI, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection)
	return args.Get(0).([]models.POI), args.Error(1)
}

func (m *MockRepositoryPOI) CountAll(ctx context.Context, tx *sql.Tx) (int, error) {
	args := m.Called(ctx, tx)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryPOI) FindById(ctx context.Context, tx *sql.Tx, id int) (models.POI, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.POI), args.Error(1)
}

func (m *MockRepositoryPOI) Update(ctx context.Context, tx *sql.Tx, poi models.POI) (models.POI, error) {
	args := m.Called(ctx, tx, poi)
	return args.Get(0).(models.POI), args.Error(1)
}

func (m *MockRepositoryPOI) DeletePointsByPOIId(ctx context.Context, tx *sql.Tx, poiId int) error {
	args := m.Called(ctx, tx, poiId)
	return args.Error(0)
}

func (m *MockRepositoryPOI) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}
