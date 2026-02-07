package savedpolygon

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositorySavedPolygonInterface interface {
	Create(ctx context.Context, tx *sql.Tx, polygon models.SavedPolygon, points []models.SavedPolygonPoint) (models.SavedPolygon, error)
	CreatePoint(ctx context.Context, tx *sql.Tx, point models.SavedPolygonPoint) (models.SavedPolygonPoint, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.SavedPolygon, error)
	CountAll(ctx context.Context, tx *sql.Tx) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.SavedPolygon, error)
	Update(ctx context.Context, tx *sql.Tx, polygon models.SavedPolygon, points []models.SavedPolygonPoint) (models.SavedPolygon, error)
	DeletePointsBySavedPolygonId(ctx context.Context, tx *sql.Tx, savedPolygonId int) error
	Delete(ctx context.Context, tx *sql.Tx, id int) error
}
