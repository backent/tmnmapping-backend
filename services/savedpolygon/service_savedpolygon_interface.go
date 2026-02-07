package savedpolygon

import (
	"context"

	webSavedPolygon "github.com/malikabdulaziz/tmn-backend/web/savedpolygon"
)

type ServiceSavedPolygonInterface interface {
	Create(ctx context.Context, request webSavedPolygon.CreateSavedPolygonRequest) webSavedPolygon.SavedPolygonResponse
	FindAll(ctx context.Context, request webSavedPolygon.SavedPolygonRequestFindAll) ([]webSavedPolygon.SavedPolygonResponse, int)
	FindById(ctx context.Context, id int) webSavedPolygon.SavedPolygonResponse
	Update(ctx context.Context, request webSavedPolygon.UpdateSavedPolygonRequest, id int) webSavedPolygon.SavedPolygonResponse
	Delete(ctx context.Context, id int)
}
