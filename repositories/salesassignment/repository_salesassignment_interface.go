package salesassignment

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositorySalesAssignmentInterface interface {
	Create(ctx context.Context, tx *sql.Tx, assignment models.SalesAssignment) (models.SalesAssignment, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.SalesAssignment, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.SalesAssignment, error)
	FindByCustomerAndBrand(ctx context.Context, tx *sql.Tx, customerId int, brandId int) (models.SalesAssignment, error)
	Update(ctx context.Context, tx *sql.Tx, assignment models.SalesAssignment) (models.SalesAssignment, error)
	Delete(ctx context.Context, tx *sql.Tx, id int) error
	CountBySalesUser(ctx context.Context, tx *sql.Tx, salesUserId int) (int, error)
}
