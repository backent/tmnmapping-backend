package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositorySavedPolygon implements repositories/savedpolygon.RepositorySavedPolygonInterface
type MockRepositorySavedPolygon struct {
	mock.Mock
}

func (m *MockRepositorySavedPolygon) Create(ctx context.Context, tx *sql.Tx, polygon models.SavedPolygon, points []models.SavedPolygonPoint) (models.SavedPolygon, error) {
	args := m.Called(ctx, tx, polygon, points)
	return args.Get(0).(models.SavedPolygon), args.Error(1)
}

func (m *MockRepositorySavedPolygon) CreatePoint(ctx context.Context, tx *sql.Tx, point models.SavedPolygonPoint) (models.SavedPolygonPoint, error) {
	args := m.Called(ctx, tx, point)
	return args.Get(0).(models.SavedPolygonPoint), args.Error(1)
}

func (m *MockRepositorySavedPolygon) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.SavedPolygon, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection)
	return args.Get(0).([]models.SavedPolygon), args.Error(1)
}

func (m *MockRepositorySavedPolygon) CountAll(ctx context.Context, tx *sql.Tx) (int, error) {
	args := m.Called(ctx, tx)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositorySavedPolygon) FindById(ctx context.Context, tx *sql.Tx, id int) (models.SavedPolygon, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.SavedPolygon), args.Error(1)
}

func (m *MockRepositorySavedPolygon) Update(ctx context.Context, tx *sql.Tx, polygon models.SavedPolygon, points []models.SavedPolygonPoint) (models.SavedPolygon, error) {
	args := m.Called(ctx, tx, polygon, points)
	return args.Get(0).(models.SavedPolygon), args.Error(1)
}

func (m *MockRepositorySavedPolygon) DeletePointsBySavedPolygonId(ctx context.Context, tx *sql.Tx, savedPolygonId int) error {
	args := m.Called(ctx, tx, savedPolygonId)
	return args.Error(0)
}

func (m *MockRepositorySavedPolygon) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}
