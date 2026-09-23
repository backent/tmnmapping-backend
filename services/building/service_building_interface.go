package building

import (
	"context"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/web"
	webBuilding "github.com/malikabdulaziz/tmn-backend/web/building"
)

type ServiceBuildingInterface interface {
	FindById(ctx context.Context, id int) webBuilding.BuildingResponse
	FindAll(ctx context.Context, request webBuilding.BuildingRequestFindAll) ([]webBuilding.BuildingResponse, int)
	Update(ctx context.Context, request webBuilding.UpdateBuildingRequest, id int) webBuilding.BuildingResponse
	SyncFromERP(ctx context.Context) error
	GetFilterOptions(ctx context.Context) map[string][]string
	FindAllForMapping(ctx context.Context, request webBuilding.MappingBuildingRequest) webBuilding.MappingBuildingsResponse
	ExportForMapping(ctx context.Context, ids []int) ([]byte, error)
	ExportForMappingWithFilters(ctx context.Context, request webBuilding.MappingBuildingRequest) ([]byte, error)
	GetLCDPresenceSummary(ctx context.Context) webBuilding.LCDPresenceSummaryResponse
	FindAllDropdown(ctx context.Context) []webBuilding.BuildingDropdownResponse

	// Spreadsheet maintenance. A blank cell CLEARS on this import, so the result
	// reports cleared fields separately from updated rows and every change is
	// written to building_changes -- the undo trail that makes it survivable.
	Import(ctx context.Context, fileBytes []byte, fileType string, dryRun bool, actor Actor) web.ImportResult
	Export(ctx context.Context) ([]byte, error)
	Template() ([]byte, error)
	FindChanges(ctx context.Context, buildingId int, take int, skip int) ([]models.BuildingChange, int)
}
