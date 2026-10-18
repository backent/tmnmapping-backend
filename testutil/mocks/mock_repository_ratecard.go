package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryRateCard implements repositories/ratecard.RepositoryRateCardInterface
type MockRepositoryRateCard struct {
	mock.Mock
}

func (m *MockRepositoryRateCard) CreateVersion(ctx context.Context, tx *sql.Tx, v models.RateCardVersion) (models.RateCardVersion, error) {
	args := m.Called(ctx, tx, v)
	return args.Get(0).(models.RateCardVersion), args.Error(1)
}

func (m *MockRepositoryRateCard) FindAllVersions(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.RateCardVersion, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search)
	return args.Get(0).([]models.RateCardVersion), args.Error(1)
}

func (m *MockRepositoryRateCard) CountAllVersions(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryRateCard) FindVersionById(ctx context.Context, tx *sql.Tx, id int) (models.RateCardVersion, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.RateCardVersion), args.Error(1)
}

func (m *MockRepositoryRateCard) FindVersionByCode(ctx context.Context, tx *sql.Tx, code string) (models.RateCardVersion, error) {
	args := m.Called(ctx, tx, code)
	return args.Get(0).(models.RateCardVersion), args.Error(1)
}

func (m *MockRepositoryRateCard) FindCurrentVersion(ctx context.Context, tx *sql.Tx) (models.RateCardVersion, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(models.RateCardVersion), args.Error(1)
}

func (m *MockRepositoryRateCard) UpdateVersion(ctx context.Context, tx *sql.Tx, v models.RateCardVersion) (models.RateCardVersion, error) {
	args := m.Called(ctx, tx, v)
	return args.Get(0).(models.RateCardVersion), args.Error(1)
}

func (m *MockRepositoryRateCard) DeleteVersion(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositoryRateCard) SetVersionStatus(ctx context.Context, tx *sql.Tx, id int, status string, publishedByUserId int) error {
	args := m.Called(ctx, tx, id, status, publishedByUserId)
	return args.Error(0)
}

func (m *MockRepositoryRateCard) DemoteCurrentVersion(ctx context.Context, tx *sql.Tx) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockRepositoryRateCard) FindBuildingPrices(ctx context.Context, tx *sql.Tx, versionId int, take int, skip int, search string) ([]models.RateCardBuildingPrice, error) {
	args := m.Called(ctx, tx, versionId, take, skip, search)
	return args.Get(0).([]models.RateCardBuildingPrice), args.Error(1)
}

func (m *MockRepositoryRateCard) CountBuildingPrices(ctx context.Context, tx *sql.Tx, versionId int, search string) (int, error) {
	args := m.Called(ctx, tx, versionId, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryRateCard) UpsertBuildingPrice(ctx context.Context, tx *sql.Tx, p models.RateCardBuildingPrice) error {
	args := m.Called(ctx, tx, p)
	return args.Error(0)
}

func (m *MockRepositoryRateCard) DeleteBuildingPrice(ctx context.Context, tx *sql.Tx, versionId int, buildingId int) error {
	args := m.Called(ctx, tx, versionId, buildingId)
	return args.Error(0)
}

func (m *MockRepositoryRateCard) FindPackagePrices(ctx context.Context, tx *sql.Tx, versionId int, take int, skip int, search string) ([]models.RateCardPackagePrice, error) {
	args := m.Called(ctx, tx, versionId, take, skip, search)
	return args.Get(0).([]models.RateCardPackagePrice), args.Error(1)
}

func (m *MockRepositoryRateCard) CountPackagePrices(ctx context.Context, tx *sql.Tx, versionId int, search string) (int, error) {
	args := m.Called(ctx, tx, versionId, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryRateCard) UpsertPackagePrice(ctx context.Context, tx *sql.Tx, p models.RateCardPackagePrice) error {
	args := m.Called(ctx, tx, p)
	return args.Error(0)
}

func (m *MockRepositoryRateCard) DeletePackagePrice(ctx context.Context, tx *sql.Tx, versionId int, packageId int) error {
	args := m.Called(ctx, tx, versionId, packageId)
	return args.Error(0)
}

func (m *MockRepositoryRateCard) ReplacePackageBuildings(ctx context.Context, tx *sql.Tx, versionId int, packageId int, buildingIds []int) error {
	args := m.Called(ctx, tx, versionId, packageId, buildingIds)
	return args.Error(0)
}

func (m *MockRepositoryRateCard) FindPackageBuildings(ctx context.Context, tx *sql.Tx, versionId int, packageId int) ([]models.RateCardPackageBuilding, error) {
	args := m.Called(ctx, tx, versionId, packageId)
	return args.Get(0).([]models.RateCardPackageBuilding), args.Error(1)
}

func (m *MockRepositoryRateCard) CopyPrices(ctx context.Context, tx *sql.Tx, fromVersionId int, toVersionId int) error {
	args := m.Called(ctx, tx, fromVersionId, toVersionId)
	return args.Error(0)
}

func (m *MockRepositoryRateCard) FindLivePackageBuildingIds(ctx context.Context, tx *sql.Tx, packageId int) ([]int, error) {
	args := m.Called(ctx, tx, packageId)
	return args.Get(0).([]int), args.Error(1)
}
