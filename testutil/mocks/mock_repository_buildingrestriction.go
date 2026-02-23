package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryBuildingRestriction implements repositories/buildingrestriction.RepositoryBuildingRestrictionInterface
type MockRepositoryBuildingRestriction struct {
	mock.Mock
}

func (m *MockRepositoryBuildingRestriction) Create(ctx context.Context, tx *sql.Tx, restriction models.BuildingRestriction, buildingIds []int) (models.BuildingRestriction, error) {
	args := m.Called(ctx, tx, restriction, buildingIds)
	return args.Get(0).(models.BuildingRestriction), args.Error(1)
}

func (m *MockRepositoryBuildingRestriction) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.BuildingRestriction, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection)
	return args.Get(0).([]models.BuildingRestriction), args.Error(1)
}

func (m *MockRepositoryBuildingRestriction) CountAll(ctx context.Context, tx *sql.Tx) (int, error) {
	args := m.Called(ctx, tx)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryBuildingRestriction) FindById(ctx context.Context, tx *sql.Tx, id int) (models.BuildingRestriction, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.BuildingRestriction), args.Error(1)
}

func (m *MockRepositoryBuildingRestriction) Update(ctx context.Context, tx *sql.Tx, restriction models.BuildingRestriction, buildingIds []int) (models.BuildingRestriction, error) {
	args := m.Called(ctx, tx, restriction, buildingIds)
	return args.Get(0).(models.BuildingRestriction), args.Error(1)
}

func (m *MockRepositoryBuildingRestriction) DeleteBuildingLinksByBuildingRestrictionId(ctx context.Context, tx *sql.Tx, buildingRestrictionId int) error {
	args := m.Called(ctx, tx, buildingRestrictionId)
	return args.Error(0)
}

func (m *MockRepositoryBuildingRestriction) CreateBuildingLink(ctx context.Context, tx *sql.Tx, buildingRestrictionId int, buildingId int) error {
	args := m.Called(ctx, tx, buildingRestrictionId, buildingId)
	return args.Error(0)
}

func (m *MockRepositoryBuildingRestriction) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}
