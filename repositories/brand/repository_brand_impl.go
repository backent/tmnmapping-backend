package brand

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBrandImpl struct{}

func NewRepositoryBrandImpl() RepositoryBrandInterface {
	return &RepositoryBrandImpl{}
}

// Brands are always read with their customer joined: every screen that lists a brand
// shows which advertiser it belongs to.
var brandSelect = `SELECT b.id, b.code, b.customer_id, c.name, c.code, b.name, b.category,
	b.status, b.attention_to, b.job_title, b.contact_phone, b.contact_email,
	b.created_at, b.updated_at FROM ` + models.BrandTable + ` b JOIN ` +
	models.CustomerTable + ` c ON c.id = b.customer_id`

var brandAllowedOrderBy = map[string]string{
	"id": "b.id", "code": "b.code", "name": "b.name", "category": "b.category",
	"status": "b.status", "created_at": "b.created_at", "updated_at": "b.updated_at",
	"customer_name": "c.name",
}

func brandSafeOrder(orderBy, orderDirection string) (string, string) {
	column, ok := brandAllowedOrderBy[orderBy]
	if !ok {
		column = "b.created_at"
	}
	if orderDirection != "ASC" && orderDirection != "DESC" {
		orderDirection = "DESC"
	}

	return column, orderDirection
}

func scanBrand(rows *sql.Rows) (models.Brand, error) {
	var n models.NullAbleBrand
	err := rows.Scan(&n.Id, &n.Code, &n.CustomerId, &n.CustomerName, &n.CustomerCode,
		&n.Name, &n.Category, &n.Status, &n.AttentionTo, &n.JobTitle,
		&n.ContactPhone, &n.ContactEmail, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.Brand{}, err
	}

	return models.NullAbleBrandToBrand(n), nil
}

// brandFilter builds the shared WHERE clause for list and count.
func brandFilter(search string, customerId int) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	if search != "" {
		args = append(args, "%"+search+"%")
		clauses = append(clauses, "(b.code ILIKE $1 OR b.name ILIKE $1 OR b.category ILIKE $1 OR c.name ILIKE $1)")
	}
	if customerId > 0 {
		args = append(args, customerId)
		if len(args) == 1 {
			clauses = append(clauses, "b.customer_id = $1")
		} else {
			clauses = append(clauses, "b.customer_id = $2")
		}
	}

	if len(clauses) == 0 {
		return "", args
	}

	where := " WHERE " + clauses[0]
	for _, clause := range clauses[1:] {
		where += " AND " + clause
	}

	return where, args
}

func (r *RepositoryBrandImpl) Create(ctx context.Context, tx *sql.Tx, brand models.Brand) (models.Brand, error) {
	SQL := "INSERT INTO " + models.BrandTable +
		" (code, customer_id, name, category, status, attention_to, job_title," +
		" contact_phone, contact_email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)" +
		" RETURNING id, created_at, updated_at"
	err := tx.QueryRowContext(ctx, SQL, brand.Code, brand.CustomerId, brand.Name,
		brand.Category, brand.Status, brand.AttentionTo, brand.JobTitle,
		brand.ContactPhone, brand.ContactEmail).Scan(&brand.Id, &brand.CreatedAt, &brand.UpdatedAt)

	return brand, err
}

func (r *RepositoryBrandImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string, customerId int) ([]models.Brand, error) {
	column, direction := brandSafeOrder(orderBy, orderDirection)
	where, args := brandFilter(search, customerId)

	SQL := brandSelect + where + " ORDER BY " + column + " " + direction + ", b.name ASC" +
		" LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, take, skip)

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Brand
	for rows.Next() {
		item, err := scanBrand(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

func (r *RepositoryBrandImpl) CountAll(ctx context.Context, tx *sql.Tx, search string, customerId int) (int, error) {
	where, args := brandFilter(search, customerId)

	var total int
	SQL := "SELECT COUNT(*) FROM " + models.BrandTable + " b JOIN " + models.CustomerTable +
		" c ON c.id = b.customer_id" + where
	err := tx.QueryRowContext(ctx, SQL, args...).Scan(&total)

	return total, err
}

func (r *RepositoryBrandImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Brand, error) {
	return r.findOne(ctx, tx, " WHERE b.id = $1", id)
}

func (r *RepositoryBrandImpl) FindByCode(ctx context.Context, tx *sql.Tx, code string) (models.Brand, error) {
	return r.findOne(ctx, tx, " WHERE b.code = $1", code)
}

func (r *RepositoryBrandImpl) findOne(ctx context.Context, tx *sql.Tx, where string, arg interface{}) (models.Brand, error) {
	rows, err := tx.QueryContext(ctx, brandSelect+where, arg)
	if err != nil {
		return models.Brand{}, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanBrand(rows)
	}

	return models.Brand{}, sql.ErrNoRows
}

func (r *RepositoryBrandImpl) Update(ctx context.Context, tx *sql.Tx, brand models.Brand) (models.Brand, error) {
	SQL := "UPDATE " + models.BrandTable +
		" SET code = $1, customer_id = $2, name = $3, category = $4, status = $5," +
		" attention_to = $6, job_title = $7, contact_phone = $8, contact_email = $9," +
		" updated_at = $10 WHERE id = $11 RETURNING updated_at"
	err := tx.QueryRowContext(ctx, SQL, brand.Code, brand.CustomerId, brand.Name,
		brand.Category, brand.Status, brand.AttentionTo, brand.JobTitle,
		brand.ContactPhone, brand.ContactEmail, time.Now(), brand.Id).Scan(&brand.UpdatedAt)

	return brand, err
}

func (r *RepositoryBrandImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM "+models.BrandTable+" WHERE id = $1", id)

	return err
}
