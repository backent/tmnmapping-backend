package acquisition

import (
	"context"
	"database/sql"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/services/erp"
	"github.com/sirupsen/logrus"
)

type ServiceAcquisitionImpl struct {
	DB        *sql.DB
	ERPClient *erp.ERPClient
	Logger    *logrus.Logger
}

func NewServiceAcquisitionImpl(db *sql.DB, erpClient *erp.ERPClient, logger *logrus.Logger) ServiceAcquisitionInterface {
	return &ServiceAcquisitionImpl{
		DB:        db,
		ERPClient: erpClient,
		Logger:    logger,
	}
}

// SyncFromERP fetches all acquisitions from ERP and replaces local data (full refresh).
func (s *ServiceAcquisitionImpl) SyncFromERP(ctx context.Context) error {
	s.Logger.Info("Starting acquisition sync from ERP")

	erpRecords, err := s.ERPClient.FetchAcquisitions()
	if err != nil {
		s.Logger.WithError(err).Error("Failed to fetch acquisitions from ERP")
		return err
	}

	s.Logger.WithField("count", len(erpRecords)).Info("Fetched acquisitions from ERP")

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, "TRUNCATE TABLE "+models.AcquisitionTable)
	if err != nil {
		return err
	}

	now := time.Now()
	insertSQL := `INSERT INTO ` + models.AcquisitionTable + `
		(external_id, workflow_state, acquisition_person, building_project, status, modified, created_at_erp, synced_at, raw_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	inserted := 0
	for _, r := range erpRecords {
		modifiedTime := parseERPTime(r.Modified)
		creationTime := parseERPTime(r.Creation)

		_, err = tx.ExecContext(ctx, insertSQL,
			r.Name,
			r.WorkflowState,
			r.AcquisitionPerson,
			r.BuildingProject,
			r.Status,
			modifiedTime,
			creationTime,
			now,
			[]byte(r.RawJSON),
		)
		if err != nil {
			s.Logger.WithError(err).WithField("name", r.Name).Warn("Failed to insert acquisition, skipping")
			err = nil // continue on row-level error
			continue
		}
		inserted++
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	s.Logger.WithFields(logrus.Fields{
		"fetched":  len(erpRecords),
		"inserted": inserted,
	}).Info("Acquisition sync completed")

	return nil
}

// parseERPTime parses ERP timestamp strings. Returns nil on failure (stored as NULL).
func parseERPTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	formats := []string{
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return &t
		}
	}
	return nil
}
