package salespackage

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositorySalesPackageInterface interface {
	Create(ctx context.Context, tx *sql.Tx, pkg models.SalesPackage, buildingIds []int) (models.SalesPackage, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.SalesPackage, error)
	CountAll(ctx context.Context, tx *sql.Tx) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.SalesPackage, error)
	Update(ctx context.Context, tx *sql.Tx, pkg models.SalesPackage, buildingIds []int) (models.SalesPackage, error)
	DeleteBuildingLinksBySalesPackageId(ctx context.Context, tx *sql.Tx, salesPackageId int) error
	CreateBuildingLink(ctx context.Context, tx *sql.Tx, salesPackageId int, buildingId int) error
	Delete(ctx context.Context, tx *sql.Tx, id int) error
}
