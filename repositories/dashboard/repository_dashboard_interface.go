package dashboard

import (
	"context"
	"database/sql"
)

// StatusCount holds a workflow_state and its count.
type StatusCount struct {
	WorkflowState string
	Count         int
}

// PersonTypeCount holds counts per building_type for one person.
type PersonTypeCount struct {
	Person       string
	BuildingType string
	Count        int
}

// PersonStatusCount holds counts per workflow_state for one person.
type PersonStatusCount struct {
	Person        string
	WorkflowState string
	Count         int
}

type RepositoryDashboardInterface interface {
	GetStatusCounts(ctx context.Context, tx *sql.Tx, table, dedupField, pic, dateFrom, dateTo string) ([]StatusCount, error)
	GetByPersonAndType(ctx context.Context, tx *sql.Tx, table, dedupField, pic, dateFrom, dateTo string) ([]PersonTypeCount, error)
	GetByPersonAndStatus(ctx context.Context, tx *sql.Tx, table, dedupField, pic, dateFrom, dateTo string) ([]PersonStatusCount, error)
	GetDistinctPICs(ctx context.Context, tx *sql.Tx, table string) ([]string, error)
}
