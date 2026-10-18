package buildingprice

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBuildingPriceInterface interface {
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, search string) ([]models.BuildingPrice, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error)

	// FindByBuildingId returns sql.ErrNoRows when the building has no price.
	FindByBuildingId(ctx context.Context, tx *sql.Tx, buildingId int) (models.BuildingPrice, error)

	// FindPricesByBuildingIds maps building id to weekly price for the given ids in
	// one query. A building with no price is absent from the map.
	FindPricesByBuildingIds(ctx context.Context, tx *sql.Tx, ids []int) (map[int]int64, error)

	// FindAllPrices maps building id to weekly price, so an import can tell a new
	// price from a changed or an unchanged one before writing anything.
	FindAllPrices(ctx context.Context, tx *sql.Tx) (map[int]int64, error)

	Upsert(ctx context.Context, tx *sql.Tx, buildingId int, price int64) error
	Delete(ctx context.Context, tx *sql.Tx, buildingId int) error
}
