package buildingprice

import (
	"context"

	"github.com/malikabdulaziz/tmn-backend/web"
	webBuildingPrice "github.com/malikabdulaziz/tmn-backend/web/buildingprice"
)

type ServiceBuildingPriceInterface interface {
	FindAll(ctx context.Context, request webBuildingPrice.BuildingPriceRequestFindAll) ([]webBuildingPrice.BuildingPriceResponse, int)
	Upsert(ctx context.Context, request webBuildingPrice.UpsertBuildingPriceRequest) webBuildingPrice.BuildingPriceResponse
	Delete(ctx context.Context, buildingId int)

	// Import checks and counts every row; with dryRun it stops there and writes
	// nothing, so an upload can be previewed before it reprices anything. Rows that
	// fail are listed in the result and left out; the valid rows apply.
	Import(ctx context.Context, fileBytes []byte, fileType string, dryRun bool) web.ImportResult

	Export(ctx context.Context) ([]byte, error)
	Template() ([]byte, error)
}
