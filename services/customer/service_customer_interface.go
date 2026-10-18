package customer

import (
	"context"

	"github.com/malikabdulaziz/tmn-backend/web"
	webCustomer "github.com/malikabdulaziz/tmn-backend/web/customer"
)

type ServiceCustomerInterface interface {
	Create(ctx context.Context, request webCustomer.CreateCustomerRequest) webCustomer.CustomerResponse
	FindAll(ctx context.Context, request webCustomer.CustomerRequestFindAll) ([]webCustomer.CustomerResponse, int)
	FindById(ctx context.Context, id int) webCustomer.CustomerResponse
	Update(ctx context.Context, request webCustomer.UpdateCustomerRequest, id int) webCustomer.CustomerResponse
	Delete(ctx context.Context, id int)
	Import(ctx context.Context, fileBytes []byte, fileType string) web.ImportResult
	Export(ctx context.Context, search string) ([]byte, error)
	Template() ([]byte, error)
}
