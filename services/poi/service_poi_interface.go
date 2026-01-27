package poi

import (
	"context"

	webPOI "github.com/malikabdulaziz/tmn-backend/web/poi"
)

type ServicePOIInterface interface {
	Create(ctx context.Context, request webPOI.CreatePOIRequest) webPOI.POIResponse
	FindAll(ctx context.Context) []webPOI.POIResponse
	FindById(ctx context.Context, id int) webPOI.POIResponse
	Update(ctx context.Context, request webPOI.UpdatePOIRequest, id int) webPOI.POIResponse
	Delete(ctx context.Context, id int)
}
