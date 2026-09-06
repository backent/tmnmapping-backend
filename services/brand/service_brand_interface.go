package brand

import (
	"context"

	"github.com/malikabdulaziz/tmn-backend/web"
	webBrand "github.com/malikabdulaziz/tmn-backend/web/brand"
)

type ServiceBrandInterface interface {
	Create(ctx context.Context, request webBrand.CreateBrandRequest) webBrand.BrandResponse
	FindAll(ctx context.Context, request webBrand.BrandRequestFindAll) ([]webBrand.BrandResponse, int)
	FindById(ctx context.Context, id int) webBrand.BrandResponse
	Update(ctx context.Context, request webBrand.UpdateBrandRequest, id int) webBrand.BrandResponse
	Delete(ctx context.Context, id int)
	Import(ctx context.Context, fileBytes []byte, fileType string) web.ImportResult
	Export(ctx context.Context, search string) ([]byte, error)
	Template() ([]byte, error)
}
