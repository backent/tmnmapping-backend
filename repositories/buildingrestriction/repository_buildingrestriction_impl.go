package buildingrestriction

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBuildingRestrictionImpl struct{}

func NewRepositoryBuildingRestrictionImpl() RepositoryBuildingRestrictionInterface {
	return &RepositoryBuildingRestrictionImpl{}
}

// allowed order columns and directions to avoid SQL injection
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

// Create inserts a new building restriction and its building links
func (r *RepositoryBuildingRestrictionImpl) Create(ctx context.Context, tx *sql.Tx, restriction models.BuildingRestriction, buildingIds []int) (models.BuildingRestriction, error) {
	SQL := `INSERT INTO ` + models.BuildingRestrictionTable + ` (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err := tx.QueryRowContext(ctx, SQL, restriction.Name).Scan(&restriction.Id, &restriction.CreatedAt, &restriction.UpdatedAt)
	if err != nil {
		return models.BuildingRestriction{}, err
	}
	for _, bid := range buildingIds {
		if err := r.CreateBuildingLink(ctx, tx, restriction.Id, bid); err != nil {
			return models.BuildingRestriction{}, err
		}
	}
	// Load building refs for response
	buildings, err := r.findBuildingRefsByBuildingRestrictionId(ctx, tx, restriction.Id)
	if err != nil {
		return models.BuildingRestriction{}, err
	}
	restriction.Buildings = buildings
	return restriction, nil
}

// FindAll retrieves all building restrictions with pagination and ordering; loads building refs per restriction
func (r *RepositoryBuildingRestrictionImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.BuildingRestriction, error) {
	orderBy, orderDirection = safeOrder(orderBy, orderDirection)
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.BuildingRestrictionTable + `
		ORDER BY ` + orderBy + ` ` + orderDirection + ` LIMIT $1 OFFSET $2`
	rows, err := tx.QueryContext(ctx, SQL, take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var restrictions []models.BuildingRestriction
	var ids []int
	for rows.Next() {
		var n models.NullAbleBuildingRestriction
		if err := rows.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		restriction := models.NullAbleBuildingRestrictionToBuildingRestriction(n)
		ids = append(ids, restriction.Id)
		restrictions = append(restrictions, restriction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return restrictions, nil
	}
	refsMap, err := r.findBuildingRefsByBuildingRestrictionIds(ctx, tx, ids)
	if err != nil {
		return nil, err
	}
	for i := range restrictions {
		restrictions[i].Buildings = refsMap[restrictions[i].Id]
	}
	return restrictions, nil
}

// CountAll returns total count of building restrictions
func (r *RepositoryBuildingRestrictionImpl) CountAll(ctx context.Context, tx *sql.Tx) (int, error) {
	SQL := `SELECT COUNT(*) FROM ` + models.BuildingRestrictionTable
	var total int
	err := tx.QueryRowContext(ctx, SQL).Scan(&total)
	return total, err
}

// FindById retrieves a building restriction by ID with its building refs
func (r *RepositoryBuildingRestrictionImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.BuildingRestriction, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.BuildingRestrictionTable + ` WHERE id = $1`
	row := tx.QueryRowContext(ctx, SQL, id)
	var n models.NullAbleBuildingRestriction
	if err := row.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return models.BuildingRestriction{}, err
	}
	restriction := models.NullAbleBuildingRestrictionToBuildingRestriction(n)
	buildings, err := r.findBuildingRefsByBuildingRestrictionId(ctx, tx, restriction.Id)
	if err != nil {
		return models.BuildingRestriction{}, err
	}
	restriction.Buildings = buildings
	return restriction, nil
}

// Update updates name and replaces building links
func (r *RepositoryBuildingRestrictionImpl) Update(ctx context.Context, tx *sql.Tx, restriction models.BuildingRestriction, buildingIds []int) (models.BuildingRestriction, error) {
	SQL := `UPDATE ` + models.BuildingRestrictionTable + ` SET name = $1, updated_at = $2 WHERE id = $3 RETURNING updated_at`
	err := tx.QueryRowContext(ctx, SQL, restriction.Name, time.Now(), restriction.Id).Scan(&restriction.UpdatedAt)
	if err != nil {
		return models.BuildingRestriction{}, err
	}
	if err := r.DeleteBuildingLinksByBuildingRestrictionId(ctx, tx, restriction.Id); err != nil {
		return models.BuildingRestriction{}, err
	}
	for _, bid := range buildingIds {
		if err := r.CreateBuildingLink(ctx, tx, restriction.Id, bid); err != nil {
			return models.BuildingRestriction{}, err
		}
	}
	buildings, err := r.findBuildingRefsByBuildingRestrictionId(ctx, tx, restriction.Id)
	if err != nil {
		return models.BuildingRestriction{}, err
	}
	restriction.Buildings = buildings
	return restriction, nil
}

// DeleteBuildingLinksByBuildingRestrictionId removes all building links for a restriction
func (r *RepositoryBuildingRestrictionImpl) DeleteBuildingLinksByBuildingRestrictionId(ctx context.Context, tx *sql.Tx, buildingRestrictionId int) error {
	SQL := `DELETE FROM ` + models.BuildingRestrictionBuildingTable + ` WHERE building_restriction_id = $1`
	_, err := tx.ExecContext(ctx, SQL, buildingRestrictionId)
	return err
}

// CreateBuildingLink inserts one building_restriction_id, building_id row
func (r *RepositoryBuildingRestrictionImpl) CreateBuildingLink(ctx context.Context, tx *sql.Tx, buildingRestrictionId int, buildingId int) error {
	SQL := `INSERT INTO ` + models.BuildingRestrictionBuildingTable + ` (building_restriction_id, building_id) VALUES ($1, $2)`
	_, err := tx.ExecContext(ctx, SQL, buildingRestrictionId, buildingId)
	return err
}

// Delete deletes a building restriction (CASCADE removes junction rows)
func (r *RepositoryBuildingRestrictionImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := `DELETE FROM ` + models.BuildingRestrictionTable + ` WHERE id = $1`
	_, err := tx.ExecContext(ctx, SQL, id)
	return err
}

// findBuildingRefsByBuildingRestrictionId returns building id+name for a restriction
func (r *RepositoryBuildingRestrictionImpl) findBuildingRefsByBuildingRestrictionId(ctx context.Context, tx *sql.Tx, buildingRestrictionId int) ([]models.BuildingRef, error) {
	SQL := `SELECT b.id, b.name FROM ` + models.BuildingTable + ` b
		INNER JOIN ` + models.BuildingRestrictionBuildingTable + ` brb ON brb.building_id = b.id
		WHERE brb.building_restriction_id = $1 ORDER BY b.name`
	rows, err := tx.QueryContext(ctx, SQL, buildingRestrictionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refs []models.BuildingRef
	for rows.Next() {
		var ref models.BuildingRef
		if err := rows.Scan(&ref.Id, &ref.Name); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

// findBuildingRefsByBuildingRestrictionIds returns map[buildingRestrictionId][]BuildingRef
func (r *RepositoryBuildingRestrictionImpl) findBuildingRefsByBuildingRestrictionIds(ctx context.Context, tx *sql.Tx, buildingRestrictionIds []int) (map[int][]models.BuildingRef, error) {
	if len(buildingRestrictionIds) == 0 {
		return make(map[int][]models.BuildingRef), nil
	}
	placeholders := make([]string, len(buildingRestrictionIds))
	args := make([]interface{}, len(buildingRestrictionIds))
	for i, id := range buildingRestrictionIds {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	SQL := `SELECT brb.building_restriction_id, b.id, b.name FROM ` + models.BuildingTable + ` b
		INNER JOIN ` + models.BuildingRestrictionBuildingTable + ` brb ON brb.building_id = b.id
		WHERE brb.building_restriction_id IN (` + strings.Join(placeholders, ",") + `) ORDER BY brb.building_restriction_id, b.name`
	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int][]models.BuildingRef)
	for rows.Next() {
		var restrictionId int
		var ref models.BuildingRef
		if err := rows.Scan(&restrictionId, &ref.Id, &ref.Name); err != nil {
			return nil, err
		}
		out[restrictionId] = append(out[restrictionId], ref)
	}
	return out, rows.Err()
}
