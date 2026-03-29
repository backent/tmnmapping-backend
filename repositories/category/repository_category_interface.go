package category

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryCategoryInterface interface {
	Create(ctx context.Context, tx *sql.Tx, category models.Category) (models.Category, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.Category, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.Category, error)
	Update(ctx context.Context, tx *sql.Tx, category models.Category) (models.Category, error)
	Delete(ctx context.Context, tx *sql.Tx, id int) error
	FindByName(ctx context.Context, tx *sql.Tx, name string) (models.Category, error)
	FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.Category, error)
}
