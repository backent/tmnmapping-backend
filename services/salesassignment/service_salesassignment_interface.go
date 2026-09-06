package salesassignment

import (
	"context"

	"github.com/malikabdulaziz/tmn-backend/web"
	webSalesAssignment "github.com/malikabdulaziz/tmn-backend/web/salesassignment"
)

type ServiceSalesAssignmentInterface interface {
	Create(ctx context.Context, request webSalesAssignment.CreateSalesAssignmentRequest) webSalesAssignment.SalesAssignmentResponse
	FindAll(ctx context.Context, request webSalesAssignment.SalesAssignmentRequestFindAll) ([]webSalesAssignment.SalesAssignmentResponse, int)
	FindById(ctx context.Context, id int) webSalesAssignment.SalesAssignmentResponse
	Update(ctx context.Context, request webSalesAssignment.UpdateSalesAssignmentRequest, id int) webSalesAssignment.SalesAssignmentResponse
	Delete(ctx context.Context, id int)
	Import(ctx context.Context, fileBytes []byte, fileType string) web.ImportResult
	Export(ctx context.Context, search string) ([]byte, error)
	Template() ([]byte, error)
}
