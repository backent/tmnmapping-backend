package salesassignment

import (
	"context"
	"database/sql"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositorySalesAssignmentImpl struct{}

func NewRepositorySalesAssignmentImpl() RepositorySalesAssignmentInterface {
	return &RepositorySalesAssignmentImpl{}
}

// Assignments are only ever useful with all three sides resolved, so every read
// joins customer, brand and the sales user.
var assignmentSelect = `SELECT a.id, a.customer_id, c.code, c.name, a.brand_id, b.code, b.name,
	a.sales_user_id, u.username, u.name, a.status, a.registration_date, a.expiry_date,
	a.created_at, a.updated_at
	FROM ` + models.SalesAssignmentTable + ` a
	JOIN ` + models.CustomerTable + ` c ON c.id = a.customer_id
	JOIN ` + models.BrandTable + ` b ON b.id = a.brand_id
	JOIN ` + models.UserTable + ` u ON u.id = a.sales_user_id`

var assignmentAllowedOrderBy = map[string]string{
	"id": "a.id", "status": "a.status", "created_at": "a.created_at",
	"updated_at": "a.updated_at", "customer_name": "c.name", "brand_name": "b.name",
	"sales_name": "u.name", "registration_date": "a.registration_date",
	"expiry_date": "a.expiry_date",
}

func assignmentSafeOrder(orderBy, orderDirection string) (string, string) {
	column, ok := assignmentAllowedOrderBy[orderBy]
	if !ok {
		column = "a.created_at"
	}
	if orderDirection != "ASC" && orderDirection != "DESC" {
		orderDirection = "DESC"
	}

	return column, orderDirection
}

const assignmentSearchClause = ` WHERE c.code ILIKE $1 OR c.name ILIKE $1 OR b.code ILIKE $1
	OR b.name ILIKE $1 OR u.username ILIKE $1 OR u.name ILIKE $1`

func scanAssignment(rows *sql.Rows) (models.SalesAssignment, error) {
	var n models.NullAbleSalesAssignment
	err := rows.Scan(&n.Id, &n.CustomerId, &n.CustomerCode, &n.CustomerName,
		&n.BrandId, &n.BrandCode, &n.BrandName, &n.SalesUserId, &n.SalesUsername,
		&n.SalesName, &n.Status, &n.RegistrationDate, &n.ExpiryDate, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.SalesAssignment{}, err
	}

	return models.NullAbleSalesAssignmentToSalesAssignment(n), nil
}

func (r *RepositorySalesAssignmentImpl) Create(ctx context.Context, tx *sql.Tx, a models.SalesAssignment) (models.SalesAssignment, error) {
	SQL := "INSERT INTO " + models.SalesAssignmentTable +
		" (customer_id, brand_id, sales_user_id, status, registration_date, expiry_date)" +
		" VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at"
	err := tx.QueryRowContext(ctx, SQL, a.CustomerId, a.BrandId, a.SalesUserId, a.Status,
		nullIfEmpty(a.RegistrationDate), nullIfEmpty(a.ExpiryDate)).
		Scan(&a.Id, &a.CreatedAt, &a.UpdatedAt)

	return a, err
}

func (r *RepositorySalesAssignmentImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.SalesAssignment, error) {
	column, direction := assignmentSafeOrder(orderBy, orderDirection)

	var rows *sql.Rows
	var err error

	if search != "" {
		SQL := assignmentSelect + assignmentSearchClause +
			" ORDER BY " + column + " " + direction + ", c.name ASC LIMIT $2 OFFSET $3"
		rows, err = tx.QueryContext(ctx, SQL, "%"+search+"%", take, skip)
	} else {
		SQL := assignmentSelect +
			" ORDER BY " + column + " " + direction + ", c.name ASC LIMIT $1 OFFSET $2"
		rows, err = tx.QueryContext(ctx, SQL, take, skip)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.SalesAssignment
	for rows.Next() {
		item, err := scanAssignment(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

func (r *RepositorySalesAssignmentImpl) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	base := "SELECT COUNT(*) FROM " + models.SalesAssignmentTable + " a" +
		" JOIN " + models.CustomerTable + " c ON c.id = a.customer_id" +
		" JOIN " + models.BrandTable + " b ON b.id = a.brand_id" +
		" JOIN " + models.UserTable + " u ON u.id = a.sales_user_id"

	var total int
	var err error
	if search != "" {
		err = tx.QueryRowContext(ctx, base+assignmentSearchClause, "%"+search+"%").Scan(&total)
	} else {
		err = tx.QueryRowContext(ctx, base).Scan(&total)
	}

	return total, err
}

func (r *RepositorySalesAssignmentImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.SalesAssignment, error) {
	rows, err := tx.QueryContext(ctx, assignmentSelect+" WHERE a.id = $1", id)
	if err != nil {
		return models.SalesAssignment{}, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanAssignment(rows)
	}

	return models.SalesAssignment{}, sql.ErrNoRows
}

func (r *RepositorySalesAssignmentImpl) FindByCustomerAndBrand(ctx context.Context, tx *sql.Tx, customerId int, brandId int) (models.SalesAssignment, error) {
	rows, err := tx.QueryContext(ctx,
		assignmentSelect+" WHERE a.customer_id = $1 AND a.brand_id = $2", customerId, brandId)
	if err != nil {
		return models.SalesAssignment{}, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanAssignment(rows)
	}

	return models.SalesAssignment{}, sql.ErrNoRows
}

func (r *RepositorySalesAssignmentImpl) Update(ctx context.Context, tx *sql.Tx, a models.SalesAssignment) (models.SalesAssignment, error) {
	SQL := "UPDATE " + models.SalesAssignmentTable +
		" SET customer_id = $1, brand_id = $2, sales_user_id = $3, status = $4," +
		" registration_date = $5, expiry_date = $6, updated_at = $7 WHERE id = $8 RETURNING updated_at"
	err := tx.QueryRowContext(ctx, SQL, a.CustomerId, a.BrandId, a.SalesUserId, a.Status,
		nullIfEmpty(a.RegistrationDate), nullIfEmpty(a.ExpiryDate), time.Now(), a.Id).
		Scan(&a.UpdatedAt)

	return a, err
}

func (r *RepositorySalesAssignmentImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM "+models.SalesAssignmentTable+" WHERE id = $1", id)

	return err
}

// CountBySalesUser backs the guard that stops a user being deleted while they still
// hold accounts. The FK is ON DELETE RESTRICT, so without this the operator would
// see a 500 instead of a readable message.
func (r *RepositorySalesAssignmentImpl) CountBySalesUser(ctx context.Context, tx *sql.Tx, salesUserId int) (int, error) {
	var total int
	err := tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM "+models.SalesAssignmentTable+" WHERE sales_user_id = $1",
		salesUserId).Scan(&total)

	return total, err
}

// nullIfEmpty keeps optional DATE columns NULL rather than "", which Postgres rejects.
func nullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}

	return value
}
