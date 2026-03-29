package poipoint

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryPOIPointInterface interface {
	Create(ctx context.Context, tx *sql.Tx, point models.POIPoint) (models.POIPoint, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.POIPoint, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.POIPoint, error)
	Update(ctx context.Context, tx *sql.Tx, point models.POIPoint) (models.POIPoint, error)
	Delete(ctx context.Context, tx *sql.Tx, id int) error
	FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.POIPoint, error)
	FindAllFlat(ctx context.Context, tx *sql.Tx, search string) ([]models.POIPoint, error)
	FindPOIRefsByPointId(ctx context.Context, tx *sql.Tx, pointId int) ([]models.POIRef, error)
	FindPOIRefsByPointIds(ctx context.Context, tx *sql.Tx, pointIds []int) (map[int][]models.POIRef, error)
	FindByPOINames(ctx context.Context, tx *sql.Tx, poiNames []string) ([]models.POIPoint, error)
	FindByNameAndAddress(ctx context.Context, tx *sql.Tx, poiName string, address string) (models.POIPoint, error)
}
