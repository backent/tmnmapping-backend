package motherbrand

import (
	"context"
	"database/sql"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryMotherBrandImpl struct{}

func NewRepositoryMotherBrandImpl() RepositoryMotherBrandInterface {
	return &RepositoryMotherBrandImpl{}
}

var mdMBAllowedOrderBy = map[string]bool{"id": true, "name": true, "created_at": true, "updated_at": true}
var mdMBAllowedOrderDir = map[string]bool{"ASC": true, "DESC": true}

func mdMBSafeOrder(orderBy, orderDirection string) (string, string) {
	if !mdMBAllowedOrderBy[orderBy] {
		orderBy = "created_at"
	}
	if !mdMBAllowedOrderDir[orderDirection] {
		orderDirection = "DESC"
	}
	return orderBy, orderDirection
}

func (r *RepositoryMotherBrandImpl) Create(ctx context.Context, tx *sql.Tx, motherBrand models.MotherBrand) (models.MotherBrand, error) {
	SQL := `INSERT INTO ` + models.MotherBrandTable + ` (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err := tx.QueryRowContext(ctx, SQL, motherBrand.Name).Scan(&motherBrand.Id, &motherBrand.CreatedAt, &motherBrand.UpdatedAt)
	return motherBrand, err
}

func (r *RepositoryMotherBrandImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.MotherBrand, error) {
	orderBy, orderDirection = mdMBSafeOrder(orderBy, orderDirection)
	var rows *sql.Rows
	var err error
	if search != "" {
		SQL := `SELECT id, name, created_at, updated_at FROM ` + models.MotherBrandTable + ` WHERE name ILIKE $1 ORDER BY ` + orderBy + ` ` + orderDirection + `, name ASC LIMIT $2 OFFSET $3`
		rows, err = tx.QueryContext(ctx, SQL, "%"+search+"%", take, skip)
	} else {
		SQL := `SELECT id, name, created_at, updated_at FROM ` + models.MotherBrandTable + ` ORDER BY ` + orderBy + ` ` + orderDirection + `, name ASC LIMIT $1 OFFSET $2`
		rows, err = tx.QueryContext(ctx, SQL, take, skip)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.MotherBrand
	for rows.Next() {
		var n models.NullAbleMotherBrand
		if err := rows.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, models.NullAbleMotherBrandToMotherBrand(n))
	}
	return list, rows.Err()
}

func (r *RepositoryMotherBrandImpl) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	var total int
	var err error
	if search != "" {
		SQL := `SELECT COUNT(*) FROM ` + models.MotherBrandTable + ` WHERE name ILIKE $1`
		err = tx.QueryRowContext(ctx, SQL, "%"+search+"%").Scan(&total)
	} else {
		SQL := `SELECT COUNT(*) FROM ` + models.MotherBrandTable
		err = tx.QueryRowContext(ctx, SQL).Scan(&total)
	}
	return total, err
}

func (r *RepositoryMotherBrandImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.MotherBrand, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.MotherBrandTable + ` WHERE id = $1`
	var n models.NullAbleMotherBrand
	err := tx.QueryRowContext(ctx, SQL, id).Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.MotherBrand{}, err
	}
	return models.NullAbleMotherBrandToMotherBrand(n), nil
}

func (r *RepositoryMotherBrandImpl) Update(ctx context.Context, tx *sql.Tx, motherBrand models.MotherBrand) (models.MotherBrand, error) {
	SQL := `UPDATE ` + models.MotherBrandTable + ` SET name = $1, updated_at = $2 WHERE id = $3 RETURNING updated_at`
	err := tx.QueryRowContext(ctx, SQL, motherBrand.Name, time.Now(), motherBrand.Id).Scan(&motherBrand.UpdatedAt)
	return motherBrand, err
}

func (r *RepositoryMotherBrandImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := `DELETE FROM ` + models.MotherBrandTable + ` WHERE id = $1`
	_, err := tx.ExecContext(ctx, SQL, id)
	return err
}

func (r *RepositoryMotherBrandImpl) FindByName(ctx context.Context, tx *sql.Tx, name string) (models.MotherBrand, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.MotherBrandTable + ` WHERE LOWER(name) = LOWER($1) LIMIT 1`
	var n models.NullAbleMotherBrand
	err := tx.QueryRowContext(ctx, SQL, name).Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.MotherBrand{}, err
	}
	return models.NullAbleMotherBrandToMotherBrand(n), nil
}

func (r *RepositoryMotherBrandImpl) FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.MotherBrand, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.MotherBrandTable + ` ORDER BY name ASC`
	rows, err := tx.QueryContext(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.MotherBrand
	for rows.Next() {
		var n models.NullAbleMotherBrand
		if err := rows.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, models.NullAbleMotherBrandToMotherBrand(n))
	}
	return list, rows.Err()
}
