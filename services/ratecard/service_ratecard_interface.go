package ratecard

import (
	"context"

	"github.com/malikabdulaziz/tmn-backend/web"
	webRateCard "github.com/malikabdulaziz/tmn-backend/web/ratecard"
)

type ServiceRateCardInterface interface {
	CreateVersion(ctx context.Context, request webRateCard.CreateVersionRequest) webRateCard.VersionResponse
	FindAllVersions(ctx context.Context, request webRateCard.RateCardRequestFindAll) ([]webRateCard.VersionResponse, int)
	FindVersionById(ctx context.Context, id int) webRateCard.VersionResponse
	FindCurrentVersion(ctx context.Context) webRateCard.VersionResponse
	UpdateVersion(ctx context.Context, request webRateCard.UpdateVersionRequest, id int) webRateCard.VersionResponse
	DeleteVersion(ctx context.Context, id int)
	PublishVersion(ctx context.Context, id int, actorUserId int) webRateCard.VersionResponse

	FindBuildingPrices(ctx context.Context, versionId int, request webRateCard.RateCardRequestFindAll) ([]webRateCard.BuildingPriceResponse, int)
	UpsertBuildingPrice(ctx context.Context, versionId int, request webRateCard.UpsertBuildingPriceRequest) webRateCard.BuildingPriceResponse
	DeleteBuildingPrice(ctx context.Context, versionId int, buildingId int)
	ImportBuildingPrices(ctx context.Context, versionId int, fileBytes []byte, fileType string) web.ImportResult
	ExportBuildingPrices(ctx context.Context, versionId int) ([]byte, error)
	BuildingPriceTemplate() ([]byte, error)

	FindPackagePrices(ctx context.Context, versionId int, request webRateCard.RateCardRequestFindAll) ([]webRateCard.PackagePriceResponse, int)
	UpsertPackagePrice(ctx context.Context, versionId int, request webRateCard.UpsertPackagePriceRequest) webRateCard.PackagePriceResponse
	DeletePackagePrice(ctx context.Context, versionId int, packageId int)
	ImportPackagePrices(ctx context.Context, versionId int, fileBytes []byte, fileType string) web.ImportResult
	ExportPackagePrices(ctx context.Context, versionId int) ([]byte, error)
	PackagePriceTemplate() ([]byte, error)
}
