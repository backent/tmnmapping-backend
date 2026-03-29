package poi

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryPOIImpl struct{}

func NewRepositoryPOIImpl() RepositoryPOIInterface {
	return &RepositoryPOIImpl{}
}

// Helper functions
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Create inserts a new POI and links it to existing points via junction table
func (repository *RepositoryPOIImpl) Create(ctx context.Context, tx *sql.Tx, poi models.POI, pointIds []int) (models.POI, error) {
	SQL := `INSERT INTO ` + models.POITable + ` (brand, color)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		poi.Brand,
		nullIfEmpty(poi.Color),
	).Scan(&poi.Id, &poi.CreatedAt, &poi.UpdatedAt)

	if err != nil {
		return models.POI{}, err
	}

	for _, pointId := range pointIds {
		if err := repository.CreatePointLink(ctx, tx, poi.Id, pointId); err != nil {
			return models.POI{}, err
		}
	}

	// Load points for response
	points, err := repository.findPointsByPOIId(ctx, tx, poi.Id)
	if err != nil {
		return models.POI{}, err
	}
	poi.Points = points

	return poi, nil
}

// CreatePointLink inserts a junction row linking a POI to a point
func (repository *RepositoryPOIImpl) CreatePointLink(ctx context.Context, tx *sql.Tx, poiId int, pointId int) error {
	SQL := `INSERT INTO ` + models.POIPointPOITable + ` (poi_id, poi_point_id) VALUES ($1, $2)`
	_, err := tx.ExecContext(ctx, SQL, poiId, pointId)
	return err
}

// DeletePointLinksByPOIId removes all point links for a POI
func (repository *RepositoryPOIImpl) DeletePointLinksByPOIId(ctx context.Context, tx *sql.Tx, poiId int) error {
	SQL := `DELETE FROM ` + models.POIPointPOITable + ` WHERE poi_id = $1`
	_, err := tx.ExecContext(ctx, SQL, poiId)
	return err
}

// FindAll retrieves all POIs with their points, with pagination, ordering, and optional search
func (repository *RepositoryPOIImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.POI, error) {
	var args []interface{}
	paramIdx := 1

	SQL := `SELECT id, brand, color, created_at, updated_at
		FROM ` + models.POITable

	if search != "" {
		SQL += ` WHERE brand ILIKE '%' || $` + strconv.Itoa(paramIdx) + ` || '%'`
		args = append(args, search)
		paramIdx++
	}

	SQL += ` ORDER BY ` + orderBy + ` ` + orderDirection
	SQL += ` LIMIT $` + strconv.Itoa(paramIdx) + ` OFFSET $` + strconv.Itoa(paramIdx+1)
	args = append(args, take, skip)

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return []models.POI{}, err
	}
	defer rows.Close()

	var pois []models.POI
	var poiIds []int
	for rows.Next() {
		nullable := models.NullAblePOI{}
		err := rows.Scan(
			&nullable.Id,
			&nullable.Brand,
			&nullable.Color,
			&nullable.CreatedAt,
			&nullable.UpdatedAt,
		)
		if err != nil {
			return []models.POI{}, err
		}

		poi := models.NullAblePOIToPOI(nullable)
		poiIds = append(poiIds, poi.Id)
		pois = append(pois, poi)
	}

	if err := rows.Err(); err != nil {
		return []models.POI{}, err
	}

	// Load all points for all POIs in one query
	if len(poiIds) > 0 {
		pointsMap, err := repository.findPointsByPOIIds(ctx, tx, poiIds)
		if err != nil {
			return []models.POI{}, err
		}

		for i := range pois {
			if points, exists := pointsMap[pois[i].Id]; exists {
				pois[i].Points = points
			} else {
				pois[i].Points = []models.POIPoint{}
			}
		}
	}

	return pois, nil
}

// CountAll returns the total count of POIs, with optional search filter
func (repository *RepositoryPOIImpl) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	SQL := `SELECT COUNT(*) FROM ` + models.POITable

	var args []interface{}
	if search != "" {
		SQL += ` WHERE brand ILIKE '%' || $1 || '%'`
		args = append(args, search)
	}

	var total int
	err := tx.QueryRowContext(ctx, SQL, args...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindAllFlat retrieves all POIs with their points, no pagination, with optional search
func (repository *RepositoryPOIImpl) FindAllFlat(ctx context.Context, tx *sql.Tx, search string) ([]models.POI, error) {
	SQL := `SELECT id, brand, color, created_at, updated_at
		FROM ` + models.POITable

	var args []interface{}
	if search != "" {
		SQL += ` WHERE brand ILIKE '%' || $1 || '%'`
		args = append(args, search)
	}

	SQL += ` ORDER BY brand ASC`

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return []models.POI{}, err
	}
	defer rows.Close()

	var pois []models.POI
	var poiIds []int
	for rows.Next() {
		nullable := models.NullAblePOI{}
		err := rows.Scan(
			&nullable.Id,
			&nullable.Brand,
			&nullable.Color,
			&nullable.CreatedAt,
			&nullable.UpdatedAt,
		)
		if err != nil {
			return []models.POI{}, err
		}

		poi := models.NullAblePOIToPOI(nullable)
		poiIds = append(poiIds, poi.Id)
		pois = append(pois, poi)
	}

	if err := rows.Err(); err != nil {
		return []models.POI{}, err
	}

	if len(poiIds) > 0 {
		pointsMap, err := repository.findPointsByPOIIds(ctx, tx, poiIds)
		if err != nil {
			return []models.POI{}, err
		}

		for i := range pois {
			if points, exists := pointsMap[pois[i].Id]; exists {
				pois[i].Points = points
			} else {
				pois[i].Points = []models.POIPoint{}
			}
		}
	}

	return pois, nil
}

// FindById retrieves a POI by ID with its points
func (repository *RepositoryPOIImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.POI, error) {
	SQL := `SELECT id, brand, color, created_at, updated_at
		FROM ` + models.POITable + `
		WHERE id = $1`

	row := tx.QueryRowContext(ctx, SQL, id)

	nullable := models.NullAblePOI{}
	err := row.Scan(
		&nullable.Id,
		&nullable.Brand,
		&nullable.Color,
		&nullable.CreatedAt,
		&nullable.UpdatedAt,
	)

	if err != nil {
		return models.POI{}, err
	}

	poi := models.NullAblePOIToPOI(nullable)

	// Load points via junction table
	points, err := repository.findPointsByPOIId(ctx, tx, poi.Id)
	if err != nil {
		return models.POI{}, err
	}
	poi.Points = points

	return poi, nil
}

// findPointsByPOIId loads points for a POI via the junction table
func (repository *RepositoryPOIImpl) findPointsByPOIId(ctx context.Context, tx *sql.Tx, poiId int) ([]models.POIPoint, error) {
	SQL := `SELECT pp.id, pp.poi_name, pp.address, pp.latitude, pp.longitude, pp.category, pp.sub_category, pp.mother_brand, pp.branch, pp.created_at, pp.updated_at
		FROM ` + models.POIPointTable + ` pp
		INNER JOIN ` + models.POIPointPOITable + ` j ON j.poi_point_id = pp.id
		WHERE j.poi_id = $1
		ORDER BY pp.created_at ASC`

	rows, err := tx.QueryContext(ctx, SQL, poiId)
	if err != nil {
		return []models.POIPoint{}, err
	}
	defer rows.Close()

	var points []models.POIPoint
	for rows.Next() {
		nullable := models.NullAblePOIPoint{}
		err := rows.Scan(
			&nullable.Id,
			&nullable.POIName,
			&nullable.Address,
			&nullable.Latitude,
			&nullable.Longitude,
			&nullable.Category,
			&nullable.SubCategory,
			&nullable.MotherBrand,
			&nullable.Branch,
			&nullable.CreatedAt,
			&nullable.UpdatedAt,
		)
		if err != nil {
			return []models.POIPoint{}, err
		}

		points = append(points, models.NullAblePOIPointToPOIPoint(nullable))
	}

	if err := rows.Err(); err != nil {
		return []models.POIPoint{}, err
	}

	if points == nil {
		points = []models.POIPoint{}
	}

	return points, nil
}

// findPointsByPOIIds loads points for multiple POIs via the junction table
func (repository *RepositoryPOIImpl) findPointsByPOIIds(ctx context.Context, tx *sql.Tx, poiIds []int) (map[int][]models.POIPoint, error) {
	if len(poiIds) == 0 {
		return make(map[int][]models.POIPoint), nil
	}

	placeholders := make([]string, len(poiIds))
	args := make([]interface{}, len(poiIds))
	for i, id := range poiIds {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	SQL := `SELECT j.poi_id, pp.id, pp.poi_name, pp.address, pp.latitude, pp.longitude, pp.category, pp.sub_category, pp.mother_brand, pp.branch, pp.created_at, pp.updated_at
		FROM ` + models.POIPointTable + ` pp
		INNER JOIN ` + models.POIPointPOITable + ` j ON j.poi_point_id = pp.id
		WHERE j.poi_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY j.poi_id, pp.created_at ASC`

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pointsMap := make(map[int][]models.POIPoint)
	for rows.Next() {
		var poiId int
		nullable := models.NullAblePOIPoint{}
		err := rows.Scan(
			&poiId,
			&nullable.Id,
			&nullable.POIName,
			&nullable.Address,
			&nullable.Latitude,
			&nullable.Longitude,
			&nullable.Category,
			&nullable.SubCategory,
			&nullable.MotherBrand,
			&nullable.Branch,
			&nullable.CreatedAt,
			&nullable.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		point := models.NullAblePOIPointToPOIPoint(nullable)
		pointsMap[poiId] = append(pointsMap[poiId], point)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pointsMap, nil
}

// Update updates a POI and replaces its point links
func (repository *RepositoryPOIImpl) Update(ctx context.Context, tx *sql.Tx, poi models.POI, pointIds []int) (models.POI, error) {
	SQL := `UPDATE ` + models.POITable + `
		SET brand = $1, color = $2, updated_at = $3
		WHERE id = $4
		RETURNING updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		poi.Brand,
		nullIfEmpty(poi.Color),
		time.Now(),
		poi.Id,
	).Scan(&poi.UpdatedAt)

	if err != nil {
		return models.POI{}, err
	}

	// Delete all existing point links
	if err := repository.DeletePointLinksByPOIId(ctx, tx, poi.Id); err != nil {
		return models.POI{}, err
	}

	// Create new point links
	for _, pointId := range pointIds {
		if err := repository.CreatePointLink(ctx, tx, poi.Id, pointId); err != nil {
			return models.POI{}, err
		}
	}

	// Load points for response
	points, err := repository.findPointsByPOIId(ctx, tx, poi.Id)
	if err != nil {
		return models.POI{}, err
	}
	poi.Points = points

	return poi, nil
}

// Delete deletes a POI (CASCADE on junction table removes links, but NOT the points themselves)
func (repository *RepositoryPOIImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := `DELETE FROM ` + models.POITable + ` WHERE id = $1`
	_, err := tx.ExecContext(ctx, SQL, id)
	return err
}

// FindByBrands returns existing POIs whose brand matches any of the given brand names
func (repository *RepositoryPOIImpl) FindByBrands(ctx context.Context, tx *sql.Tx, brands []string) ([]models.POI, error) {
	if len(brands) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(brands))
	args := make([]interface{}, len(brands))
	for i, b := range brands {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = b
	}

	SQL := `SELECT id, brand, color, created_at, updated_at FROM ` + models.POITable +
		` WHERE brand IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pois []models.POI
	for rows.Next() {
		var poi models.POI
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&poi.Id, &poi.Brand, &poi.Color, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		pois = append(pois, poi)
	}
	return pois, rows.Err()
}
