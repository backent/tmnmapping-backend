package loi

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// StartLOISyncScheduler starts a background goroutine that syncs LOIs from ERP periodically.
func StartLOISyncScheduler(service ServiceLOIInterface, logger *logrus.Logger, intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 30
	}

	interval := time.Duration(intervalMinutes) * time.Minute
	ticker := time.NewTicker(interval)

	logger.WithField("interval", interval.String()).Info("Starting LOI sync scheduler")

	go func() {
		ctx := context.Background()

		logger.Info("Running initial LOI sync")
		if err := service.SyncFromERP(ctx); err != nil {
			logger.WithError(err).Error("Initial LOI sync failed")
		}

		for range ticker.C {
			logger.Info("Running scheduled LOI sync")
			if err := service.SyncFromERP(ctx); err != nil {
				logger.WithError(err).Error("Scheduled LOI sync failed")
			}
		}
	}()
}
