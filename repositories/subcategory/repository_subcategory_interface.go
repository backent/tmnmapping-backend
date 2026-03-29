package subcategory

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositorySubCategoryInterface interface {
	Create(ctx context.Context, tx *sql.Tx, subCategory models.SubCategory) (models.SubCategory, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.SubCategory, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.SubCategory, error)
	Update(ctx context.Context, tx *sql.Tx, subCategory models.SubCategory) (models.SubCategory, error)
	Delete(ctx context.Context, tx *sql.Tx, id int) error
	FindByName(ctx context.Context, tx *sql.Tx, name string) (models.SubCategory, error)
	FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.SubCategory, error)
}
