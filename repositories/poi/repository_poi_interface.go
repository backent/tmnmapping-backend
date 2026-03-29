package poi

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryPOIInterface interface {
	Create(ctx context.Context, tx *sql.Tx, poi models.POI, pointIds []int) (models.POI, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.POI, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error)
	FindAllFlat(ctx context.Context, tx *sql.Tx, search string) ([]models.POI, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.POI, error)
	FindByBrands(ctx context.Context, tx *sql.Tx, brands []string) ([]models.POI, error)
	Update(ctx context.Context, tx *sql.Tx, poi models.POI, pointIds []int) (models.POI, error)
	CreatePointLink(ctx context.Context, tx *sql.Tx, poiId int, pointId int) error
	DeletePointLinksByPOIId(ctx context.Context, tx *sql.Tx, poiId int) error
	Delete(ctx context.Context, tx *sql.Tx, id int) error
}
