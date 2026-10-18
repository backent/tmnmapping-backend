package dashboard

import (
	"context"

	webDashboard "github.com/malikabdulaziz/tmn-backend/web/dashboard"
)

type ServiceDashboardInterface interface {
	GetLOIReport(ctx context.Context, pic, dateFrom, dateTo string) webDashboard.DashboardReport
}
