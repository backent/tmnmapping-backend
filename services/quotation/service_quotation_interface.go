package quotation

import (
	"context"

	webQuotation "github.com/malikabdulaziz/tmn-backend/web/quotation"
)

// Actor is the caller, resolved by RequireAuth. Quotation rules depend on both who
// you are and what role you hold, so both travel together.
type Actor struct {
	UserId int
	Role   string
}

type ServiceQuotationInterface interface {
	Create(ctx context.Context, request webQuotation.CreateQuotationRequest, actor Actor) webQuotation.QuotationResponse
	FindAll(ctx context.Context, request webQuotation.QuotationRequestFindAll, actor Actor) ([]webQuotation.QuotationResponse, int)
	FindById(ctx context.Context, id int, actor Actor) webQuotation.QuotationResponse
	Update(ctx context.Context, request webQuotation.UpdateQuotationRequest, id int, actor Actor) webQuotation.QuotationResponse
	Delete(ctx context.Context, id int, actor Actor)

	Submit(ctx context.Context, id int, actor Actor) webQuotation.QuotationResponse
	Approve(ctx context.Context, id int, actor Actor) webQuotation.QuotationResponse
	Return(ctx context.Context, id int, comment string, actor Actor) webQuotation.QuotationResponse

	PreviewPricing(ctx context.Context, request webQuotation.PricingPreviewRequest, actor Actor) webQuotation.PricingPreviewResponse
	DashboardCounts(ctx context.Context, actor Actor) webQuotation.DashboardCounts
}
