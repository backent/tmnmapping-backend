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
		cbd_area, building_status, competitor_location, competitor_exclusive, competitor_presence, sellable, connectivity, 
		resource_type, subdistrict, citytown, province, grade_resource, building_type, completion_year, latitude, longitude, location, images, lcd_presence_status, synced_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, 
		CASE WHEN $21::DOUBLE PRECISION IS NOT NULL AND $22::DOUBLE PRECISION IS NOT NULL AND ($21::DOUBLE PRECISION) != 0 AND ($22::DOUBLE PRECISION) != 0 THEN ST_SetSRID(ST_MakePoint($22::DOUBLE PRECISION, $21::DOUBLE PRECISION), 4326)::geography ELSE NULL END, $23, $24, $25) 
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
		building.CompetitorExclusive,
		building.CompetitorPresence,
		building.Sellable,
		building.Connectivity,
		nullIfEmpty(building.ResourceType),
		nullIfEmpty(building.Subdistrict),
		nullIfEmpty(building.Citytown),
		nullIfEmpty(building.Province),
		nullIfEmpty(building.GradeResource),
		nullIfEmpty(building.BuildingType),
		nullIfZero(building.CompletionYear),
		nullIfZeroFloat(building.Latitude),
		nullIfZeroFloat(building.Longitude),
		imagesJSON,
		nullIfEmpty(building.LcdPresenceStatus),
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
		impression, cbd_area, building_status, competitor_location, competitor_exclusive, competitor_presence, sellable, connectivity, 
		resource_type, subdistrict, citytown, province, grade_resource, building_type, completion_year, latitude, longitude, images, lcd_presence_status, synced_at, created_at, updated_at 
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
			&building.CompetitorExclusive,
			&building.CompetitorPresence,
			&building.Sellable,
			&building.Connectivity,
			&building.ResourceType,
			&building.Subdistrict,
			&building.Citytown,
			&building.Province,
			&building.GradeResource,
			&building.BuildingType,
			&building.CompletionYear,
			&building.Latitude,
			&building.Longitude,
			&building.Images,
			&building.LcdPresenceStatus,
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
		impression, cbd_area, building_status, competitor_location, competitor_exclusive, competitor_presence, sellable, connectivity, 
		resource_type, subdistrict, citytown, province, grade_resource, building_type, completion_year, latitude, longitude, images, lcd_presence_status, synced_at, created_at, updated_at 
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
			&building.CompetitorExclusive,
			&building.CompetitorPresence,
			&building.Sellable,
			&building.Connectivity,
			&building.ResourceType,
			&building.Subdistrict,
			&building.Citytown,
			&building.Province,
			&building.GradeResource,
			&building.BuildingType,
			&building.CompletionYear,
			&building.Latitude,
			&building.Longitude,
			&building.Images,
			&building.LcdPresenceStatus,
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
		impression, cbd_area, building_status, competitor_location, competitor_exclusive, competitor_presence, sellable, connectivity, 
		resource_type, subdistrict, citytown, province, grade_resource, building_type, completion_year, latitude, longitude, images, lcd_presence_status, synced_at, created_at, updated_at 
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
			&building.CompetitorExclusive,
			&building.CompetitorPresence,
			&building.Sellable,
			&building.Connectivity,
			&building.ResourceType,
			&building.Subdistrict,
			&building.Citytown,
			&building.Province,
			&building.GradeResource,
			&building.BuildingType,
			&building.CompletionYear,
			&building.Latitude,
			&building.Longitude,
			&building.Images,
			&building.LcdPresenceStatus,
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
func (repository *RepositoryBuildingImpl) FindAllForMapping(ctx context.Context, tx *sql.Tx, buildingType string, buildingGrade string, year string, subdistrict string, progress string, sellable string, connectivity string, lcdPresence string, salesPackageIds string, lat *float64, lng *float64, radius *int, poiPoints []struct{ Lat float64; Lng float64 }, polygonPoints []struct{ Lat float64; Lng float64 }, minLat *float64, maxLat *float64, minLng *float64, maxLng *float64) ([]models.Building, error) {
	SQL := `SELECT DISTINCT b.id, b.external_building_id, b.iris_code, b.name, b.project_name, b.audience, 
		b.impression, b.cbd_area, b.building_status, b.competitor_location, b.competitor_exclusive, b.competitor_presence, b.sellable, b.connectivity, 
		b.resource_type, b.subdistrict, b.citytown, b.province, b.grade_resource, b.building_type, b.completion_year, b.latitude, b.longitude, b.images, b.lcd_presence_status, b.synced_at, b.created_at, b.updated_at 
		FROM ` + models.BuildingTable + ` b`

	args := []interface{}{}
	argIndex := 1
	whereConditions := []string{}
	joinClauses := []string{}

	// Add sales package filter - JOIN with sales_package_buildings table
	if salesPackageIds != "" {
		if strings.Contains(salesPackageIds, ",") {
			// Multiple values: use IN clause
			packageIds := strings.Split(salesPackageIds, ",")
			placeholders := make([]string, len(packageIds))
			for i := range packageIds {
				placeholders[i] = "$" + strconv.Itoa(argIndex+i)
				args = append(args, strings.TrimSpace(packageIds[i]))
			}
			joinClauses = append(joinClauses, `INNER JOIN `+models.SalesPackageBuildingTable+` spb ON b.id = spb.building_id`)
			whereConditions = append(whereConditions, `spb.sales_package_id IN (`+strings.Join(placeholders, ",")+`)`)
			argIndex += len(packageIds)
		} else {
			// Single value
			joinClauses = append(joinClauses, `INNER JOIN `+models.SalesPackageBuildingTable+` spb ON b.id = spb.building_id`)
			whereConditions = append(whereConditions, `spb.sales_package_id = $`+strconv.Itoa(argIndex))
			args = append(args, strings.TrimSpace(salesPackageIds))
			argIndex++
		}
	}

	// Add building_type filter - handle comma-separated values with case-insensitive comparison
	if buildingType != "" {
		if strings.Contains(buildingType, ",") {
			// Multiple values: use IN clause with case-insensitive comparison
			types := strings.Split(buildingType, ",")
			placeholders := make([]string, len(types))
			for i := range types {
				placeholders[i] = "$" + strconv.Itoa(argIndex+i)
				// Convert to lowercase for case-insensitive comparison
				args = append(args, strings.ToLower(strings.TrimSpace(types[i])))
			}
			// Use LOWER() on database column for case-insensitive comparison
			whereConditions = append(whereConditions, `LOWER(building_type) IN (`+strings.Join(placeholders, ",")+`)`)
			argIndex += len(types)
		} else {
			// Single value with case-insensitive comparison
			whereConditions = append(whereConditions, `LOWER(building_type) = $`+strconv.Itoa(argIndex))
			args = append(args, strings.ToLower(buildingType))
			argIndex++
		}
	}

	// Add grade_resource filter (mapped from building_grade) - handle comma-separated values with case-insensitive comparison
	if buildingGrade != "" {
		if strings.Contains(buildingGrade, ",") {
			// Multiple values: use IN clause with case-insensitive comparison
			grades := strings.Split(buildingGrade, ",")
			placeholders := make([]string, len(grades))
			for i := range grades {
				placeholders[i] = "$" + strconv.Itoa(argIndex+i)
				// Convert to lowercase for case-insensitive comparison
				args = append(args, strings.ToLower(strings.TrimSpace(grades[i])))
			}
			// Use LOWER() on database column for case-insensitive comparison
			whereConditions = append(whereConditions, `LOWER(grade_resource) IN (`+strings.Join(placeholders, ",")+`)`)
			argIndex += len(grades)
		} else {
			// Single value with case-insensitive comparison
			whereConditions = append(whereConditions, `LOWER(grade_resource) = $`+strconv.Itoa(argIndex))
			args = append(args, strings.ToLower(buildingGrade))
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

	// Add subdistrict filter - handle comma-separated values with OR logic
	if subdistrict != "" {
		if strings.Contains(subdistrict, ",") {
			// Multiple values: use OR conditions with ILIKE for each value
			subdistricts := strings.Split(subdistrict, ",")
			orConditions := make([]string, len(subdistricts))
			for i, sd := range subdistricts {
				orConditions[i] = `subdistrict ILIKE $` + strconv.Itoa(argIndex+i)
				args = append(args, "%"+strings.TrimSpace(sd)+"%")
			}
			whereConditions = append(whereConditions, `(`+strings.Join(orConditions, " OR ")+`)`)
			argIndex += len(subdistricts)
		} else {
			// Single value: use ILIKE pattern matching
			whereConditions = append(whereConditions, `subdistrict ILIKE $`+strconv.Itoa(argIndex))
			args = append(args, "%"+subdistrict+"%")
			argIndex++
		}
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

	// Add lcd_presence_status filter - handle comma-separated values and null
	if lcdPresence != "" {
		if strings.Contains(lcdPresence, ",") {
			// Multiple values: use IN clause or IS NULL
			statuses := strings.Split(lcdPresence, ",")
			placeholders := make([]string, 0, len(statuses))
			hasNull := false
			for _, status := range statuses {
				trimmedStatus := strings.TrimSpace(status)
				if trimmedStatus == "Opportunity" || trimmedStatus == "" {
					// Include NULL values for Opportunity
					hasNull = true
				}
				if trimmedStatus != "" {
					placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
					args = append(args, trimmedStatus)
					argIndex++
				}
			}
			if len(placeholders) > 0 && hasNull {
				whereConditions = append(whereConditions, `(lcd_presence_status IN (`+strings.Join(placeholders, ",")+`) OR lcd_presence_status IS NULL)`)
			} else if len(placeholders) > 0 {
				whereConditions = append(whereConditions, `lcd_presence_status IN (`+strings.Join(placeholders, ",")+`)`)
			} else if hasNull {
				whereConditions = append(whereConditions, `lcd_presence_status IS NULL`)
			}
		} else {
			// Single value
			trimmedStatus := strings.TrimSpace(lcdPresence)
			if trimmedStatus == "Opportunity" {
				// For Opportunity, include both NULL and 'Opportunity' values
				whereConditions = append(whereConditions, `(lcd_presence_status = $`+strconv.Itoa(argIndex)+` OR lcd_presence_status IS NULL)`)
				args = append(args, trimmedStatus)
				argIndex++
			} else {
				whereConditions = append(whereConditions, `lcd_presence_status = $`+strconv.Itoa(argIndex))
				args = append(args, trimmedStatus)
				argIndex++
			}
		}
	}

	// Spatial filter: polygon (ST_Within) takes priority; else POI/radius (ST_DWithin)
	if len(polygonPoints) >= 3 {
		// Build closed WKT POLYGON: lng lat order, first point = last point
		ringParts := make([]string, 0, len(polygonPoints)+1)
		for _, p := range polygonPoints {
			ringParts = append(ringParts, strconv.FormatFloat(p.Lng, 'f', -1, 64)+" "+strconv.FormatFloat(p.Lat, 'f', -1, 64))
		}
		ringParts = append(ringParts, strconv.FormatFloat(polygonPoints[0].Lng, 'f', -1, 64)+" "+strconv.FormatFloat(polygonPoints[0].Lat, 'f', -1, 64))
		wkt := "POLYGON((" + strings.Join(ringParts, ", ") + "))"
		// ST_Within(geography, geography) does not exist; cast to geometry for the check
		whereConditions = append(whereConditions, `ST_Within(location::geometry, ST_SetSRID(ST_GeomFromText($`+strconv.Itoa(argIndex)+`), 4326)::geometry)`)
		args = append(args, wkt)
		argIndex++
	} else if len(poiPoints) > 0 && radius != nil && *radius > 0 {
		orConditions := make([]string, len(poiPoints))
		for i, point := range poiPoints {
			orConditions[i] = `ST_DWithin(location, ST_SetSRID(ST_MakePoint($` + strconv.Itoa(argIndex+i*2) + `, $` + strconv.Itoa(argIndex+i*2+1) + `), 4326)::geography, $` + strconv.Itoa(argIndex+len(poiPoints)*2) + `)`
			args = append(args, point.Lng, point.Lat)
		}
		args = append(args, *radius)
		whereConditions = append(whereConditions, `(`+strings.Join(orConditions, " OR ")+`)`)
		argIndex += len(poiPoints)*2 + 1
	} else if lat != nil && lng != nil && radius != nil && *radius > 0 {
		whereConditions = append(whereConditions, `ST_DWithin(location, ST_SetSRID(ST_MakePoint($`+strconv.Itoa(argIndex)+`, $`+strconv.Itoa(argIndex+1)+`), 4326)::geography, $`+strconv.Itoa(argIndex+2)+`)`)
		args = append(args, *lng, *lat, *radius)
		argIndex += 3
	}

	// Bounds filter (map viewport): only when all four are provided
	if minLat != nil && maxLat != nil && minLng != nil && maxLng != nil {
		whereConditions = append(whereConditions, `latitude >= $`+strconv.Itoa(argIndex)+` AND latitude <= $`+strconv.Itoa(argIndex+1)+` AND longitude >= $`+strconv.Itoa(argIndex+2)+` AND longitude <= $`+strconv.Itoa(argIndex+3))
		args = append(args, *minLat, *maxLat, *minLng, *maxLng)
		argIndex += 4
	}

	// Build JOIN clauses
	if len(joinClauses) > 0 {
		for _, join := range joinClauses {
			SQL += ` ` + join
		}
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
			&building.CompetitorExclusive,
			&building.CompetitorPresence,
			&building.Sellable,
			&building.Connectivity,
			&building.ResourceType,
			&building.Subdistrict,
			&building.Citytown,
			&building.Province,
			&building.GradeResource,
			&building.BuildingType,
			&building.CompletionYear,
			&building.Latitude,
			&building.Longitude,
			&building.Images,
			&building.LcdPresenceStatus,
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
		competitor_location = $9, competitor_exclusive = $10, competitor_presence = $11, subdistrict = $12, citytown = $13, province = $14, 
		grade_resource = $15, building_type = $16, completion_year = $17, 
		latitude = $18, longitude = $19, 
		location = CASE WHEN $18::DOUBLE PRECISION IS NOT NULL AND $19::DOUBLE PRECISION IS NOT NULL AND ($18::DOUBLE PRECISION) != 0 AND ($19::DOUBLE PRECISION) != 0 THEN ST_SetSRID(ST_MakePoint($19::DOUBLE PRECISION, $18::DOUBLE PRECISION), 4326)::geography ELSE NULL END,
		images = $20, lcd_presence_status = $21, synced_at = $22, updated_at = $23 
		WHERE id = $24 
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
		building.CompetitorExclusive,
		building.CompetitorPresence,
		nullIfEmpty(building.Subdistrict),
		nullIfEmpty(building.Citytown),
		nullIfEmpty(building.Province),
		nullIfEmpty(building.GradeResource),
		nullIfEmpty(building.BuildingType),
		nullIfZero(building.CompletionYear),
		nullIfZeroFloat(building.Latitude),
		nullIfZeroFloat(building.Longitude),
		imagesJSON,
		nullIfEmpty(building.LcdPresenceStatus),
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

func nullIfZeroFloat(value float64) interface{} {
	if value == 0 {
		return nil
	}
	return value
}
