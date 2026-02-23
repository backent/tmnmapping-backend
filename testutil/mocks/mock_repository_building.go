package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryBuilding implements repositories/building.RepositoryBuildingInterface
type MockRepositoryBuilding struct {
	mock.Mock
}

func (m *MockRepositoryBuilding) Create(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error) {
	args := m.Called(ctx, tx, building)
	return args.Get(0).(models.Building), args.Error(1)
}

func (m *MockRepositoryBuilding) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Building, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.Building), args.Error(1)
}

func (m *MockRepositoryBuilding) FindByExternalId(ctx context.Context, tx *sql.Tx, externalId string) (models.Building, error) {
	args := m.Called(ctx, tx, externalId)
	return args.Get(0).(models.Building), args.Error(1)
}

func (m *MockRepositoryBuilding) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string, buildingStatus string, sellable string, connectivity string, resourceType string, competitorLocation *bool, cbdArea string, subdistrict string, citytown string, province string, gradeResource string, buildingType string) ([]models.Building, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search, buildingStatus, sellable, connectivity, resourceType, competitorLocation, cbdArea, subdistrict, citytown, province, gradeResource, buildingType)
	return args.Get(0).([]models.Building), args.Error(1)
}

func (m *MockRepositoryBuilding) CountAll(ctx context.Context, tx *sql.Tx, search string, buildingStatus string, sellable string, connectivity string, resourceType string, competitorLocation *bool, cbdArea string, subdistrict string, citytown string, province string, gradeResource string, buildingType string) (int, error) {
	args := m.Called(ctx, tx, search, buildingStatus, sellable, connectivity, resourceType, competitorLocation, cbdArea, subdistrict, citytown, province, gradeResource, buildingType)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryBuilding) GetDistinctValues(ctx context.Context, tx *sql.Tx, columnName string) ([]string, error) {
	args := m.Called(ctx, tx, columnName)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRepositoryBuilding) Update(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error) {
	args := m.Called(ctx, tx, building)
	return args.Get(0).(models.Building), args.Error(1)
}

func (m *MockRepositoryBuilding) UpdateFromSync(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error) {
	args := m.Called(ctx, tx, building)
	return args.Get(0).(models.Building), args.Error(1)
}

func (m *MockRepositoryBuilding) FindAllForMapping(ctx context.Context, tx *sql.Tx, buildingType string, buildingGrade string, year string, subdistrict string, progress string, sellable string, connectivity string, lcdPresence string, salesPackageIds string, buildingRestrictionIds string, lat *float64, lng *float64, radius *int, poiPoints []struct{ Lat float64; Lng float64 }, polygonPoints []struct{ Lat float64; Lng float64 }, minLat *float64, maxLat *float64, minLng *float64, maxLng *float64) ([]models.Building, error) {
	args := m.Called(ctx, tx, buildingType, buildingGrade, year, subdistrict, progress, sellable, connectivity, lcdPresence, salesPackageIds, buildingRestrictionIds, lat, lng, radius, poiPoints, polygonPoints, minLat, maxLat, minLng, maxLng)
	return args.Get(0).([]models.Building), args.Error(1)
}

func (m *MockRepositoryBuilding) FindByIds(ctx context.Context, tx *sql.Tx, ids []int) ([]models.Building, error) {
	args := m.Called(ctx, tx, ids)
	return args.Get(0).([]models.Building), args.Error(1)
}

func (m *MockRepositoryBuilding) GetLCDPresenceSummary(ctx context.Context, tx *sql.Tx) ([]repositoriesBuilding.LCDPresenceCountRow, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]repositoriesBuilding.LCDPresenceCountRow), args.Error(1)
}

func (m *MockRepositoryBuilding) FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.Building, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]models.Building), args.Error(1)
}
