package salespackage

import (
	"context"

	webSalesPackage "github.com/malikabdulaziz/tmn-backend/web/salespackage"
)

type ServiceSalesPackageInterface interface {
	Create(ctx context.Context, request webSalesPackage.CreateSalesPackageRequest) webSalesPackage.SalesPackageResponse
	FindAll(ctx context.Context, request webSalesPackage.SalesPackageRequestFindAll) ([]webSalesPackage.SalesPackageResponse, int)
	FindById(ctx context.Context, id int) webSalesPackage.SalesPackageResponse
	Update(ctx context.Context, request webSalesPackage.UpdateSalesPackageRequest, id int) webSalesPackage.SalesPackageResponse
	Delete(ctx context.Context, id int)
}
