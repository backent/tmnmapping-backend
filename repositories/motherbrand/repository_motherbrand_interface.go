package motherbrand

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryMotherBrandInterface interface {
	Create(ctx context.Context, tx *sql.Tx, motherBrand models.MotherBrand) (models.MotherBrand, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.MotherBrand, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.MotherBrand, error)
	Update(ctx context.Context, tx *sql.Tx, motherBrand models.MotherBrand) (models.MotherBrand, error)
	Delete(ctx context.Context, tx *sql.Tx, id int) error
	FindByName(ctx context.Context, tx *sql.Tx, name string) (models.MotherBrand, error)
	FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.MotherBrand, error)
}
