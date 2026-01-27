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

func nullIfZeroFloat(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}

// Create inserts a new POI
func (repository *RepositoryPOIImpl) Create(ctx context.Context, tx *sql.Tx, poi models.POI) (models.POI, error) {
	SQL := `INSERT INTO ` + models.POITable + ` (name, color) 
		VALUES ($1, $2) 
		RETURNING id, created_at, updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		poi.Name,
		nullIfEmpty(poi.Color),
	).Scan(&poi.Id, &poi.CreatedAt, &poi.UpdatedAt)

	if err != nil {
		return models.POI{}, err
	}

	// Create all points
	for i := range poi.Points {
		poi.Points[i].POIId = poi.Id
		point, err := repository.CreatePoint(ctx, tx, poi.Points[i])
		if err != nil {
			return models.POI{}, err
		}
		poi.Points[i] = point
	}

	return poi, nil
}

// CreatePoint inserts a new POI point
func (repository *RepositoryPOIImpl) CreatePoint(ctx context.Context, tx *sql.Tx, point models.POIPoint) (models.POIPoint, error) {
	SQL := `INSERT INTO ` + models.POIPointTable + ` 
		(poi_id, place_name, address, latitude, longitude, location) 
		VALUES ($1, $2, $3, $4, $5, 
		CASE WHEN $4::DOUBLE PRECISION IS NOT NULL AND $5::DOUBLE PRECISION IS NOT NULL AND ($4::DOUBLE PRECISION) != 0 AND ($5::DOUBLE PRECISION) != 0 THEN ST_SetSRID(ST_MakePoint($5::DOUBLE PRECISION, $4::DOUBLE PRECISION), 4326)::geography ELSE NULL END) 
		RETURNING id, created_at`

	err := tx.QueryRowContext(ctx, SQL,
		point.POIId,
		nullIfEmpty(point.PlaceName),
		nullIfEmpty(point.Address),
		nullIfZeroFloat(point.Latitude),
		nullIfZeroFloat(point.Longitude),
	).Scan(&point.Id, &point.CreatedAt)

	if err != nil {
		return models.POIPoint{}, err
	}

	return point, nil
}

// FindAll retrieves all POIs with their points, with pagination and ordering
func (repository *RepositoryPOIImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.POI, error) {
	SQL := `SELECT id, name, color, created_at, updated_at 
		FROM ` + models.POITable + ` 
		ORDER BY ` + orderBy + ` ` + orderDirection + ` 
		LIMIT $1 OFFSET $2`

	rows, err := tx.QueryContext(ctx, SQL, take, skip)
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
			&nullable.Name,
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

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return []models.POI{}, err
	}

	// Load all points for all POIs in one query
	if len(poiIds) > 0 {
		pointsMap, err := repository.findPointsByPOIIds(ctx, tx, poiIds)
		if err != nil {
			return []models.POI{}, err
		}

		// Assign points to their respective POIs
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

// CountAll returns the total count of POIs
func (repository *RepositoryPOIImpl) CountAll(ctx context.Context, tx *sql.Tx) (int, error) {
	SQL := `SELECT COUNT(*) FROM ` + models.POITable

	var total int
	err := tx.QueryRowContext(ctx, SQL).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindById retrieves a POI by ID with its points
func (repository *RepositoryPOIImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.POI, error) {
	SQL := `SELECT id, name, color, created_at, updated_at 
		FROM ` + models.POITable + ` 
		WHERE id = $1`

	row := tx.QueryRowContext(ctx, SQL, id)

	nullable := models.NullAblePOI{}
	err := row.Scan(
		&nullable.Id,
		&nullable.Name,
		&nullable.Color,
		&nullable.CreatedAt,
		&nullable.UpdatedAt,
	)

	if err != nil {
		return models.POI{}, err
	}

	poi := models.NullAblePOIToPOI(nullable)

	// Load points
	points, err := repository.findPointsByPOIId(ctx, tx, poi.Id)
	if err != nil {
		return models.POI{}, err
	}
	poi.Points = points

	return poi, nil
}

// findPointsByPOIId is a helper to load points for a POI
func (repository *RepositoryPOIImpl) findPointsByPOIId(ctx context.Context, tx *sql.Tx, poiId int) ([]models.POIPoint, error) {
	SQL := `SELECT id, poi_id, place_name, address, latitude, longitude, created_at 
		FROM ` + models.POIPointTable + ` 
		WHERE poi_id = $1 
		ORDER BY created_at ASC`

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
			&nullable.POIId,
			&nullable.PlaceName,
			&nullable.Address,
			&nullable.Latitude,
			&nullable.Longitude,
			&nullable.CreatedAt,
		)
		if err != nil {
			return []models.POIPoint{}, err
		}

		points = append(points, models.NullAblePOIPointToPOIPoint(nullable))
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return []models.POIPoint{}, err
	}

	return points, nil
}

// findPointsByPOIIds loads points for multiple POIs in one query
func (repository *RepositoryPOIImpl) findPointsByPOIIds(ctx context.Context, tx *sql.Tx, poiIds []int) (map[int][]models.POIPoint, error) {
	if len(poiIds) == 0 {
		return make(map[int][]models.POIPoint), nil
	}

	// Build IN clause with placeholders
	placeholders := make([]string, len(poiIds))
	args := make([]interface{}, len(poiIds))
	for i, id := range poiIds {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	SQL := `SELECT id, poi_id, place_name, address, latitude, longitude, created_at 
		FROM ` + models.POIPointTable + ` 
		WHERE poi_id IN (` + strings.Join(placeholders, ",") + `) 
		ORDER BY poi_id, created_at ASC`

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pointsMap := make(map[int][]models.POIPoint)
	for rows.Next() {
		nullable := models.NullAblePOIPoint{}
		err := rows.Scan(
			&nullable.Id,
			&nullable.POIId,
			&nullable.PlaceName,
			&nullable.Address,
			&nullable.Latitude,
			&nullable.Longitude,
			&nullable.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		point := models.NullAblePOIPointToPOIPoint(nullable)
		pointsMap[int(nullable.POIId.Int64)] = append(pointsMap[int(nullable.POIId.Int64)], point)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pointsMap, nil
}

// Update updates a POI and replaces all its points
func (repository *RepositoryPOIImpl) Update(ctx context.Context, tx *sql.Tx, poi models.POI) (models.POI, error) {
	// Update POI
	SQL := `UPDATE ` + models.POITable + ` 
		SET name = $1, color = $2, updated_at = $3 
		WHERE id = $4 
		RETURNING updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		poi.Name,
		nullIfEmpty(poi.Color),
		time.Now(),
		poi.Id,
	).Scan(&poi.UpdatedAt)

	if err != nil {
		return models.POI{}, err
	}

	// Delete all existing points
	err = repository.DeletePointsByPOIId(ctx, tx, poi.Id)
	if err != nil {
		return models.POI{}, err
	}

	// Create new points
	newPoints := make([]models.POIPoint, len(poi.Points))
	for i := range poi.Points {
		poi.Points[i].POIId = poi.Id
		point, err := repository.CreatePoint(ctx, tx, poi.Points[i])
		if err != nil {
			return models.POI{}, err
		}
		newPoints[i] = point
	}
	poi.Points = newPoints

	return poi, nil
}

// DeletePointsByPOIId deletes all points for a POI
func (repository *RepositoryPOIImpl) DeletePointsByPOIId(ctx context.Context, tx *sql.Tx, poiId int) error {
	SQL := `DELETE FROM ` + models.POIPointTable + ` WHERE poi_id = $1`
	_, err := tx.ExecContext(ctx, SQL, poiId)
	return err
}

// Delete deletes a POI (cascade will delete points)
func (repository *RepositoryPOIImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := `DELETE FROM ` + models.POITable + ` WHERE id = $1`
	_, err := tx.ExecContext(ctx, SQL, id)
	return err
}
