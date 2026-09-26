package brand

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBrandInterface interface {
	Create(ctx context.Context, tx *sql.Tx, brand models.Brand) (models.Brand, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string, customerId int) ([]models.Brand, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string, customerId int) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.Brand, error)
	FindByCode(ctx context.Context, tx *sql.Tx, code string) (models.Brand, error)
	Update(ctx context.Context, tx *sql.Tx, brand models.Brand) (models.Brand, error)
	Delete(ctx context.Context, tx *sql.Tx, id int) error
}
