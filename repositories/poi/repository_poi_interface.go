package poi

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryPOIInterface interface {
	Create(ctx context.Context, tx *sql.Tx, poi models.POI) (models.POI, error)
	CreatePoint(ctx context.Context, tx *sql.Tx, point models.POIPoint) (models.POIPoint, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.POI, error)
	CountAll(ctx context.Context, tx *sql.Tx) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.POI, error)
	Update(ctx context.Context, tx *sql.Tx, poi models.POI) (models.POI, error)
	DeletePointsByPOIId(ctx context.Context, tx *sql.Tx, poiId int) error
	Delete(ctx context.Context, tx *sql.Tx, id int) error
}
