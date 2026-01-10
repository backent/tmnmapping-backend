package building

import (
	"context"

	webBuilding "github.com/malikabdulaziz/tmn-backend/web/building"
)

type ServiceBuildingInterface interface {
	FindById(ctx context.Context, id int) webBuilding.BuildingResponse
	FindAll(ctx context.Context, request webBuilding.BuildingRequestFindAll) ([]webBuilding.BuildingResponse, int)
	Update(ctx context.Context, request webBuilding.UpdateBuildingRequest, id int) webBuilding.BuildingResponse
	SyncFromERP(ctx context.Context) error
	GetFilterOptions(ctx context.Context) map[string][]string
	FindAllForMapping(ctx context.Context, request webBuilding.MappingBuildingRequest) webBuilding.MappingBuildingsResponse
}

