package savedpolygon

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositorySavedPolygonImpl struct{}

func NewRepositorySavedPolygonImpl() RepositorySavedPolygonInterface {
	return &RepositorySavedPolygonImpl{}
}

var allowedOrderBy = map[string]bool{"id": true, "name": true, "created_at": true, "updated_at": true}
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

// Create inserts a new saved polygon and its points
func (r *RepositorySavedPolygonImpl) Create(ctx context.Context, tx *sql.Tx, polygon models.SavedPolygon, points []models.SavedPolygonPoint) (models.SavedPolygon, error) {
	SQL := `INSERT INTO ` + models.SavedPolygonTable + ` (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err := tx.QueryRowContext(ctx, SQL, polygon.Name).Scan(&polygon.Id, &polygon.CreatedAt, &polygon.UpdatedAt)
	if err != nil {
		return models.SavedPolygon{}, err
	}
	for i := range points {
		points[i].SavedPolygonId = polygon.Id
		points[i].Ord = i
		pt, err := r.CreatePoint(ctx, tx, points[i])
		if err != nil {
			return models.SavedPolygon{}, err
		}
		polygon.Points = append(polygon.Points, pt)
	}
	return polygon, nil
}

// CreatePoint inserts a new saved polygon point
func (r *RepositorySavedPolygonImpl) CreatePoint(ctx context.Context, tx *sql.Tx, point models.SavedPolygonPoint) (models.SavedPolygonPoint, error) {
	SQL := `INSERT INTO ` + models.SavedPolygonPointTable + ` (saved_polygon_id, ord, lat, lng) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err := tx.QueryRowContext(ctx, SQL, point.SavedPolygonId, point.Ord, point.Lat, point.Lng).Scan(&point.Id, &point.CreatedAt)
	if err != nil {
		return models.SavedPolygonPoint{}, err
	}
	return point, nil
}

// FindAll retrieves all saved polygons with points, with pagination and ordering
func (r *RepositorySavedPolygonImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.SavedPolygon, error) {
	orderBy, orderDirection = safeOrder(orderBy, orderDirection)
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.SavedPolygonTable + `
		ORDER BY ` + orderBy + ` ` + orderDirection + ` LIMIT $1 OFFSET $2`
	rows, err := tx.QueryContext(ctx, SQL, take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var polygons []models.SavedPolygon
	var ids []int
	for rows.Next() {
		var n models.NullAbleSavedPolygon
		if err := rows.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		poly := models.NullAbleSavedPolygonToSavedPolygon(n)
		ids = append(ids, poly.Id)
		polygons = append(polygons, poly)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return polygons, nil
	}
	pointsMap, err := r.findPointsBySavedPolygonIds(ctx, tx, ids)
	if err != nil {
		return nil, err
	}
	for i := range polygons {
		polygons[i].Points = pointsMap[polygons[i].Id]
	}
	return polygons, nil
}

// CountAll returns total count of saved polygons
func (r *RepositorySavedPolygonImpl) CountAll(ctx context.Context, tx *sql.Tx) (int, error) {
	SQL := `SELECT COUNT(*) FROM ` + models.SavedPolygonTable
	var total int
	err := tx.QueryRowContext(ctx, SQL).Scan(&total)
	return total, err
}

// FindById retrieves a saved polygon by ID with its points
func (r *RepositorySavedPolygonImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.SavedPolygon, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.SavedPolygonTable + ` WHERE id = $1`
	row := tx.QueryRowContext(ctx, SQL, id)
	var n models.NullAbleSavedPolygon
	if err := row.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return models.SavedPolygon{}, err
	}
	poly := models.NullAbleSavedPolygonToSavedPolygon(n)
	points, err := r.findPointsBySavedPolygonId(ctx, tx, poly.Id)
	if err != nil {
		return models.SavedPolygon{}, err
	}
	poly.Points = points
	return poly, nil
}

func (r *RepositorySavedPolygonImpl) findPointsBySavedPolygonId(ctx context.Context, tx *sql.Tx, savedPolygonId int) ([]models.SavedPolygonPoint, error) {
	SQL := `SELECT id, saved_polygon_id, ord, lat, lng, created_at FROM ` + models.SavedPolygonPointTable + `
		WHERE saved_polygon_id = $1 ORDER BY ord ASC`
	rows, err := tx.QueryContext(ctx, SQL, savedPolygonId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []models.SavedPolygonPoint
	for rows.Next() {
		var n models.NullAbleSavedPolygonPoint
		if err := rows.Scan(&n.Id, &n.SavedPolygonId, &n.Ord, &n.Lat, &n.Lng, &n.CreatedAt); err != nil {
			return nil, err
		}
		points = append(points, models.NullAbleSavedPolygonPointToSavedPolygonPoint(n))
	}
	return points, rows.Err()
}

func (r *RepositorySavedPolygonImpl) findPointsBySavedPolygonIds(ctx context.Context, tx *sql.Tx, savedPolygonIds []int) (map[int][]models.SavedPolygonPoint, error) {
	if len(savedPolygonIds) == 0 {
		return make(map[int][]models.SavedPolygonPoint), nil
	}
	placeholders := make([]string, len(savedPolygonIds))
	args := make([]interface{}, len(savedPolygonIds))
	for i, id := range savedPolygonIds {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	SQL := `SELECT id, saved_polygon_id, ord, lat, lng, created_at FROM ` + models.SavedPolygonPointTable + `
		WHERE saved_polygon_id IN (` + strings.Join(placeholders, ",") + `) ORDER BY saved_polygon_id, ord ASC`
	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int][]models.SavedPolygonPoint)
	for rows.Next() {
		var n models.NullAbleSavedPolygonPoint
		if err := rows.Scan(&n.Id, &n.SavedPolygonId, &n.Ord, &n.Lat, &n.Lng, &n.CreatedAt); err != nil {
			return nil, err
		}
		pt := models.NullAbleSavedPolygonPointToSavedPolygonPoint(n)
		out[pt.SavedPolygonId] = append(out[pt.SavedPolygonId], pt)
	}
	return out, rows.Err()
}

// Update updates a saved polygon and replaces its points
func (r *RepositorySavedPolygonImpl) Update(ctx context.Context, tx *sql.Tx, polygon models.SavedPolygon, points []models.SavedPolygonPoint) (models.SavedPolygon, error) {
	SQL := `UPDATE ` + models.SavedPolygonTable + ` SET name = $1, updated_at = $2 WHERE id = $3 RETURNING updated_at`
	err := tx.QueryRowContext(ctx, SQL, polygon.Name, time.Now(), polygon.Id).Scan(&polygon.UpdatedAt)
	if err != nil {
		return models.SavedPolygon{}, err
	}
	if err := r.DeletePointsBySavedPolygonId(ctx, tx, polygon.Id); err != nil {
		return models.SavedPolygon{}, err
	}
	polygon.Points = nil
	for i := range points {
		points[i].SavedPolygonId = polygon.Id
		points[i].Ord = i
		pt, err := r.CreatePoint(ctx, tx, points[i])
		if err != nil {
			return models.SavedPolygon{}, err
		}
		polygon.Points = append(polygon.Points, pt)
	}
	return polygon, nil
}

// DeletePointsBySavedPolygonId deletes all points for a saved polygon
func (r *RepositorySavedPolygonImpl) DeletePointsBySavedPolygonId(ctx context.Context, tx *sql.Tx, savedPolygonId int) error {
	SQL := `DELETE FROM ` + models.SavedPolygonPointTable + ` WHERE saved_polygon_id = $1`
	_, err := tx.ExecContext(ctx, SQL, savedPolygonId)
	return err
}

// Delete deletes a saved polygon (CASCADE deletes points)
func (r *RepositorySavedPolygonImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := `DELETE FROM ` + models.SavedPolygonTable + ` WHERE id = $1`
	_, err := tx.ExecContext(ctx, SQL, id)
	return err
}
