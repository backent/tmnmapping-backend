package buildingprice

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBuildingPriceImpl struct{}

func NewRepositoryBuildingPriceImpl() RepositoryBuildingPriceInterface {
	return &RepositoryBuildingPriceImpl{}
}

var buildingPriceSelect = `SELECT p.id, p.building_id, b.name, b.iris_code, b.building_type, b.citytown,
	p.price_idr_per_week, p.created_at, p.updated_at
	FROM ` + models.BuildingPriceTable + ` p
	JOIN ` + models.BuildingTable + ` b ON b.id = p.building_id`

const buildingPriceSearch = " (b.name ILIKE $1 OR b.iris_code ILIKE $1 OR b.citytown ILIKE $1)"

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanBuildingPrice(row rowScanner) (models.BuildingPrice, error) {
	var n models.NullAbleBuildingPrice
	err := row.Scan(&n.Id, &n.BuildingId, &n.BuildingName, &n.BuildingIrisCode,
		&n.BuildingType, &n.Citytown, &n.PriceIdrPerWeek, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.BuildingPrice{}, err
	}

	return models.NullAbleBuildingPriceToBuildingPrice(n), nil
}

func (r *RepositoryBuildingPriceImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, search string) ([]models.BuildingPrice, error) {
	var rows *sql.Rows
	var err error

	if search != "" {
		rows, err = tx.QueryContext(ctx, buildingPriceSelect+" WHERE"+buildingPriceSearch+
			" ORDER BY b.name ASC LIMIT $2 OFFSET $3", "%"+search+"%", take, skip)
	} else {
		rows, err = tx.QueryContext(ctx, buildingPriceSelect+
			" ORDER BY b.name ASC LIMIT $1 OFFSET $2", take, skip)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.BuildingPrice{}
	for rows.Next() {
		item, err := scanBuildingPrice(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

func (r *RepositoryBuildingPriceImpl) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	base := "SELECT COUNT(*) FROM " + models.BuildingPriceTable + " p JOIN " +
		models.BuildingTable + " b ON b.id = p.building_id"

	var total int
	var err error
	if search != "" {
		err = tx.QueryRowContext(ctx, base+" WHERE"+buildingPriceSearch, "%"+search+"%").Scan(&total)
	} else {
		err = tx.QueryRowContext(ctx, base).Scan(&total)
	}

	return total, err
}

func (r *RepositoryBuildingPriceImpl) FindByBuildingId(ctx context.Context, tx *sql.Tx, buildingId int) (models.BuildingPrice, error) {
	return scanBuildingPrice(tx.QueryRowContext(ctx, buildingPriceSelect+" WHERE p.building_id = $1", buildingId))
}

func (r *RepositoryBuildingPriceImpl) FindAllPrices(ctx context.Context, tx *sql.Tx) (map[int]int64, error) {
	rows, err := tx.QueryContext(ctx, "SELECT building_id, price_idr_per_week FROM "+models.BuildingPriceTable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prices := map[int]int64{}
	for rows.Next() {
		var buildingId int
		var price sql.NullInt64
		if err := rows.Scan(&buildingId, &price); err != nil {
			return nil, err
		}
		prices[buildingId] = price.Int64
	}

	return prices, rows.Err()
}

// Upsert lets an upload and a hand edit share one path: pricing a building that
// already has a price replaces it rather than failing on the unique key.
func (r *RepositoryBuildingPriceImpl) Upsert(ctx context.Context, tx *sql.Tx, buildingId int, price int64) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO "+models.BuildingPriceTable+
		" (building_id, price_idr_per_week) VALUES ($1, $2)"+
		" ON CONFLICT (building_id) DO UPDATE SET price_idr_per_week = EXCLUDED.price_idr_per_week,"+
		" updated_at = CURRENT_TIMESTAMP", buildingId, price)

	return err
}

func (r *RepositoryBuildingPriceImpl) Delete(ctx context.Context, tx *sql.Tx, buildingId int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM "+models.BuildingPriceTable+" WHERE building_id = $1", buildingId)

	return err
}
