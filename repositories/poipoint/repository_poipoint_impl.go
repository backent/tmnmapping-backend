package poipoint

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryPOIPointImpl struct{}

func NewRepositoryPOIPointImpl() RepositoryPOIPointInterface {
	return &RepositoryPOIPointImpl{}
}

// allowed order columns and directions to avoid SQL injection
var allowedOrderBy = map[string]bool{"id": true, "poi_name": true, "created_at": true, "updated_at": true}
var allowedOrderDir = map[string]bool{"ASC": true, "DESC": true}

func safeOrder(orderBy, orderDirection string) (string, string) {
	if !allowedOrderBy[orderBy] {
		orderBy = "created_at"
	}
	if !allowedOrderDir[orderDirection] {
		orderDirection = "DESC"
	}
	return orderBy, orderDirection
}

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

// Create inserts a new standalone POI point
func (r *RepositoryPOIPointImpl) Create(ctx context.Context, tx *sql.Tx, point models.POIPoint) (models.POIPoint, error) {
	SQL := `INSERT INTO ` + models.POIPointTable + `
		(poi_name, address, latitude, longitude, location, category, sub_category, mother_brand, branch)
		VALUES ($1, $2, $3, $4,
		CASE WHEN $3::DOUBLE PRECISION IS NOT NULL AND $4::DOUBLE PRECISION IS NOT NULL AND ($3::DOUBLE PRECISION) != 0 AND ($4::DOUBLE PRECISION) != 0 THEN ST_SetSRID(ST_MakePoint($4::DOUBLE PRECISION, $3::DOUBLE PRECISION), 4326)::geography ELSE NULL END,
		$5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		nullIfEmpty(point.POIName),
		nullIfEmpty(point.Address),
		nullIfZeroFloat(point.Latitude),
		nullIfZeroFloat(point.Longitude),
		nullIfEmpty(point.Category),
		nullIfEmpty(point.SubCategory),
		nullIfEmpty(point.MotherBrand),
		nullIfEmpty(point.Branch),
	).Scan(&point.Id, &point.CreatedAt, &point.UpdatedAt)

	if err != nil {
		return models.POIPoint{}, err
	}

	point.POIs = []models.POIRef{}
	return point, nil
}

// FindAll retrieves POI points with pagination, ordering, and optional search by poi_name
func (r *RepositoryPOIPointImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.POIPoint, error) {
	orderBy, orderDirection = safeOrder(orderBy, orderDirection)

	var args []interface{}
	paramIdx := 1

	SQL := `SELECT id, poi_name, address, latitude, longitude, category, sub_category, mother_brand, branch, created_at, updated_at
		FROM ` + models.POIPointTable

	if search != "" {
		SQL += ` WHERE poi_name ILIKE '%' || $` + strconv.Itoa(paramIdx) + ` || '%'`
		args = append(args, search)
		paramIdx++
	}

	SQL += ` ORDER BY ` + orderBy + ` ` + orderDirection
	SQL += ` LIMIT $` + strconv.Itoa(paramIdx) + ` OFFSET $` + strconv.Itoa(paramIdx+1)
	args = append(args, take, skip)

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.POIPoint
	var ids []int
	for rows.Next() {
		var n models.NullAblePOIPoint
		if err := rows.Scan(&n.Id, &n.POIName, &n.Address, &n.Latitude, &n.Longitude, &n.Category, &n.SubCategory, &n.MotherBrand, &n.Branch, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		point := models.NullAblePOIPointToPOIPoint(n)
		ids = append(ids, point.Id)
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) > 0 {
		refsMap, err := r.FindPOIRefsByPointIds(ctx, tx, ids)
		if err != nil {
			return nil, err
		}
		for i := range points {
			if refs, ok := refsMap[points[i].Id]; ok {
				points[i].POIs = refs
			}
		}
	}

	return points, nil
}

// CountAll returns total count of POI points, with optional search
func (r *RepositoryPOIPointImpl) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	SQL := `SELECT COUNT(*) FROM ` + models.POIPointTable
	var args []interface{}
	if search != "" {
		SQL += ` WHERE poi_name ILIKE '%' || $1 || '%'`
		args = append(args, search)
	}
	var total int
	err := tx.QueryRowContext(ctx, SQL, args...).Scan(&total)
	return total, err
}

// FindById retrieves a POI point by ID with its POI refs
func (r *RepositoryPOIPointImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.POIPoint, error) {
	SQL := `SELECT id, poi_name, address, latitude, longitude, category, sub_category, mother_brand, branch, created_at, updated_at
		FROM ` + models.POIPointTable + ` WHERE id = $1`
	var n models.NullAblePOIPoint
	if err := tx.QueryRowContext(ctx, SQL, id).Scan(&n.Id, &n.POIName, &n.Address, &n.Latitude, &n.Longitude, &n.Category, &n.SubCategory, &n.MotherBrand, &n.Branch, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return models.POIPoint{}, err
	}
	point := models.NullAblePOIPointToPOIPoint(n)
	refs, err := r.FindPOIRefsByPointId(ctx, tx, point.Id)
	if err != nil {
		return models.POIPoint{}, err
	}
	point.POIs = refs
	return point, nil
}

// Update updates a POI point's fields
func (r *RepositoryPOIPointImpl) Update(ctx context.Context, tx *sql.Tx, point models.POIPoint) (models.POIPoint, error) {
	SQL := `UPDATE ` + models.POIPointTable + `
		SET poi_name = $1, address = $2, latitude = $3, longitude = $4,
		location = CASE WHEN $3::DOUBLE PRECISION IS NOT NULL AND $4::DOUBLE PRECISION IS NOT NULL AND ($3::DOUBLE PRECISION) != 0 AND ($4::DOUBLE PRECISION) != 0 THEN ST_SetSRID(ST_MakePoint($4::DOUBLE PRECISION, $3::DOUBLE PRECISION), 4326)::geography ELSE NULL END,
		category = $5, sub_category = $6, mother_brand = $7, branch = $8, updated_at = $9
		WHERE id = $10
		RETURNING updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		nullIfEmpty(point.POIName),
		nullIfEmpty(point.Address),
		nullIfZeroFloat(point.Latitude),
		nullIfZeroFloat(point.Longitude),
		nullIfEmpty(point.Category),
		nullIfEmpty(point.SubCategory),
		nullIfEmpty(point.MotherBrand),
		nullIfEmpty(point.Branch),
		time.Now(),
		point.Id,
	).Scan(&point.UpdatedAt)

	if err != nil {
		return models.POIPoint{}, err
	}

	refs, err := r.FindPOIRefsByPointId(ctx, tx, point.Id)
	if err != nil {
		return models.POIPoint{}, err
	}
	point.POIs = refs
	return point, nil
}

// Delete deletes a POI point (CASCADE removes junction rows)
func (r *RepositoryPOIPointImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := `DELETE FROM ` + models.POIPointTable + ` WHERE id = $1`
	_, err := tx.ExecContext(ctx, SQL, id)
	return err
}

// FindAllDropdown returns a lightweight list of all POI points (id + poi_name) for autocomplete
func (r *RepositoryPOIPointImpl) FindAllDropdown(ctx context.Context, tx *sql.Tx) ([]models.POIPoint, error) {
	SQL := `SELECT id, poi_name, address, latitude, longitude, category, sub_category, mother_brand, branch, created_at, updated_at
		FROM ` + models.POIPointTable + ` ORDER BY poi_name ASC`
	rows, err := tx.QueryContext(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.POIPoint
	for rows.Next() {
		var n models.NullAblePOIPoint
		if err := rows.Scan(&n.Id, &n.POIName, &n.Address, &n.Latitude, &n.Longitude, &n.Category, &n.SubCategory, &n.MotherBrand, &n.Branch, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		point := models.NullAblePOIPointToPOIPoint(n)
		points = append(points, point)
	}
	return points, rows.Err()
}

// FindAllFlat returns all POI points with POI refs, optionally filtered by search (no pagination)
func (r *RepositoryPOIPointImpl) FindAllFlat(ctx context.Context, tx *sql.Tx, search string) ([]models.POIPoint, error) {
	SQL := `SELECT id, poi_name, address, latitude, longitude, category, sub_category, mother_brand, branch, created_at, updated_at
		FROM ` + models.POIPointTable

	var args []interface{}
	if search != "" {
		SQL += ` WHERE poi_name ILIKE '%' || $1 || '%'`
		args = append(args, search)
	}
	SQL += ` ORDER BY poi_name ASC`

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.POIPoint
	var ids []int
	for rows.Next() {
		var n models.NullAblePOIPoint
		if err := rows.Scan(&n.Id, &n.POIName, &n.Address, &n.Latitude, &n.Longitude, &n.Category, &n.SubCategory, &n.MotherBrand, &n.Branch, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		point := models.NullAblePOIPointToPOIPoint(n)
		ids = append(ids, point.Id)
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) > 0 {
		refsMap, err := r.FindPOIRefsByPointIds(ctx, tx, ids)
		if err != nil {
			return nil, err
		}
		for i := range points {
			if refs, ok := refsMap[points[i].Id]; ok {
				points[i].POIs = refs
			}
		}
	}

	return points, nil
}

// FindPOIRefsByPointId returns the POIs linked to a specific point
func (r *RepositoryPOIPointImpl) FindPOIRefsByPointId(ctx context.Context, tx *sql.Tx, pointId int) ([]models.POIRef, error) {
	SQL := `SELECT p.id, p.brand FROM ` + models.POITable + ` p
		INNER JOIN ` + models.POIPointPOITable + ` pp ON pp.poi_id = p.id
		WHERE pp.poi_point_id = $1 ORDER BY p.brand`
	rows, err := tx.QueryContext(ctx, SQL, pointId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refs []models.POIRef
	for rows.Next() {
		var ref models.POIRef
		if err := rows.Scan(&ref.Id, &ref.Brand); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	if refs == nil {
		refs = []models.POIRef{}
	}
	return refs, rows.Err()
}

// FindPOIRefsByPointIds returns map[pointId][]POIRef for batch loading
func (r *RepositoryPOIPointImpl) FindPOIRefsByPointIds(ctx context.Context, tx *sql.Tx, pointIds []int) (map[int][]models.POIRef, error) {
	if len(pointIds) == 0 {
		return make(map[int][]models.POIRef), nil
	}
	placeholders := make([]string, len(pointIds))
	args := make([]interface{}, len(pointIds))
	for i, id := range pointIds {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	SQL := `SELECT pp.poi_point_id, p.id, p.brand FROM ` + models.POITable + ` p
		INNER JOIN ` + models.POIPointPOITable + ` pp ON pp.poi_id = p.id
		WHERE pp.poi_point_id IN (` + strings.Join(placeholders, ",") + `) ORDER BY pp.poi_point_id, p.brand`
	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int][]models.POIRef)
	for rows.Next() {
		var pointId int
		var ref models.POIRef
		if err := rows.Scan(&pointId, &ref.Id, &ref.Brand); err != nil {
			return nil, err
		}
		out[pointId] = append(out[pointId], ref)
	}
	return out, rows.Err()
}

// FindByNameAndAddress returns a POI point matching the given name and address
func (r *RepositoryPOIPointImpl) FindByNameAndAddress(ctx context.Context, tx *sql.Tx, poiName string, address string) (models.POIPoint, error) {
	SQL := `SELECT id, poi_name, address, latitude, longitude, category, sub_category, mother_brand, branch, created_at, updated_at
		FROM ` + models.POIPointTable + ` WHERE poi_name = $1 AND address = $2 LIMIT 1`
	var n models.NullAblePOIPoint
	if err := tx.QueryRowContext(ctx, SQL, poiName, address).Scan(&n.Id, &n.POIName, &n.Address, &n.Latitude, &n.Longitude, &n.Category, &n.SubCategory, &n.MotherBrand, &n.Branch, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return models.POIPoint{}, err
	}
	point := models.NullAblePOIPointToPOIPoint(n)
	point.POIs = []models.POIRef{}
	return point, nil
}

// FindByPOINames returns POI points whose poi_name matches any in the given list
func (r *RepositoryPOIPointImpl) FindByPOINames(ctx context.Context, tx *sql.Tx, poiNames []string) ([]models.POIPoint, error) {
	if len(poiNames) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(poiNames))
	args := make([]interface{}, len(poiNames))
	for i, name := range poiNames {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = name
	}
	SQL := `SELECT id, poi_name, address, latitude, longitude, category, sub_category, mother_brand, branch, created_at, updated_at
		FROM ` + models.POIPointTable + ` WHERE poi_name IN (` + strings.Join(placeholders, ",") + `)`
	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.POIPoint
	for rows.Next() {
		var n models.NullAblePOIPoint
		if err := rows.Scan(&n.Id, &n.POIName, &n.Address, &n.Latitude, &n.Longitude, &n.Category, &n.SubCategory, &n.MotherBrand, &n.Branch, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		points = append(points, models.NullAblePOIPointToPOIPoint(n))
	}
	return points, rows.Err()
}
