package acquisition

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// StartAcquisitionSyncScheduler starts a background goroutine that syncs acquisitions from ERP periodically.
func StartAcquisitionSyncScheduler(service ServiceAcquisitionInterface, logger *logrus.Logger, intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 30
	}

	interval := time.Duration(intervalMinutes) * time.Minute
	ticker := time.NewTicker(interval)

	logger.WithField("interval", interval.String()).Info("Starting acquisition sync scheduler")

	go func() {
		ctx := context.Background()

		logger.Info("Running initial acquisition sync")
		if err := service.SyncFromERP(ctx); err != nil {
			logger.WithError(err).Error("Initial acquisition sync failed")
		}

		for range ticker.C {
			logger.Info("Running scheduled acquisition sync")
			if err := service.SyncFromERP(ctx); err != nil {
				logger.WithError(err).Error("Scheduled acquisition sync failed")
			}
		}
	}()
}
