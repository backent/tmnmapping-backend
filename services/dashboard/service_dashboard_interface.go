package dashboard

import (
	"context"

	webDashboard "github.com/malikabdulaziz/tmn-backend/web/dashboard"
)

type ServiceDashboardInterface interface {
	GetAcquisitionReport(ctx context.Context, pic, year, month string) webDashboard.DashboardReport
	GetBuildingProposalReport(ctx context.Context, pic, year, month string) webDashboard.DashboardReport
	GetLOIReport(ctx context.Context, pic, year, month string) webDashboard.DashboardReport
}
