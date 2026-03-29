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

func (m *MockRepositoryPOI) Create(ctx context.Context, tx *sql.Tx, poi models.POI, pointIds []int) (models.POI, error) {
	args := m.Called(ctx, tx, poi, pointIds)
	return args.Get(0).(models.POI), args.Error(1)
}

func (m *MockRepositoryPOI) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.POI, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search)
	return args.Get(0).([]models.POI), args.Error(1)
}

func (m *MockRepositoryPOI) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryPOI) FindAllFlat(ctx context.Context, tx *sql.Tx, search string) ([]models.POI, error) {
	args := m.Called(ctx, tx, search)
	return args.Get(0).([]models.POI), args.Error(1)
}

func (m *MockRepositoryPOI) FindById(ctx context.Context, tx *sql.Tx, id int) (models.POI, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.POI), args.Error(1)
}

func (m *MockRepositoryPOI) Update(ctx context.Context, tx *sql.Tx, poi models.POI, pointIds []int) (models.POI, error) {
	args := m.Called(ctx, tx, poi, pointIds)
	return args.Get(0).(models.POI), args.Error(1)
}

func (m *MockRepositoryPOI) CreatePointLink(ctx context.Context, tx *sql.Tx, poiId int, pointId int) error {
	args := m.Called(ctx, tx, poiId, pointId)
	return args.Error(0)
}

func (m *MockRepositoryPOI) DeletePointLinksByPOIId(ctx context.Context, tx *sql.Tx, poiId int) error {
	args := m.Called(ctx, tx, poiId)
	return args.Error(0)
}

func (m *MockRepositoryPOI) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositoryPOI) FindByBrands(ctx context.Context, tx *sql.Tx, brands []string) ([]models.POI, error) {
	args := m.Called(ctx, tx, brands)
	return args.Get(0).([]models.POI), args.Error(1)
}
