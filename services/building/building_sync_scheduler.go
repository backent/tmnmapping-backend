package building

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// StartBuildingSyncScheduler starts a background goroutine that syncs buildings from ERP periodically
func StartBuildingSyncScheduler(service ServiceBuildingInterface, logger *logrus.Logger, intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 30 // Default to 30 minutes
	}

	interval := time.Duration(intervalMinutes) * time.Minute
	ticker := time.NewTicker(interval)

	logger.WithField("interval", interval.String()).Info("Starting building sync scheduler")

	go func() {
		ctx := context.Background()

		// Initial sync on startup
		logger.Info("Running initial building sync")
		if err := service.SyncFromERP(ctx); err != nil {
			logger.WithError(err).Error("Initial building sync failed")
		}

		// Periodic sync
		for range ticker.C {
			logger.Info("Running scheduled building sync")
			if err := service.SyncFromERP(ctx); err != nil {
				logger.WithError(err).Error("Scheduled building sync failed")
			}
		}
	}()
}
