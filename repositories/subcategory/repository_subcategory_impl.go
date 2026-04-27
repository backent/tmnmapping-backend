package subcategory

import (
	"context"
	"database/sql"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositorySubCategoryImpl struct{}

func NewRepositorySubCategoryImpl() RepositorySubCategoryInterface {
	return &RepositorySubCategoryImpl{}
}

var mdSubCatAllowedOrderBy = map[string]bool{"id": true, "name": true, "created_at": true, "updated_at": true}
var mdSubCatAllowedOrderDir = map[string]bool{"ASC": true, "DESC": true}

func mdSubCatSafeOrder(orderBy, orderDirection string) (string, string) {
	if !mdSubCatAllowedOrderBy[orderBy] {
		orderBy = "created_at"
	}
	if !mdSubCatAllowedOrderDir[orderDirection] {
		orderDirection = "DESC"
	}
	return orderBy, orderDirection
}

func (r *RepositorySubCategoryImpl) Create(ctx context.Context, tx *sql.Tx, subCategory models.SubCategory) (models.SubCategory, error) {
	SQL := `INSERT INTO ` + models.SubCategoryTable + ` (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err := tx.QueryRowContext(ctx, SQL, subCategory.Name).Scan(&subCategory.Id, &subCategory.CreatedAt, &subCategory.UpdatedAt)
	return subCategory, err
}

func (r *RepositorySubCategoryImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.SubCategory, error) {
	orderBy, orderDirection = mdSubCatSafeOrder(orderBy, orderDirection)
	var rows *sql.Rows
	var err error
	if search != "" {
		SQL := `SELECT id, name, created_at, updated_at FROM ` + models.SubCategoryTable + ` WHERE name ILIKE $1 ORDER BY ` + orderBy + ` ` + orderDirection + `, name ASC LIMIT $2 OFFSET $3`
		rows, err = tx.QueryContext(ctx, SQL, "%"+search+"%", take, skip)
	} else {
		SQL := `SELECT id, name, created_at, updated_at FROM ` + models.SubCategoryTable + ` ORDER BY ` + orderBy + ` ` + orderDirection + `, name ASC LIMIT $1 OFFSET $2`
		rows, err = tx.QueryContext(ctx, SQL, take, skip)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.SubCategory
	for rows.Next() {
		var n models.NullAbleSubCategory
		if err := rows.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, models.NullAbleSubCategoryToSubCategory(n))
	}
	return list, rows.Err()
}

func (r *RepositorySubCategoryImpl) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	var total int
	var err error
	if search != "" {
		SQL := `SELECT COUNT(*) FROM ` + models.SubCategoryTable + ` WHERE name ILIKE $1`
		err = tx.QueryRowContext(ctx, SQL, "%"+search+"%").Scan(&total)
	} else {
		SQL := `SELECT COUNT(*) FROM ` + models.SubCategoryTable
		err = tx.QueryRowContext(ctx, SQL).Scan(&total)
	}
	return total, err
}

func (r *RepositorySubCategoryImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.SubCategory, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.SubCategoryTable + ` WHERE id = $1`
	var n models.NullAbleSubCategory
	err := tx.QueryRowContext(ctx, SQL, id).Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.SubCategory{}, err
	}
	return models.NullAbleSubCategoryToSubCategory(n), nil
}

func (r *RepositorySubCategoryImpl) Update(ctx context.Context, tx *sql.Tx, subCategory models.SubCategory) (models.SubCategory, error) {
	SQL := `UPDATE ` + models.SubCategoryTable + ` SET name = $1, updated_at = $2 WHERE id = $3 RETURNING updated_at`
	err := tx.QueryRowContext(ctx, SQL, subCategory.Name, time.Now(), subCategory.Id).Scan(&subCategory.UpdatedAt)
	return subCategory, err
}

func (r *RepositorySubCategoryImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := `DELETE FROM ` + models.SubCategoryTable + ` WHERE id = $1`
	_, err := tx.ExecContext(ctx, SQL, id)
	return err
}

func (r *RepositorySubCategoryImpl) FindByName(ctx context.Context, tx *sql.Tx, name string) (models.SubCategory, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.SubCategoryTable + ` WHERE LOWER(name) = LOWER($1) LIMIT 1`
	var n models.NullAbleSubCategory
	err := tx.QueryRowContext(ctx, SQL, name).Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.SubCategory{}, err
	}
	return models.NullAbleSubCategoryToSubCategory(n), nil
}

func (r *RepositorySubCategoryImpl) FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.SubCategory, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.SubCategoryTable + ` ORDER BY name ASC`
	rows, err := tx.QueryContext(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.SubCategory
	for rows.Next() {
		var n models.NullAbleSubCategory
		if err := rows.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, models.NullAbleSubCategoryToSubCategory(n))
	}
	return list, rows.Err()
}
