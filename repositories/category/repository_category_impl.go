package category

import (
	"context"
	"database/sql"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryCategoryImpl struct{}

func NewRepositoryCategoryImpl() RepositoryCategoryInterface {
	return &RepositoryCategoryImpl{}
}

var mdCatAllowedOrderBy = map[string]bool{"id": true, "name": true, "created_at": true, "updated_at": true}
var mdCatAllowedOrderDir = map[string]bool{"ASC": true, "DESC": true}

func mdCatSafeOrder(orderBy, orderDirection string) (string, string) {
	if !mdCatAllowedOrderBy[orderBy] {
		orderBy = "created_at"
	}
	if !mdCatAllowedOrderDir[orderDirection] {
		orderDirection = "DESC"
	}
	return orderBy, orderDirection
}

func (r *RepositoryCategoryImpl) Create(ctx context.Context, tx *sql.Tx, category models.Category) (models.Category, error) {
	SQL := `INSERT INTO ` + models.CategoryTable + ` (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err := tx.QueryRowContext(ctx, SQL, category.Name).Scan(&category.Id, &category.CreatedAt, &category.UpdatedAt)
	return category, err
}

func (r *RepositoryCategoryImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.Category, error) {
	orderBy, orderDirection = mdCatSafeOrder(orderBy, orderDirection)
	var rows *sql.Rows
	var err error
	if search != "" {
		SQL := `SELECT id, name, created_at, updated_at FROM ` + models.CategoryTable + ` WHERE name ILIKE $1 ORDER BY ` + orderBy + ` ` + orderDirection + ` LIMIT $2 OFFSET $3`
		rows, err = tx.QueryContext(ctx, SQL, "%"+search+"%", take, skip)
	} else {
		SQL := `SELECT id, name, created_at, updated_at FROM ` + models.CategoryTable + ` ORDER BY ` + orderBy + ` ` + orderDirection + ` LIMIT $1 OFFSET $2`
		rows, err = tx.QueryContext(ctx, SQL, take, skip)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Category
	for rows.Next() {
		var n models.NullAbleCategory
		if err := rows.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, models.NullAbleCategoryToCategory(n))
	}
	return list, rows.Err()
}

func (r *RepositoryCategoryImpl) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	var total int
	var err error
	if search != "" {
		SQL := `SELECT COUNT(*) FROM ` + models.CategoryTable + ` WHERE name ILIKE $1`
		err = tx.QueryRowContext(ctx, SQL, "%"+search+"%").Scan(&total)
	} else {
		SQL := `SELECT COUNT(*) FROM ` + models.CategoryTable
		err = tx.QueryRowContext(ctx, SQL).Scan(&total)
	}
	return total, err
}

func (r *RepositoryCategoryImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Category, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.CategoryTable + ` WHERE id = $1`
	var n models.NullAbleCategory
	err := tx.QueryRowContext(ctx, SQL, id).Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.Category{}, err
	}
	return models.NullAbleCategoryToCategory(n), nil
}

func (r *RepositoryCategoryImpl) Update(ctx context.Context, tx *sql.Tx, category models.Category) (models.Category, error) {
	SQL := `UPDATE ` + models.CategoryTable + ` SET name = $1, updated_at = $2 WHERE id = $3 RETURNING updated_at`
	err := tx.QueryRowContext(ctx, SQL, category.Name, time.Now(), category.Id).Scan(&category.UpdatedAt)
	return category, err
}

func (r *RepositoryCategoryImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := `DELETE FROM ` + models.CategoryTable + ` WHERE id = $1`
	_, err := tx.ExecContext(ctx, SQL, id)
	return err
}

func (r *RepositoryCategoryImpl) FindByName(ctx context.Context, tx *sql.Tx, name string) (models.Category, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.CategoryTable + ` WHERE LOWER(name) = LOWER($1) LIMIT 1`
	var n models.NullAbleCategory
	err := tx.QueryRowContext(ctx, SQL, name).Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.Category{}, err
	}
	return models.NullAbleCategoryToCategory(n), nil
}

func (r *RepositoryCategoryImpl) FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.Category, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.CategoryTable + ` ORDER BY name ASC`
	rows, err := tx.QueryContext(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Category
	for rows.Next() {
		var n models.NullAbleCategory
		if err := rows.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, models.NullAbleCategoryToCategory(n))
	}
	return list, rows.Err()
}
