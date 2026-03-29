package branch

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBranchInterface interface {
	Create(ctx context.Context, tx *sql.Tx, branch models.Branch) (models.Branch, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.Branch, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.Branch, error)
	Update(ctx context.Context, tx *sql.Tx, branch models.Branch) (models.Branch, error)
	Delete(ctx context.Context, tx *sql.Tx, id int) error
	FindByName(ctx context.Context, tx *sql.Tx, name string) (models.Branch, error)
	FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.Branch, error)
}
