package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryPOIPoint implements repositories/poipoint.RepositoryPOIPointInterface
type MockRepositoryPOIPoint struct {
	mock.Mock
}

func (m *MockRepositoryPOIPoint) Create(ctx context.Context, tx *sql.Tx, point models.POIPoint) (models.POIPoint, error) {
	args := m.Called(ctx, tx, point)
	return args.Get(0).(models.POIPoint), args.Error(1)
}

func (m *MockRepositoryPOIPoint) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.POIPoint, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search)
	return args.Get(0).([]models.POIPoint), args.Error(1)
}

func (m *MockRepositoryPOIPoint) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryPOIPoint) FindById(ctx context.Context, tx *sql.Tx, id int) (models.POIPoint, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.POIPoint), args.Error(1)
}

func (m *MockRepositoryPOIPoint) Update(ctx context.Context, tx *sql.Tx, point models.POIPoint) (models.POIPoint, error) {
	args := m.Called(ctx, tx, point)
	return args.Get(0).(models.POIPoint), args.Error(1)
}

func (m *MockRepositoryPOIPoint) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositoryPOIPoint) FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.POIPoint, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]models.POIPoint), args.Error(1)
}

func (m *MockRepositoryPOIPoint) FindAllFlat(ctx context.Context, tx *sql.Tx, search string) ([]models.POIPoint, error) {
	args := m.Called(ctx, tx, search)
	return args.Get(0).([]models.POIPoint), args.Error(1)
}

func (m *MockRepositoryPOIPoint) FindPOIRefsByPointId(ctx context.Context, tx *sql.Tx, pointId int) ([]models.POIRef, error) {
	args := m.Called(ctx, tx, pointId)
	return args.Get(0).([]models.POIRef), args.Error(1)
}

func (m *MockRepositoryPOIPoint) FindPOIRefsByPointIds(ctx context.Context, tx *sql.Tx, pointIds []int) (map[int][]models.POIRef, error) {
	args := m.Called(ctx, tx, pointIds)
	return args.Get(0).(map[int][]models.POIRef), args.Error(1)
}

func (m *MockRepositoryPOIPoint) FindByNameAndAddress(ctx context.Context, tx *sql.Tx, poiName string, address string) (models.POIPoint, error) {
	args := m.Called(ctx, tx, poiName, address)
	return args.Get(0).(models.POIPoint), args.Error(1)
}

func (m *MockRepositoryPOIPoint) FindByPOINames(ctx context.Context, tx *sql.Tx, poiNames []string) ([]models.POIPoint, error) {
	args := m.Called(ctx, tx, poiNames)
	return args.Get(0).([]models.POIPoint), args.Error(1)
}
