package mocks

import (
	"context"
	"database/sql"

	"github.com/stretchr/testify/mock"

	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuildingProject "github.com/malikabdulaziz/tmn-backend/repositories/buildingproject"
)

type MockRepositoryBuildingProject struct {
	mock.Mock
}

func (m *MockRepositoryBuildingProject) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, filter repositoriesBuildingProject.ListFilter, includeFinance bool) ([]models.BuildingProject, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, filter, includeFinance)
	return args.Get(0).([]models.BuildingProject), args.Error(1)
}

func (m *MockRepositoryBuildingProject) CountAll(ctx context.Context, tx *sql.Tx, filter repositoriesBuildingProject.ListFilter) (int, error) {
	args := m.Called(ctx, tx, filter)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryBuildingProject) FindById(ctx context.Context, tx *sql.Tx, id int, includeFinance bool) (models.BuildingProject, error) {
	args := m.Called(ctx, tx, id, includeFinance)
	return args.Get(0).(models.BuildingProject), args.Error(1)
}

func (m *MockRepositoryBuildingProject) FindByIris(ctx context.Context, tx *sql.Tx, iris string, includeFinance bool) (models.BuildingProject, error) {
	args := m.Called(ctx, tx, iris, includeFinance)
	return args.Get(0).(models.BuildingProject), args.Error(1)
}

func (m *MockRepositoryBuildingProject) Create(ctx context.Context, tx *sql.Tx, project models.BuildingProject) (models.BuildingProject, error) {
	args := m.Called(ctx, tx, project)
	return args.Get(0).(models.BuildingProject), args.Error(1)
}

func (m *MockRepositoryBuildingProject) Update(ctx context.Context, tx *sql.Tx, project models.BuildingProject) (models.BuildingProject, error) {
	args := m.Called(ctx, tx, project)
	return args.Get(0).(models.BuildingProject), args.Error(1)
}

func (m *MockRepositoryBuildingProject) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositoryBuildingProject) RecordChanges(ctx context.Context, tx *sql.Tx, changes []models.BuildingProjectChange) error {
	args := m.Called(ctx, tx, changes)
	return args.Error(0)
}

func (m *MockRepositoryBuildingProject) FindChanges(ctx context.Context, tx *sql.Tx, projectId int, take int, skip int) ([]models.BuildingProjectChange, error) {
	args := m.Called(ctx, tx, projectId, take, skip)
	return args.Get(0).([]models.BuildingProjectChange), args.Error(1)
}

func (m *MockRepositoryBuildingProject) CountChanges(ctx context.Context, tx *sql.Tx, projectId int) (int, error) {
	args := m.Called(ctx, tx, projectId)
	return args.Int(0), args.Error(1)
}
