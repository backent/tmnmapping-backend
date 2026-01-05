package building

import (
	"context"
	"database/sql"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBuildingImpl struct{}

func NewRepositoryBuildingImpl() RepositoryBuildingInterface {
	return &RepositoryBuildingImpl{}
}

// Create inserts a new building
func (repository *RepositoryBuildingImpl) Create(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error) {
	SQL := `INSERT INTO ` + models.BuildingTable + ` 
		(external_building_id, iris_code, name, project_name, audience, impression, 
		cbd_area, building_status, competitor_location, sellable, connectivity, 
		resource_type, synced_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
		RETURNING id, created_at, updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		building.ExternalBuildingId,
		building.IrisCode,
		building.Name,
		building.ProjectName,
		building.Audience,
		building.Impression,
		building.CbdArea,
		building.BuildingStatus,
		building.CompetitorLocation,
		building.Sellable,
		building.Connectivity,
		nullIfEmpty(building.ResourceType),
		nullIfEmpty(building.SyncedAt),
	).Scan(&building.Id, &building.CreatedAt, &building.UpdatedAt)

	if err != nil {
		return models.Building{}, err
	}

	return building, nil
}

// FindById retrieves a building by ID
func (repository *RepositoryBuildingImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Building, error) {
	SQL := `SELECT id, external_building_id, iris_code, name, project_name, audience, 
		impression, cbd_area, building_status, competitor_location, sellable, connectivity, 
		resource_type, synced_at, created_at, updated_at 
		FROM ` + models.BuildingTable + ` WHERE id = $1`

	rows, err := tx.QueryContext(ctx, SQL, id)
	if err != nil {
		return models.Building{}, err
	}
	defer rows.Close()

	building := models.NullAbleBuilding{}
	if rows.Next() {
		err := rows.Scan(
			&building.Id,
			&building.ExternalBuildingId,
			&building.IrisCode,
			&building.Name,
			&building.ProjectName,
			&building.Audience,
			&building.Impression,
			&building.CbdArea,
			&building.BuildingStatus,
			&building.CompetitorLocation,
			&building.Sellable,
			&building.Connectivity,
			&building.ResourceType,
			&building.SyncedAt,
			&building.CreatedAt,
			&building.UpdatedAt,
		)
		if err != nil {
			return models.Building{}, err
		}
		return models.NullAbleBuildingToBuilding(building), nil
	}

	return models.Building{}, sql.ErrNoRows
}

// FindByExternalId retrieves a building by external ERP ID
func (repository *RepositoryBuildingImpl) FindByExternalId(ctx context.Context, tx *sql.Tx, externalId string) (models.Building, error) {
	SQL := `SELECT id, external_building_id, iris_code, name, project_name, audience, 
		impression, cbd_area, building_status, competitor_location, sellable, connectivity, 
		resource_type, synced_at, created_at, updated_at 
		FROM ` + models.BuildingTable + ` WHERE external_building_id = $1`

	rows, err := tx.QueryContext(ctx, SQL, externalId)
	if err != nil {
		return models.Building{}, err
	}
	defer rows.Close()

	building := models.NullAbleBuilding{}
	if rows.Next() {
		err := rows.Scan(
			&building.Id,
			&building.ExternalBuildingId,
			&building.IrisCode,
			&building.Name,
			&building.ProjectName,
			&building.Audience,
			&building.Impression,
			&building.CbdArea,
			&building.BuildingStatus,
			&building.CompetitorLocation,
			&building.Sellable,
			&building.Connectivity,
			&building.ResourceType,
			&building.SyncedAt,
			&building.CreatedAt,
			&building.UpdatedAt,
		)
		if err != nil {
			return models.Building{}, err
		}
		return models.NullAbleBuildingToBuilding(building), nil
	}

	return models.Building{}, sql.ErrNoRows
}

// FindAll retrieves all buildings with pagination and sorting
func (repository *RepositoryBuildingImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.Building, error) {
	SQL := `SELECT id, external_building_id, iris_code, name, project_name, audience, 
		impression, cbd_area, building_status, competitor_location, sellable, connectivity, 
		resource_type, synced_at, created_at, updated_at 
		FROM ` + models.BuildingTable + ` ORDER BY ` + orderBy + ` ` + orderDirection + ` LIMIT $1 OFFSET $2`

	rows, err := tx.QueryContext(ctx, SQL, take, skip)
	if err != nil {
		return []models.Building{}, err
	}
	defer rows.Close()

	var buildings []models.Building
	for rows.Next() {
		building := models.NullAbleBuilding{}
		err := rows.Scan(
			&building.Id,
			&building.ExternalBuildingId,
			&building.IrisCode,
			&building.Name,
			&building.ProjectName,
			&building.Audience,
			&building.Impression,
			&building.CbdArea,
			&building.BuildingStatus,
			&building.CompetitorLocation,
			&building.Sellable,
			&building.Connectivity,
			&building.ResourceType,
			&building.SyncedAt,
			&building.CreatedAt,
			&building.UpdatedAt,
		)
		if err != nil {
			return []models.Building{}, err
		}
		buildings = append(buildings, models.NullAbleBuildingToBuilding(building))
	}

	return buildings, nil
}

// CountAll returns the total count of buildings
func (repository *RepositoryBuildingImpl) CountAll(ctx context.Context, tx *sql.Tx) (int, error) {
	SQL := "SELECT COUNT(*) FROM " + models.BuildingTable
	row := tx.QueryRowContext(ctx, SQL)

	var total int
	err := row.Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// Update updates user-editable fields only
func (repository *RepositoryBuildingImpl) Update(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error) {
	SQL := `UPDATE ` + models.BuildingTable + ` 
		SET sellable = $1, connectivity = $2, resource_type = $3, updated_at = $4 
		WHERE id = $5 
		RETURNING updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		building.Sellable,
		building.Connectivity,
		nullIfEmpty(building.ResourceType),
		time.Now(),
		building.Id,
	).Scan(&building.UpdatedAt)

	if err != nil {
		return models.Building{}, err
	}

	return building, nil
}

// UpdateFromSync updates ERP-sourced fields only (preserves user inputs)
func (repository *RepositoryBuildingImpl) UpdateFromSync(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error) {
	SQL := `UPDATE ` + models.BuildingTable + ` 
		SET external_building_id = $1, iris_code = $2, name = $3, project_name = $4, 
		audience = $5, impression = $6, cbd_area = $7, building_status = $8, 
		competitor_location = $9, synced_at = $10, updated_at = $11 
		WHERE id = $12 
		RETURNING updated_at`

	err := tx.QueryRowContext(ctx, SQL,
		building.ExternalBuildingId,
		building.IrisCode,
		building.Name,
		building.ProjectName,
		building.Audience,
		building.Impression,
		building.CbdArea,
		building.BuildingStatus,
		building.CompetitorLocation,
		time.Now(),
		time.Now(),
		building.Id,
	).Scan(&building.UpdatedAt)

	if err != nil {
		return models.Building{}, err
	}

	return building, nil
}

// Helper functions
func nullIfZero(value int) interface{} {
	if value == 0 {
		return nil
	}
	return value
}

func nullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

