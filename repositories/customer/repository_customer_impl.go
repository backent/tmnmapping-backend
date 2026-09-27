package customer

import (
	"context"
	"database/sql"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryCustomerImpl struct{}

func NewRepositoryCustomerImpl() RepositoryCustomerInterface {
	return &RepositoryCustomerImpl{}
}

const customerColumns = "id, code, name, industry, status, created_at, updated_at"

var customerAllowedOrderBy = map[string]bool{
	"id": true, "code": true, "name": true, "industry": true,
	"status": true, "created_at": true, "updated_at": true,
}

func customerSafeOrder(orderBy, orderDirection string) (string, string) {
	if !customerAllowedOrderBy[orderBy] {
		orderBy = "created_at"
	}
	if orderDirection != "ASC" && orderDirection != "DESC" {
		orderDirection = "DESC"
	}

	return orderBy, orderDirection
}

func scanCustomer(rows *sql.Rows) (models.Customer, error) {
	var n models.NullAbleCustomer
	if err := rows.Scan(&n.Id, &n.Code, &n.Name, &n.Industry, &n.Status, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return models.Customer{}, err
	}

	return models.NullAbleCustomerToCustomer(n), nil
}

func (r *RepositoryCustomerImpl) Create(ctx context.Context, tx *sql.Tx, customer models.Customer) (models.Customer, error) {
	SQL := "INSERT INTO " + models.CustomerTable +
		" (code, name, industry, status) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at"
	err := tx.QueryRowContext(ctx, SQL, customer.Code, customer.Name, customer.Industry, customer.Status).
		Scan(&customer.Id, &customer.CreatedAt, &customer.UpdatedAt)

	return customer, err
}

func (r *RepositoryCustomerImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.Customer, error) {
	orderBy, orderDirection = customerSafeOrder(orderBy, orderDirection)

	var rows *sql.Rows
	var err error

	if search != "" {
		SQL := "SELECT " + customerColumns + " FROM " + models.CustomerTable +
			" WHERE code ILIKE $1 OR name ILIKE $1 OR industry ILIKE $1" +
			" ORDER BY " + orderBy + " " + orderDirection + ", name ASC LIMIT $2 OFFSET $3"
		rows, err = tx.QueryContext(ctx, SQL, "%"+search+"%", take, skip)
	} else {
		SQL := "SELECT " + customerColumns + " FROM " + models.CustomerTable +
			" ORDER BY " + orderBy + " " + orderDirection + ", name ASC LIMIT $1 OFFSET $2"
		rows, err = tx.QueryContext(ctx, SQL, take, skip)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Customer
	for rows.Next() {
		item, err := scanCustomer(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

func (r *RepositoryCustomerImpl) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	var total int
	var err error

	if search != "" {
		SQL := "SELECT COUNT(*) FROM " + models.CustomerTable +
			" WHERE code ILIKE $1 OR name ILIKE $1 OR industry ILIKE $1"
		err = tx.QueryRowContext(ctx, SQL, "%"+search+"%").Scan(&total)
	} else {
		err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+models.CustomerTable).Scan(&total)
	}

	return total, err
}

func (r *RepositoryCustomerImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Customer, error) {
	return r.findOne(ctx, tx, "id = $1", id)
}

func (r *RepositoryCustomerImpl) FindByCode(ctx context.Context, tx *sql.Tx, code string) (models.Customer, error) {
	return r.findOne(ctx, tx, "code = $1", code)
}

func (r *RepositoryCustomerImpl) findOne(ctx context.Context, tx *sql.Tx, where string, arg interface{}) (models.Customer, error) {
	SQL := "SELECT " + customerColumns + " FROM " + models.CustomerTable + " WHERE " + where
	rows, err := tx.QueryContext(ctx, SQL, arg)
	if err != nil {
		return models.Customer{}, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanCustomer(rows)
	}

	return models.Customer{}, sql.ErrNoRows
}

func (r *RepositoryCustomerImpl) Update(ctx context.Context, tx *sql.Tx, customer models.Customer) (models.Customer, error) {
	SQL := "UPDATE " + models.CustomerTable +
		" SET code = $1, name = $2, industry = $3, status = $4, updated_at = $5 WHERE id = $6 RETURNING updated_at"
	err := tx.QueryRowContext(ctx, SQL, customer.Code, customer.Name, customer.Industry,
		customer.Status, time.Now(), customer.Id).Scan(&customer.UpdatedAt)

	return customer, err
}

func (r *RepositoryCustomerImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM "+models.CustomerTable+" WHERE id = $1", id)

	return err
}

// CountBrands backs the "cannot delete a customer that still has brands" guard, so
// the operator gets a readable message instead of a foreign-key violation.
func (r *RepositoryCustomerImpl) CountBrands(ctx context.Context, tx *sql.Tx, customerId int) (int, error) {
	var total int
	err := tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM "+models.BrandTable+" WHERE customer_id = $1", customerId).Scan(&total)

	return total, err
}
