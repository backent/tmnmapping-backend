package motherbrand

import (
	"context"

	webMotherBrand "github.com/malikabdulaziz/tmn-backend/web/motherbrand"
)

type ServiceMotherBrandInterface interface {
	Create(ctx context.Context, request webMotherBrand.CreateMotherBrandRequest) webMotherBrand.MotherBrandResponse
	FindAll(ctx context.Context, request webMotherBrand.MotherBrandRequestFindAll) ([]webMotherBrand.MotherBrandResponse, int)
	FindById(ctx context.Context, id int) webMotherBrand.MotherBrandResponse
	Update(ctx context.Context, request webMotherBrand.UpdateMotherBrandRequest, id int) webMotherBrand.MotherBrandResponse
	Delete(ctx context.Context, id int)
	Import(ctx context.Context, fileBytes []byte, fileType string) []webMotherBrand.MotherBrandResponse
	Export(ctx context.Context, search string) ([]byte, error)
}
