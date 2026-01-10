package building

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBuildingImpl struct{}

func NewRepositoryBuildingImpl() RepositoryBuildingInterface {
	return &RepositoryBuildingImpl{}
}

// Create inserts a new building
func (repository *RepositoryBuildingImpl) Create(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error) {
	// Marshal images to JSON
	imagesJSON, err := json.Marshal(building.Images)
	if err != nil {
		imagesJSON = []byte("[]")
	}
	if len(building.Images) == 0 {
		imagesJSON = nil
	}

	SQL := `INSERT INTO ` + models.BuildingTable + ` 
		(external_building_id, iris_code, name, project_name, audience, impression, 
		cbd_area, building_status, competitor_location, sellable, connectivity, 
		resource_type, subdistrict, citytown, province, grade_resource, building_type, completion_year, images, synced_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20) 
		RETURNING id, created_at, updated_at`

	err = tx.QueryRowContext(ctx, SQL,
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
		nullIfEmpty(building.Subdistrict),
		nullIfEmpty(building.Citytown),
		nullIfEmpty(building.Province),
		nullIfEmpty(building.GradeResource),
		nullIfEmpty(building.BuildingType),
		nullIfZero(building.CompletionYear),
		imagesJSON,
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
		resource_type, subdistrict, citytown, province, grade_resource, building_type, completion_year, images, synced_at, created_at, updated_at 
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
			&building.Subdistrict,
			&building.Citytown,
			&building.Province,
			&building.GradeResource,
			&building.BuildingType,
			&building.CompletionYear,
			&building.Images,
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
		resource_type, subdistrict, citytown, province, grade_resource, building_type, completion_year, images, synced_at, created_at, updated_at 
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
			&building.Subdistrict,
			&building.Citytown,
			&building.Province,
			&building.GradeResource,
			&building.BuildingType,
			&building.CompletionYear,
			&building.Images,
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

// FindAll retrieves all buildings with pagination, sorting, search, and filters
func (repository *RepositoryBuildingImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string, buildingStatus string, sellable string, connectivity string, resourceType string, competitorLocation *bool, cbdArea string, subdistrict string, citytown string, province string, gradeResource string, buildingType string) ([]models.Building, error) {
	SQL := `SELECT id, external_building_id, iris_code, name, project_name, audience, 
		impression, cbd_area, building_status, competitor_location, sellable, connectivity, 
		resource_type, subdistrict, citytown, province, grade_resource, building_type, completion_year, images, synced_at, created_at, updated_at 
		FROM ` + models.BuildingTable

	args := []interface{}{}
	argIndex := 1
	whereConditions := []string{}

	// Add search filter
	if search != "" {
		whereConditions = append(whereConditions, `name ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	// Add building_status filter
	if buildingStatus != "" {
		whereConditions = append(whereConditions, `building_status = $`+strconv.Itoa(argIndex))
		args = append(args, buildingStatus)
		argIndex++
	}

	// Add sellable filter
	if sellable != "" {
		whereConditions = append(whereConditions, `sellable = $`+strconv.Itoa(argIndex))
		args = append(args, sellable)
		argIndex++
	}

	// Add connectivity filter
	if connectivity != "" {
		whereConditions = append(whereConditions, `connectivity = $`+strconv.Itoa(argIndex))
		args = append(args, connectivity)
		argIndex++
	}

	// Add resource_type filter
	if resourceType != "" {
		whereConditions = append(whereConditions, `resource_type = $`+strconv.Itoa(argIndex))
		args = append(args, resourceType)
		argIndex++
	}

	// Add competitor_location filter
	if competitorLocation != nil {
		whereConditions = append(whereConditions, `competitor_location = $`+strconv.Itoa(argIndex))
		args = append(args, *competitorLocation)
		argIndex++
	}

	// Add cbd_area filter
	if cbdArea != "" {
		whereConditions = append(whereConditions, `cbd_area ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+cbdArea+"%")
		argIndex++
	}

	// Add subdistrict filter
	if subdistrict != "" {
		whereConditions = append(whereConditions, `subdistrict ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+subdistrict+"%")
		argIndex++
	}

	// Add citytown filter
	if citytown != "" {
		whereConditions = append(whereConditions, `citytown ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+citytown+"%")
		argIndex++
	}

	// Add province filter
	if province != "" {
		whereConditions = append(whereConditions, `province ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+province+"%")
		argIndex++
	}

	// Add grade_resource filter
	if gradeResource != "" {
		whereConditions = append(whereConditions, `grade_resource = $`+strconv.Itoa(argIndex))
		args = append(args, gradeResource)
		argIndex++
	}

	// Add building_type filter
	if buildingType != "" {
		whereConditions = append(whereConditions, `building_type = $`+strconv.Itoa(argIndex))
		args = append(args, buildingType)
		argIndex++
	}

	// Build WHERE clause
	if len(whereConditions) > 0 {
		SQL += ` WHERE ` + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			SQL += ` AND ` + whereConditions[i]
		}
	}

	SQL += ` ORDER BY ` + orderBy + ` ` + orderDirection + ` LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)
	args = append(args, take, skip)

	rows, err := tx.QueryContext(ctx, SQL, args...)
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
			&building.Subdistrict,
			&building.Citytown,
			&building.Province,
			&building.GradeResource,
			&building.BuildingType,
			&building.CompletionYear,
			&building.Images,
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

// CountAll returns the total count of buildings with optional search and filters
func (repository *RepositoryBuildingImpl) CountAll(ctx context.Context, tx *sql.Tx, search string, buildingStatus string, sellable string, connectivity string, resourceType string, competitorLocation *bool, cbdArea string, subdistrict string, citytown string, province string, gradeResource string, buildingType string) (int, error) {
	SQL := "SELECT COUNT(*) FROM " + models.BuildingTable

	args := []interface{}{}
	argIndex := 1
	whereConditions := []string{}

	// Add search filter
	if search != "" {
		whereConditions = append(whereConditions, `name ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	// Add building_status filter
	if buildingStatus != "" {
		whereConditions = append(whereConditions, `building_status = $`+strconv.Itoa(argIndex))
		args = append(args, buildingStatus)
		argIndex++
	}

	// Add sellable filter
	if sellable != "" {
		whereConditions = append(whereConditions, `sellable = $`+strconv.Itoa(argIndex))
		args = append(args, sellable)
		argIndex++
	}

	// Add connectivity filter
	if connectivity != "" {
		whereConditions = append(whereConditions, `connectivity = $`+strconv.Itoa(argIndex))
		args = append(args, connectivity)
		argIndex++
	}

	// Add resource_type filter
	if resourceType != "" {
		whereConditions = append(whereConditions, `resource_type = $`+strconv.Itoa(argIndex))
		args = append(args, resourceType)
		argIndex++
	}

	// Add competitor_location filter
	if competitorLocation != nil {
		whereConditions = append(whereConditions, `competitor_location = $`+strconv.Itoa(argIndex))
		args = append(args, *competitorLocation)
		argIndex++
	}

	// Add cbd_area filter
	if cbdArea != "" {
		whereConditions = append(whereConditions, `cbd_area ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+cbdArea+"%")
		argIndex++
	}

	// Add subdistrict filter
	if subdistrict != "" {
		whereConditions = append(whereConditions, `subdistrict ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+subdistrict+"%")
		argIndex++
	}

	// Add citytown filter
	if citytown != "" {
		whereConditions = append(whereConditions, `citytown ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+citytown+"%")
		argIndex++
	}

	// Add province filter
	if province != "" {
		whereConditions = append(whereConditions, `province ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+province+"%")
		argIndex++
	}

	// Add grade_resource filter
	if gradeResource != "" {
		whereConditions = append(whereConditions, `grade_resource = $`+strconv.Itoa(argIndex))
		args = append(args, gradeResource)
		argIndex++
	}

	// Add building_type filter
	if buildingType != "" {
		whereConditions = append(whereConditions, `building_type = $`+strconv.Itoa(argIndex))
		args = append(args, buildingType)
		argIndex++
	}

	// Build WHERE clause
	if len(whereConditions) > 0 {
		SQL += " WHERE " + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			SQL += " AND " + whereConditions[i]
		}
	}

	row := tx.QueryRowContext(ctx, SQL, args...)
	var total int
	err := row.Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetDistinctValues returns distinct values for a given column
func (repository *RepositoryBuildingImpl) GetDistinctValues(ctx context.Context, tx *sql.Tx, columnName string) ([]string, error) {
	// Validate column name to prevent SQL injection
	validColumns := map[string]bool{
		"building_status": true,
		"sellable":        true,
		"connectivity":    true,
		"resource_type":   true,
		"cbd_area":        true,
		"subdistrict":     true,
		"citytown":        true,
		"province":        true,
		"grade_resource":  true,
		"building_type":   true,
	}

	if !validColumns[columnName] {
		return []string{}, nil
	}

	SQL := "SELECT DISTINCT " + columnName + " FROM " + models.BuildingTable + " WHERE " + columnName + " IS NOT NULL AND " + columnName + " != '' ORDER BY " + columnName

	rows, err := tx.QueryContext(ctx, SQL)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var value string
		err := rows.Scan(&value)
		if err != nil {
			return []string{}, err
		}
		values = append(values, value)
	}

	return values, nil
}

// FindAllForMapping retrieves all buildings for mapping with filters (no pagination)
func (repository *RepositoryBuildingImpl) FindAllForMapping(ctx context.Context, tx *sql.Tx, buildingType string, buildingGrade string, year string, subdistrict string, progress string, sellable string, connectivity string) ([]models.Building, error) {
	SQL := `SELECT id, external_building_id, iris_code, name, project_name, audience, 
		impression, cbd_area, building_status, competitor_location, sellable, connectivity, 
		resource_type, subdistrict, citytown, province, grade_resource, building_type, completion_year, images, synced_at, created_at, updated_at 
		FROM ` + models.BuildingTable

	args := []interface{}{}
	argIndex := 1
	whereConditions := []string{}

	// Add building_type filter - handle comma-separated values
	if buildingType != "" {
		if strings.Contains(buildingType, ",") {
			// Multiple values: use IN clause
			types := strings.Split(buildingType, ",")
			placeholders := make([]string, len(types))
			for i := range types {
				placeholders[i] = "$" + strconv.Itoa(argIndex+i)
				args = append(args, strings.TrimSpace(types[i]))
			}
			whereConditions = append(whereConditions, `building_type IN (`+strings.Join(placeholders, ",")+`)`)
			argIndex += len(types)
		} else {
			// Single value
			whereConditions = append(whereConditions, `building_type = $`+strconv.Itoa(argIndex))
			args = append(args, buildingType)
			argIndex++
		}
	}

	// Add grade_resource filter (mapped from building_grade) - handle comma-separated values
	if buildingGrade != "" {
		if strings.Contains(buildingGrade, ",") {
			// Multiple values: use IN clause
			grades := strings.Split(buildingGrade, ",")
			placeholders := make([]string, len(grades))
			for i := range grades {
				placeholders[i] = "$" + strconv.Itoa(argIndex+i)
				args = append(args, strings.TrimSpace(grades[i]))
			}
			whereConditions = append(whereConditions, `grade_resource IN (`+strings.Join(placeholders, ",")+`)`)
			argIndex += len(grades)
		} else {
			// Single value
			whereConditions = append(whereConditions, `grade_resource = $`+strconv.Itoa(argIndex))
			args = append(args, buildingGrade)
			argIndex++
		}
	}

	// Add completion_year filter (supports range: "2010,2020" or single value)
	if year != "" {
		if strings.Contains(year, ",") {
			// Range: parse as min,max
			parts := strings.Split(year, ",")
			if len(parts) == 2 {
				minYear := strings.TrimSpace(parts[0])
				maxYear := strings.TrimSpace(parts[1])
				if minYear != "" && maxYear != "" {
					whereConditions = append(whereConditions, `completion_year >= $`+strconv.Itoa(argIndex)+` AND completion_year <= $`+strconv.Itoa(argIndex+1))
					args = append(args, minYear, maxYear)
					argIndex += 2
				} else if minYear != "" {
					whereConditions = append(whereConditions, `completion_year >= $`+strconv.Itoa(argIndex))
					args = append(args, minYear)
					argIndex++
				} else if maxYear != "" {
					whereConditions = append(whereConditions, `completion_year <= $`+strconv.Itoa(argIndex))
					args = append(args, maxYear)
					argIndex++
				}
			}
		} else {
			// Single value: exact match
			whereConditions = append(whereConditions, `completion_year = $`+strconv.Itoa(argIndex))
			args = append(args, year)
			argIndex++
		}
	}

	// Add subdistrict filter
	if subdistrict != "" {
		whereConditions = append(whereConditions, `subdistrict ILIKE $`+strconv.Itoa(argIndex))
		args = append(args, "%"+subdistrict+"%")
		argIndex++
	}

	// Add building_status filter (mapped from progress)
	if progress != "" {
		whereConditions = append(whereConditions, `building_status = $`+strconv.Itoa(argIndex))
		args = append(args, progress)
		argIndex++
	}

	// Add sellable filter
	if sellable != "" {
		whereConditions = append(whereConditions, `sellable = $`+strconv.Itoa(argIndex))
		args = append(args, sellable)
		argIndex++
	}

	// Add connectivity filter
	if connectivity != "" {
		whereConditions = append(whereConditions, `connectivity = $`+strconv.Itoa(argIndex))
		args = append(args, connectivity)
		argIndex++
	}

	// Build WHERE clause
	if len(whereConditions) > 0 {
		SQL += ` WHERE ` + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			SQL += ` AND ` + whereConditions[i]
		}
	}

	rows, err := tx.QueryContext(ctx, SQL, args...)
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
			&building.Subdistrict,
			&building.Citytown,
			&building.Province,
			&building.GradeResource,
			&building.BuildingType,
			&building.CompletionYear,
			&building.Images,
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
	// Marshal images to JSON
	imagesJSON, err := json.Marshal(building.Images)
	if err != nil {
		imagesJSON = []byte("[]")
	}
	if len(building.Images) == 0 {
		imagesJSON = nil
	}

	SQL := `UPDATE ` + models.BuildingTable + ` 
		SET external_building_id = $1, iris_code = $2, name = $3, project_name = $4, 
		audience = $5, impression = $6, cbd_area = $7, building_status = $8, 
		competitor_location = $9, subdistrict = $10, citytown = $11, province = $12, 
		grade_resource = $13, building_type = $14, completion_year = $15, images = $16, synced_at = $17, updated_at = $18 
		WHERE id = $19 
		RETURNING updated_at`

	err = tx.QueryRowContext(ctx, SQL,
		building.ExternalBuildingId,
		building.IrisCode,
		building.Name,
		building.ProjectName,
		building.Audience,
		building.Impression,
		building.CbdArea,
		building.BuildingStatus,
		building.CompetitorLocation,
		nullIfEmpty(building.Subdistrict),
		nullIfEmpty(building.Citytown),
		nullIfEmpty(building.Province),
		nullIfEmpty(building.GradeResource),
		nullIfEmpty(building.BuildingType),
		nullIfZero(building.CompletionYear),
		imagesJSON,
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
