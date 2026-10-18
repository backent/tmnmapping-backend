package buildingproject

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBuildingProjectImpl struct{}

func NewRepositoryBuildingProjectImpl() RepositoryBuildingProjectInterface {
	return &RepositoryBuildingProjectImpl{}
}

// Dates are rendered in SQL rather than scanned as time.Time and formatted in Go.
// DATE columns come back from the driver as timestamps, and formatting them in one
// place -- here -- keeps every caller free of timezone questions the data never had.

// buildingCount is a correlated subquery rather than a GROUP BY: the list is paged,
// and grouping the whole buildings table to page 25 projects reads far more than it
// needs to. buildings.project_id arrives in migration 024.
var buildingCount = `(SELECT COUNT(*) FROM ` + models.BuildingTable + ` b WHERE b.project_id = p.id)`

// selectColumns builds the column list. When includeFinance is false the three gated
// columns are replaced by typed NULLs, so the query returns the same shape and the
// values never leave the database.
func selectColumns(includeFinance bool) string {
	rental, company, contractNo := "NULL::bigint", "NULL::varchar", "NULL::varchar"
	if includeFinance {
		rental, company, contractNo = "p.annual_rental", "p.company_name", "p.contract_no"
	}

	return `p.id, p.project_id_iris, p.name, p.building_type, p.grade, p.pic,
		p.tmn_project_status, p.no_of_tower, p.no_of_screen, ` +
		`TO_CHAR(p.created_date,'YYYY-MM-DD'), p.remark, p.contract_type, ` + contractNo + `, ` +
		`TO_CHAR(p.contract_date,'YYYY-MM-DD'), TO_CHAR(p.contract_start,'YYYY-MM-DD'),
		TO_CHAR(p.contract_end,'YYYY-MM-DD'), p.period_month, ` + rental + `, p.payment_term, ` +
		company + `, p.exclusivity, p.doc_type, p.contract_status, ` +
		`TO_CHAR(p.cancelled_at,'YYYY-MM-DD'), p.cancel_last_status, p.cancel_reason, ` +
		buildingCount + `, p.created_at, p.updated_at`
}

func scanProject(scanner interface{ Scan(...interface{}) error }) (models.BuildingProject, error) {
	var n models.NullAbleBuildingProject
	err := scanner.Scan(&n.Id, &n.ProjectIdIris, &n.Name, &n.BuildingType, &n.Grade, &n.Pic,
		&n.TmnProjectStatus, &n.NoOfTower, &n.NoOfScreen, &n.CreatedDate, &n.Remark,
		&n.ContractType, &n.ContractNo, &n.ContractDate, &n.ContractStart, &n.ContractEnd,
		&n.PeriodMonth, &n.AnnualRental, &n.PaymentTerm, &n.CompanyName, &n.Exclusivity,
		&n.DocType, &n.ContractStatus, &n.CancelledAt, &n.CancelLastStatus, &n.CancelReason,
		&n.BuildingCount, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.BuildingProject{}, err
	}

	return models.NullAbleBuildingProjectToBuildingProject(n), nil
}

var allowedOrderBy = map[string]string{
	"id": "p.id", "project_id_iris": "p.project_id_iris", "name": "p.name",
	"tmn_project_status": "p.tmn_project_status", "pic": "p.pic",
	"contract_start": "p.contract_start", "contract_end": "p.contract_end",
	"no_of_screen": "p.no_of_screen", "no_of_tower": "p.no_of_tower",
	"created_at": "p.created_at", "updated_at": "p.updated_at",
}

// safeOrder maps a caller-supplied sort onto a known column. annual_rental is
// deliberately absent: ordering by it would let a caller without the finance
// permission read the ranking, which is most of the value of the number.
func safeOrder(orderBy, orderDirection string) (string, string) {
	column, ok := allowedOrderBy[orderBy]
	if !ok {
		column = "p.created_at"
	}
	if orderDirection != "ASC" && orderDirection != "DESC" {
		orderDirection = "DESC"
	}

	return column, orderDirection
}

// filterClause builds the shared WHERE for list and count. Placeholders are numbered
// from the args already collected so the two callers can prepend their own.
func filterClause(filter ListFilter) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	next := func(v interface{}) string {
		args = append(args, v)

		return "$" + strconv.Itoa(len(args))
	}

	if search := strings.TrimSpace(filter.Search); search != "" {
		p := next("%" + search + "%")
		clauses = append(clauses, fmt.Sprintf(
			"(p.project_id_iris ILIKE %s OR p.name ILIKE %s OR p.pic ILIKE %s)", p, p, p))
	}
	if filter.Status != "" {
		clauses = append(clauses, "p.tmn_project_status = "+next(filter.Status))
	}
	if filter.ContractType != "" {
		clauses = append(clauses, "p.contract_type = "+next(filter.ContractType))
	}
	if filter.Pic != "" {
		clauses = append(clauses, "p.pic = "+next(filter.Pic))
	}

	if len(clauses) == 0 {
		return "", nil
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *RepositoryBuildingProjectImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, filter ListFilter, includeFinance bool) ([]models.BuildingProject, error) {
	where, args := filterClause(filter)
	column, direction := safeOrder(orderBy, orderDirection)

	SQL := `SELECT ` + selectColumns(includeFinance) + ` FROM ` + models.BuildingProjectTable + ` p` + where +
		fmt.Sprintf(" ORDER BY %s %s, p.id DESC LIMIT $%d OFFSET $%d", column, direction, len(args)+1, len(args)+2)
	args = append(args, take, skip)

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []models.BuildingProject{}
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (r *RepositoryBuildingProjectImpl) CountAll(ctx context.Context, tx *sql.Tx, filter ListFilter) (int, error) {
	where, args := filterClause(filter)
	SQL := `SELECT COUNT(*) FROM ` + models.BuildingProjectTable + ` p` + where

	var total int
	err := tx.QueryRowContext(ctx, SQL, args...).Scan(&total)

	return total, err
}

func (r *RepositoryBuildingProjectImpl) FindById(ctx context.Context, tx *sql.Tx, id int, includeFinance bool) (models.BuildingProject, error) {
	SQL := `SELECT ` + selectColumns(includeFinance) + ` FROM ` + models.BuildingProjectTable + ` p WHERE p.id = $1`

	return scanProject(tx.QueryRowContext(ctx, SQL, id))
}

func (r *RepositoryBuildingProjectImpl) FindByIris(ctx context.Context, tx *sql.Tx, iris string, includeFinance bool) (models.BuildingProject, error) {
	SQL := `SELECT ` + selectColumns(includeFinance) + ` FROM ` + models.BuildingProjectTable + ` p WHERE p.project_id_iris = $1`

	return scanProject(tx.QueryRowContext(ctx, SQL, iris))
}

// writeArgs is the column order shared by Create and Update, so the two cannot drift.
// Empty strings become NULL: the model has no null-string, and storing ” would make
// "never filled in" and "deliberately cleared" indistinguishable in the database.
func writeArgs(p models.BuildingProject) []interface{} {
	return []interface{}{
		p.ProjectIdIris, p.Name, nullIfEmpty(p.BuildingType), nullIfEmpty(p.Grade),
		nullIfEmpty(p.Pic), nullIfEmpty(p.TmnProjectStatus),
		nullIfZero(int64(p.NoOfTower)), nullIfZero(int64(p.NoOfScreen)),
		nullIfEmpty(p.CreatedDate), nullIfEmpty(p.Remark),
		nullIfEmpty(p.ContractType), nullIfEmpty(p.ContractNo), nullIfEmpty(p.ContractDate),
		nullIfEmpty(p.ContractStart), nullIfEmpty(p.ContractEnd),
		nullIfZero(int64(p.PeriodMonth)), nullIfZero(p.AnnualRental),
		nullIfEmpty(p.PaymentTerm), nullIfEmpty(p.CompanyName), nullIfEmpty(p.Exclusivity),
		nullIfEmpty(p.DocType), nullIfEmpty(p.ContractStatus),
		nullIfEmpty(p.CancelledAt), nullIfEmpty(p.CancelLastStatus), nullIfEmpty(p.CancelReason),
	}
}

func nullIfEmpty(v string) interface{} {
	if strings.TrimSpace(v) == "" {
		return nil
	}

	return v
}

func nullIfZero(v int64) interface{} {
	if v == 0 {
		return nil
	}

	return v
}

const writeColumns = `project_id_iris, name, building_type, grade, pic, tmn_project_status,
	no_of_tower, no_of_screen, created_date, remark, contract_type, contract_no,
	contract_date, contract_start, contract_end, period_month, annual_rental,
	payment_term, company_name, exclusivity, doc_type, contract_status,
	cancelled_at, cancel_last_status, cancel_reason`

// The DATE columns arrive as strings and are cast explicitly, so an empty value is a
// NULL rather than a cast error on ”.
const writePlaceholders = `$1, $2, $3, $4, $5, $6, $7, $8, $9::date, $10, $11, $12,
	$13::date, $14::date, $15::date, $16, $17, $18, $19, $20, $21, $22, $23::date, $24, $25`

func (r *RepositoryBuildingProjectImpl) Create(ctx context.Context, tx *sql.Tx, project models.BuildingProject) (models.BuildingProject, error) {
	SQL := `INSERT INTO ` + models.BuildingProjectTable + ` (` + writeColumns + `) VALUES (` +
		writePlaceholders + `) RETURNING id`

	var id int
	if err := tx.QueryRowContext(ctx, SQL, writeArgs(project)...).Scan(&id); err != nil {
		return models.BuildingProject{}, err
	}

	project.Id = id

	return project, nil
}

func (r *RepositoryBuildingProjectImpl) Update(ctx context.Context, tx *sql.Tx, project models.BuildingProject) (models.BuildingProject, error) {
	SQL := `UPDATE ` + models.BuildingProjectTable + ` SET
		project_id_iris = $1, name = $2, building_type = $3, grade = $4, pic = $5,
		tmn_project_status = $6, no_of_tower = $7, no_of_screen = $8,
		created_date = $9::date, remark = $10, contract_type = $11, contract_no = $12,
		contract_date = $13::date, contract_start = $14::date, contract_end = $15::date,
		period_month = $16, annual_rental = $17, payment_term = $18, company_name = $19,
		exclusivity = $20, doc_type = $21, contract_status = $22,
		cancelled_at = $23::date, cancel_last_status = $24, cancel_reason = $25,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = $26`

	args := append(writeArgs(project), project.Id)
	if _, err := tx.ExecContext(ctx, SQL, args...); err != nil {
		return models.BuildingProject{}, err
	}

	return project, nil
}

func (r *RepositoryBuildingProjectImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM `+models.BuildingProjectTable+` WHERE id = $1`, id)

	return err
}

// changesPerInsert keeps one statement well inside Postgres' 65535-parameter limit at
// 9 parameters per row. An import of a few thousand projects is a handful of
// statements rather than a few thousand round trips.
const changeColumns = 10
const changesPerInsert = 500

func (r *RepositoryBuildingProjectImpl) RecordChanges(ctx context.Context, tx *sql.Tx, changes []models.BuildingProjectChange) error {
	if len(changes) == 0 {
		return nil
	}

	for start := 0; start < len(changes); start += changesPerInsert {
		end := start + changesPerInsert
		if end > len(changes) {
			end = len(changes)
		}

		SQL, args := buildChangesInsert(changes[start:end])
		if _, err := tx.ExecContext(ctx, SQL, args...); err != nil {
			return err
		}
	}

	return nil
}

func buildChangesInsert(changes []models.BuildingProjectChange) (string, []interface{}) {
	placeholders := make([]string, 0, len(changes))
	args := make([]interface{}, 0, len(changes)*changeColumns)

	for i, change := range changes {
		base := i * changeColumns
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d::uuid, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10))

		args = append(args,
			nullIfZero(int64(change.ProjectId)), change.ProjectIdIris,
			nullIfZero(int64(change.ActorUserId)), nullIfEmpty(change.ActorRole),
			change.Action, change.Source, nullIfEmpty(change.BatchId),
			nullIfEmpty(change.Field), nullIfEmpty(change.OldValue), nullIfEmpty(change.NewValue))
	}

	return `INSERT INTO ` + models.BuildingProjectChangeTable + `
		(project_id, project_id_iris, actor_user_id, actor_role, action, source, batch_id,
		 field, old_value, new_value)
		VALUES ` + strings.Join(placeholders, ", "), args
}

func (r *RepositoryBuildingProjectImpl) FindChanges(ctx context.Context, tx *sql.Tx, projectId int, take int, skip int) ([]models.BuildingProjectChange, error) {
	SQL := `SELECT c.id, c.project_id, c.project_id_iris, c.actor_user_id, u.name, c.actor_role,
		c.action, c.source, c.batch_id, c.field, c.old_value, c.new_value, c.created_at
		FROM ` + models.BuildingProjectChangeTable + ` c
		LEFT JOIN ` + models.UserTable + ` u ON u.id = c.actor_user_id
		WHERE c.project_id = $1
		ORDER BY c.created_at DESC, c.id DESC LIMIT $2 OFFSET $3`

	rows, err := tx.QueryContext(ctx, SQL, projectId, take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changes := []models.BuildingProjectChange{}
	for rows.Next() {
		var n models.NullAbleBuildingProjectChange
		if err := rows.Scan(&n.Id, &n.ProjectId, &n.ProjectIdIris, &n.ActorUserId, &n.ActorName,
			&n.ActorRole, &n.Action, &n.Source, &n.BatchId, &n.Field, &n.OldValue,
			&n.NewValue, &n.CreatedAt); err != nil {
			return nil, err
		}
		changes = append(changes, models.NullAbleBuildingProjectChangeToChange(n))
	}

	return changes, rows.Err()
}

func (r *RepositoryBuildingProjectImpl) CountChanges(ctx context.Context, tx *sql.Tx, projectId int) (int, error) {
	var total int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM `+models.BuildingProjectChangeTable+` WHERE project_id = $1`,
		projectId).Scan(&total)

	return total, err
}
