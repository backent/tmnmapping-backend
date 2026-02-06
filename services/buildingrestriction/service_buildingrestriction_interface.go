package buildingrestriction

import (
	"context"

	webBuildingRestriction "github.com/malikabdulaziz/tmn-backend/web/buildingrestriction"
)

type ServiceBuildingRestrictionInterface interface {
	Create(ctx context.Context, request webBuildingRestriction.CreateBuildingRestrictionRequest) webBuildingRestriction.BuildingRestrictionResponse
	FindAll(ctx context.Context, request webBuildingRestriction.BuildingRestrictionRequestFindAll) ([]webBuildingRestriction.BuildingRestrictionResponse, int)
	FindById(ctx context.Context, id int) webBuildingRestriction.BuildingRestrictionResponse
	Update(ctx context.Context, request webBuildingRestriction.UpdateBuildingRestrictionRequest, id int) webBuildingRestriction.BuildingRestrictionResponse
	Delete(ctx context.Context, id int)
}
