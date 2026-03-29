package poipoint

import (
	"context"

	webPOIPoint "github.com/malikabdulaziz/tmn-backend/web/poipoint"
)

type ServicePOIPointInterface interface {
	Create(ctx context.Context, request webPOIPoint.CreatePOIPointRequest) webPOIPoint.POIPointResponse
	FindAll(ctx context.Context, request webPOIPoint.POIPointRequestFindAll) ([]webPOIPoint.POIPointResponse, int)
	FindById(ctx context.Context, id int) webPOIPoint.POIPointResponse
	Update(ctx context.Context, request webPOIPoint.UpdatePOIPointRequest, id int) webPOIPoint.POIPointResponse
	Delete(ctx context.Context, id int)
	GetPointUsage(ctx context.Context, id int) webPOIPoint.POIPointUsageResponse
	Import(ctx context.Context, fileBytes []byte, fileType string) []webPOIPoint.POIPointResponse
	Export(ctx context.Context, search string) ([]byte, error)
}
