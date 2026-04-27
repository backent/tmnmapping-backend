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

// ----- helpers -----

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

func nullIntPtr(p *int) interface{} {
	if p == nil {
		return nil
	}
	return *p
}

const poiSelectCols = `p.id, p.brand, p.color,
	p.category_id, p.sub_category_id, p.mother_brand_id,
	c.name, sc.name, mb.name,
	p.created_at, p.updated_at`

const poiJoins = ` LEFT JOIN categories c ON c.id = p.category_id
	LEFT JOIN sub_categories sc ON sc.id = p.sub_category_id
	LEFT JOIN mother_brands mb ON mb.id = p.mother_brand_id`

func scanPOI(scanner interface {
	Scan(dest ...any) error
}) (models.NullAblePOI, error) {
	var n models.NullAblePOI
	err := scanner.Scan(
		&n.Id, &n.Brand, &n.Color,
		&n.CategoryId, &n.SubCategoryId, &n.MotherBrandId,
		&n.CategoryName, &n.SubCategoryName, &n.MotherBrandName,
		&n.CreatedAt, &n.UpdatedAt,
	)
	return n, err
}

// ----- writes -----

// Create inserts a POI plus its owned points in one transaction.
func (repository *RepositoryPOIImpl) Create(ctx context.Context, tx *sql.Tx, poi models.POI, points []models.POIPoint) (models.POI, error) {
	SQL := `INSERT INTO ` + models.POITable + `
		(brand, color, category_id, sub_category_id, mother_brand_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		poi.Brand,
		nullIfEmpty(poi.Color),
		nullIntPtr(poi.CategoryId),
		nullIntPtr(poi.SubCategoryId),
		nullIntPtr(poi.MotherBrandId),
	).Scan(&poi.Id, &poi.CreatedAt, &poi.UpdatedAt)
	if err != nil {
		return models.POI{}, err
	}

	saved, err := repository.replacePoints(ctx, tx, poi.Id, points)
	if err != nil {
		return models.POI{}, err
	}
	poi.Points = saved

	// Re-fetch to fill in metadata names from the FK joins.
	loaded, err := repository.FindById(ctx, tx, poi.Id)
	if err != nil {
		return models.POI{}, err
	}
	loaded.Points = saved
	return loaded, nil
}

// Update updates POI fields and replaces the owned points.
func (repository *RepositoryPOIImpl) Update(ctx context.Context, tx *sql.Tx, poi models.POI, points []models.POIPoint) (models.POI, error) {
	SQL := `UPDATE ` + models.POITable + `
		SET brand = $1, color = $2,
		    category_id = $3, sub_category_id = $4, mother_brand_id = $5,
		    updated_at = $6
		WHERE id = $7
		RETURNING updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		poi.Brand,
		nullIfEmpty(poi.Color),
		nullIntPtr(poi.CategoryId),
		nullIntPtr(poi.SubCategoryId),
		nullIntPtr(poi.MotherBrandId),
		time.Now(),
		poi.Id,
	).Scan(&poi.UpdatedAt)
	if err != nil {
		return models.POI{}, err
	}

	saved, err := repository.replacePoints(ctx, tx, poi.Id, points)
	if err != nil {
		return models.POI{}, err
	}
	poi.Points = saved

	loaded, err := repository.FindById(ctx, tx, poi.Id)
	if err != nil {
		return models.POI{}, err
	}
	loaded.Points = saved
	return loaded, nil
}

// replacePoints deletes all poi_points for a POI and reinserts the given list.
func (repository *RepositoryPOIImpl) replacePoints(ctx context.Context, tx *sql.Tx, poiId int, points []models.POIPoint) ([]models.POIPoint, error) {
	if _, err := tx.ExecContext(ctx, `DELETE FROM `+models.POIPointTable+` WHERE poi_id = $1`, poiId); err != nil {
		return nil, err
	}

	saved := make([]models.POIPoint, 0, len(points))
	for _, pt := range points {
		inserted, err := repository.insertPoint(ctx, tx, poiId, pt)
		if err != nil {
			return nil, err
		}
		saved = append(saved, inserted)
	}
	return saved, nil
}

func (repository *RepositoryPOIImpl) insertPoint(ctx context.Context, tx *sql.Tx, poiId int, pt models.POIPoint) (models.POIPoint, error) {
	SQL := `INSERT INTO ` + models.POIPointTable + `
		(poi_id, poi_name, address, latitude, longitude, location, branch_id)
		VALUES ($1, $2, $3, $4, $5,
		CASE WHEN $4::DOUBLE PRECISION IS NOT NULL AND $5::DOUBLE PRECISION IS NOT NULL AND ($4::DOUBLE PRECISION) != 0 AND ($5::DOUBLE PRECISION) != 0
		     THEN ST_SetSRID(ST_MakePoint($5::DOUBLE PRECISION, $4::DOUBLE PRECISION), 4326)::geography
		     ELSE NULL END,
		$6)
		RETURNING id, created_at, updated_at`

	pt.POIId = poiId
	err := tx.QueryRowContext(ctx, SQL,
		poiId,
		nullIfEmpty(pt.POIName),
		nullIfEmpty(pt.Address),
		nullIfZeroFloat(pt.Latitude),
		nullIfZeroFloat(pt.Longitude),
		nullIntPtr(pt.BranchId),
	).Scan(&pt.Id, &pt.CreatedAt, &pt.UpdatedAt)
	if err != nil {
		return models.POIPoint{}, err
	}

	if pt.BranchId != nil {
		if err := tx.QueryRowContext(ctx,
			`SELECT name FROM branches WHERE id = $1`, *pt.BranchId,
		).Scan(&pt.BranchName); err != nil && err != sql.ErrNoRows {
			return models.POIPoint{}, err
		}
	}

	return pt, nil
}

// Delete cascades to poi_points via the FK on poi_points.poi_id.
func (repository *RepositoryPOIImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM `+models.POITable+` WHERE id = $1`, id)
	return err
}

// ----- reads -----

func (repository *RepositoryPOIImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.POI, error) {
	var args []interface{}
	paramIdx := 1

	SQL := `SELECT ` + poiSelectCols + ` FROM ` + models.POITable + ` p` + poiJoins
	if search != "" {
		SQL += ` WHERE p.brand ILIKE '%' || $` + strconv.Itoa(paramIdx) + ` || '%'`
		args = append(args, search)
		paramIdx++
	}
	SQL += ` ORDER BY p.` + orderBy + ` ` + orderDirection + `, p.brand ASC`
	SQL += ` LIMIT $` + strconv.Itoa(paramIdx) + ` OFFSET $` + strconv.Itoa(paramIdx+1)
	args = append(args, take, skip)

	return repository.queryPOIs(ctx, tx, SQL, args)
}

func (repository *RepositoryPOIImpl) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	SQL := `SELECT COUNT(*) FROM ` + models.POITable
	var args []interface{}
	if search != "" {
		SQL += ` WHERE brand ILIKE '%' || $1 || '%'`
		args = append(args, search)
	}
	var total int
	err := tx.QueryRowContext(ctx, SQL, args...).Scan(&total)
	return total, err
}

func (repository *RepositoryPOIImpl) FindAllFlat(ctx context.Context, tx *sql.Tx, search string) ([]models.POI, error) {
	SQL := `SELECT ` + poiSelectCols + ` FROM ` + models.POITable + ` p` + poiJoins
	var args []interface{}
	if search != "" {
		SQL += ` WHERE p.brand ILIKE '%' || $1 || '%'`
		args = append(args, search)
	}
	SQL += ` ORDER BY p.brand ASC`
	return repository.queryPOIs(ctx, tx, SQL, args)
}

func (repository *RepositoryPOIImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.POI, error) {
	SQL := `SELECT ` + poiSelectCols + ` FROM ` + models.POITable + ` p` + poiJoins + ` WHERE p.id = $1`
	n, err := scanPOI(tx.QueryRowContext(ctx, SQL, id))
	if err != nil {
		return models.POI{}, err
	}
	poi := models.NullAblePOIToPOI(n)
	points, err := repository.findPointsByPOIId(ctx, tx, poi.Id)
	if err != nil {
		return models.POI{}, err
	}
	poi.Points = points
	return poi, nil
}

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

	SQL := `SELECT ` + poiSelectCols + ` FROM ` + models.POITable + ` p` + poiJoins +
		` WHERE p.brand IN (` + strings.Join(placeholders, ",") + `)`

	return repository.queryPOIs(ctx, tx, SQL, args)
}

// queryPOIs runs a SELECT that returns POI rows + metadata names, then loads points by poi_id.
func (repository *RepositoryPOIImpl) queryPOIs(ctx context.Context, tx *sql.Tx, SQL string, args []interface{}) ([]models.POI, error) {
	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return []models.POI{}, err
	}
	defer rows.Close()

	var pois []models.POI
	var poiIds []int
	for rows.Next() {
		n, err := scanPOI(rows)
		if err != nil {
			return []models.POI{}, err
		}
		poi := models.NullAblePOIToPOI(n)
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
			if pts, ok := pointsMap[pois[i].Id]; ok {
				pois[i].Points = pts
			} else {
				pois[i].Points = []models.POIPoint{}
			}
		}
	}
	return pois, nil
}

const poiPointSelectCols = `pp.id, pp.poi_id, pp.poi_name, pp.address, pp.latitude, pp.longitude,
	pp.branch_id, b.name,
	pp.created_at, pp.updated_at`

const poiPointJoins = ` LEFT JOIN branches b ON b.id = pp.branch_id`

func scanPOIPoint(scanner interface {
	Scan(dest ...any) error
}) (models.NullAblePOIPoint, error) {
	var n models.NullAblePOIPoint
	err := scanner.Scan(
		&n.Id, &n.POIId, &n.POIName, &n.Address, &n.Latitude, &n.Longitude,
		&n.BranchId, &n.BranchName,
		&n.CreatedAt, &n.UpdatedAt,
	)
	return n, err
}

func (repository *RepositoryPOIImpl) findPointsByPOIId(ctx context.Context, tx *sql.Tx, poiId int) ([]models.POIPoint, error) {
	SQL := `SELECT ` + poiPointSelectCols + ` FROM ` + models.POIPointTable + ` pp` + poiPointJoins +
		` WHERE pp.poi_id = $1 ORDER BY pp.created_at ASC`
	rows, err := tx.QueryContext(ctx, SQL, poiId)
	if err != nil {
		return []models.POIPoint{}, err
	}
	defer rows.Close()

	var points []models.POIPoint
	for rows.Next() {
		n, err := scanPOIPoint(rows)
		if err != nil {
			return []models.POIPoint{}, err
		}
		points = append(points, models.NullAblePOIPointToPOIPoint(n))
	}
	if err := rows.Err(); err != nil {
		return []models.POIPoint{}, err
	}
	if points == nil {
		points = []models.POIPoint{}
	}
	return points, nil
}

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

	SQL := `SELECT ` + poiPointSelectCols + ` FROM ` + models.POIPointTable + ` pp` + poiPointJoins +
		` WHERE pp.poi_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY pp.poi_id, pp.created_at ASC`

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pointsMap := make(map[int][]models.POIPoint)
	for rows.Next() {
		n, err := scanPOIPoint(rows)
		if err != nil {
			return nil, err
		}
		point := models.NullAblePOIPointToPOIPoint(n)
		pointsMap[point.POIId] = append(pointsMap[point.POIId], point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pointsMap, nil
}
