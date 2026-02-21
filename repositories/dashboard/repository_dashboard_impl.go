package dashboard

import (
	"context"
	"database/sql"
	"fmt"
)

// allowedDedupFields is the allowlist of valid column names for the DISTINCT ON field.
// This prevents SQL injection when the dedup field is configured via env var.
var allowedDedupFields = map[string]bool{
	"building_project":   true,
	"external_id":        true,
	"acquisition_person": true,
}

type RepositoryDashboardImpl struct{}

func NewRepositoryDashboardImpl() RepositoryDashboardInterface {
	return &RepositoryDashboardImpl{}
}

// validateDedupField returns dedupField if it is in the allowlist, otherwise returns "building_project".
func validateDedupField(dedupField string) string {
	if allowedDedupFields[dedupField] {
		return dedupField
	}
	return "building_project"
}

// GetStatusCounts returns per-workflow_state counts using DISTINCT ON to pick latest record per dedupField.
func (r *RepositoryDashboardImpl) GetStatusCounts(ctx context.Context, tx *sql.Tx, table, dedupField, pic, month string) ([]StatusCount, error) {
	dedup := validateDedupField(dedupField)

	SQL := fmt.Sprintf(`
		WITH latest AS (
			SELECT DISTINCT ON (%s) *
			FROM %s
			WHERE ($1 = '' OR acquisition_person = $1)
			  AND ($2 = '' OR TO_CHAR(created_at_erp, 'YYYY-MM') = $2)
			ORDER BY %s, modified DESC NULLS LAST
		)
		SELECT COALESCE(workflow_state, '') AS workflow_state, COUNT(*) AS count
		FROM latest
		GROUP BY workflow_state
		ORDER BY count DESC
	`, dedup, table, dedup)

	rows, err := tx.QueryContext(ctx, SQL, pic, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []StatusCount
	for rows.Next() {
		var sc StatusCount
		if err := rows.Scan(&sc.WorkflowState, &sc.Count); err != nil {
			return nil, err
		}
		results = append(results, sc)
	}
	return results, rows.Err()
}

// GetByPersonAndType returns counts grouped by acquisition_person and building_type,
// joined with buildings table to resolve building_type. Uses DISTINCT ON for dedup.
func (r *RepositoryDashboardImpl) GetByPersonAndType(ctx context.Context, tx *sql.Tx, table, dedupField, pic, month string) ([]PersonTypeCount, error) {
	dedup := validateDedupField(dedupField)

	SQL := fmt.Sprintf(`
		WITH latest AS (
			SELECT DISTINCT ON (t.%s) t.*
			FROM %s t
			WHERE ($1 = '' OR t.acquisition_person = $1)
			  AND ($2 = '' OR TO_CHAR(t.created_at_erp, 'YYYY-MM') = $2)
			ORDER BY t.%s, t.modified DESC NULLS LAST
		)
		SELECT
			l.acquisition_person,
			COALESCE(b.building_type, 'Unknown') AS building_type,
			COUNT(*) AS count
		FROM latest l
		LEFT JOIN buildings b ON b.project_name = l.building_project
		GROUP BY l.acquisition_person, b.building_type
		ORDER BY l.acquisition_person, count DESC
	`, dedup, table, dedup)

	rows, err := tx.QueryContext(ctx, SQL, pic, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PersonTypeCount
	for rows.Next() {
		var ptc PersonTypeCount
		if err := rows.Scan(&ptc.Person, &ptc.BuildingType, &ptc.Count); err != nil {
			return nil, err
		}
		results = append(results, ptc)
	}
	return results, rows.Err()
}

// GetByPersonAndStatus returns counts grouped by acquisition_person and workflow_state.
// Uses DISTINCT ON for dedup.
func (r *RepositoryDashboardImpl) GetByPersonAndStatus(ctx context.Context, tx *sql.Tx, table, dedupField, pic, month string) ([]PersonStatusCount, error) {
	dedup := validateDedupField(dedupField)

	SQL := fmt.Sprintf(`
		WITH latest AS (
			SELECT DISTINCT ON (%s) *
			FROM %s
			WHERE ($1 = '' OR acquisition_person = $1)
			  AND ($2 = '' OR TO_CHAR(created_at_erp, 'YYYY-MM') = $2)
			ORDER BY %s, modified DESC NULLS LAST
		)
		SELECT
			COALESCE(acquisition_person, '') AS acquisition_person,
			COALESCE(workflow_state, '') AS workflow_state,
			COUNT(*) AS count
		FROM latest
		GROUP BY acquisition_person, workflow_state
		ORDER BY acquisition_person, count DESC
	`, dedup, table, dedup)

	rows, err := tx.QueryContext(ctx, SQL, pic, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PersonStatusCount
	for rows.Next() {
		var psc PersonStatusCount
		if err := rows.Scan(&psc.Person, &psc.WorkflowState, &psc.Count); err != nil {
			return nil, err
		}
		results = append(results, psc)
	}
	return results, rows.Err()
}

// GetDistinctPICs returns the list of distinct acquisition_person values for the filter dropdown.
func (r *RepositoryDashboardImpl) GetDistinctPICs(ctx context.Context, tx *sql.Tx, table string) ([]string, error) {
	SQL := fmt.Sprintf(`
		SELECT DISTINCT acquisition_person
		FROM %s
		WHERE acquisition_person IS NOT NULL AND acquisition_person != ''
		ORDER BY acquisition_person
	`, table)

	rows, err := tx.QueryContext(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pics []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		pics = append(pics, p)
	}
	return pics, rows.Err()
}
