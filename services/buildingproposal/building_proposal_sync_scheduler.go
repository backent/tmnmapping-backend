package buildingproposal

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// StartBuildingProposalSyncScheduler starts a background goroutine that syncs building proposals from ERP periodically.
func StartBuildingProposalSyncScheduler(service ServiceBuildingProposalInterface, logger *logrus.Logger, intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 30
	}

	interval := time.Duration(intervalMinutes) * time.Minute
	ticker := time.NewTicker(interval)

	logger.WithField("interval", interval.String()).Info("Starting building proposal sync scheduler")

	go func() {
		ctx := context.Background()

		logger.Info("Running initial building proposal sync")
		if err := service.SyncFromERP(ctx); err != nil {
			logger.WithError(err).Error("Initial building proposal sync failed")
		}

		for range ticker.C {
			logger.Info("Running scheduled building proposal sync")
			if err := service.SyncFromERP(ctx); err != nil {
				logger.WithError(err).Error("Scheduled building proposal sync failed")
			}
		}
	}()
}
