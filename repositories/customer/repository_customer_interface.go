package customer

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryCustomerInterface interface {
	Create(ctx context.Context, tx *sql.Tx, customer models.Customer) (models.Customer, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.Customer, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.Customer, error)
	FindByCode(ctx context.Context, tx *sql.Tx, code string) (models.Customer, error)
	Update(ctx context.Context, tx *sql.Tx, customer models.Customer) (models.Customer, error)
	Delete(ctx context.Context, tx *sql.Tx, id int) error
	CountBrands(ctx context.Context, tx *sql.Tx, customerId int) (int, error)
}
