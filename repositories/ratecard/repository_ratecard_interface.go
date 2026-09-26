package ratecard

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryRateCardInterface interface {
	// Versions
	CreateVersion(ctx context.Context, tx *sql.Tx, version models.RateCardVersion) (models.RateCardVersion, error)
	FindAllVersions(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.RateCardVersion, error)
	CountAllVersions(ctx context.Context, tx *sql.Tx, search string) (int, error)
	FindVersionById(ctx context.Context, tx *sql.Tx, id int) (models.RateCardVersion, error)
	FindVersionByCode(ctx context.Context, tx *sql.Tx, code string) (models.RateCardVersion, error)
	FindCurrentVersion(ctx context.Context, tx *sql.Tx) (models.RateCardVersion, error)
	UpdateVersion(ctx context.Context, tx *sql.Tx, version models.RateCardVersion) (models.RateCardVersion, error)
	DeleteVersion(ctx context.Context, tx *sql.Tx, id int) error
	SetVersionStatus(ctx context.Context, tx *sql.Tx, id int, status string, publishedByUserId int) error
	DemoteCurrentVersion(ctx context.Context, tx *sql.Tx) error

	// Building prices
	FindBuildingPrices(ctx context.Context, tx *sql.Tx, versionId int, take int, skip int, search string) ([]models.RateCardBuildingPrice, error)
	CountBuildingPrices(ctx context.Context, tx *sql.Tx, versionId int, search string) (int, error)
	UpsertBuildingPrice(ctx context.Context, tx *sql.Tx, price models.RateCardBuildingPrice) error
	DeleteBuildingPrice(ctx context.Context, tx *sql.Tx, versionId int, buildingId int) error

	// Package prices and the composition as priced
	FindPackagePrices(ctx context.Context, tx *sql.Tx, versionId int, take int, skip int, search string) ([]models.RateCardPackagePrice, error)
	CountPackagePrices(ctx context.Context, tx *sql.Tx, versionId int, search string) (int, error)
	UpsertPackagePrice(ctx context.Context, tx *sql.Tx, price models.RateCardPackagePrice) error
	DeletePackagePrice(ctx context.Context, tx *sql.Tx, versionId int, packageId int) error
	ReplacePackageBuildings(ctx context.Context, tx *sql.Tx, versionId int, packageId int, buildingIds []int) error
	FindPackageBuildings(ctx context.Context, tx *sql.Tx, versionId int, packageId int) ([]models.RateCardPackageBuilding, error)

	// Copying an existing version into a new draft
	CopyPrices(ctx context.Context, tx *sql.Tx, fromVersionId int, toVersionId int) error

	// FindLivePackageBuildingIds reads the CURRENT membership from the sales package
	// master, which publishing freezes into rate_card_package_buildings.
	FindLivePackageBuildingIds(ctx context.Context, tx *sql.Tx, packageId int) ([]int, error)
}
