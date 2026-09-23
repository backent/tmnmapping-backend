package building

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/models"
)

// changesPerInsert keeps one statement inside Postgres' 65535-parameter limit at 11
// parameters per row. A 3,771-row import touching several fields each is a handful of
// statements rather than thousands of round trips.
const changeColumns = 11
const changesPerInsert = 500

// RecordChanges writes the audit rows. Always called inside the same transaction as
// the write it describes, so a change cannot be logged for a write that rolled back,
// nor a write land unlogged.
func (r *RepositoryBuildingImpl) RecordChanges(ctx context.Context, tx *sql.Tx, changes []models.BuildingChange) error {
	if len(changes) == 0 {
		return nil
	}

	for start := 0; start < len(changes); start += changesPerInsert {
		end := start + changesPerInsert
		if end > len(changes) {
			end = len(changes)
		}

		SQL, args := buildChangesInsert(changes[start:end])
		if _, err := tx.ExecContext(ctx, SQL, args...); err != nil {
			return err
		}
	}

	return nil
}

func buildChangesInsert(changes []models.BuildingChange) (string, []interface{}) {
	placeholders := make([]string, 0, len(changes))
	args := make([]interface{}, 0, len(changes)*changeColumns)

	for i, change := range changes {
		base := i * changeColumns
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d::uuid, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11))

		args = append(args,
			changeNullIfZero(int64(change.BuildingId)), changeNullIfEmpty(change.ExternalBuildingId),
			change.BuildingName,
			changeNullIfZero(int64(change.ActorUserId)), changeNullIfEmpty(change.ActorRole),
			change.Action, change.Source, changeNullIfEmpty(change.BatchId),
			changeNullIfEmpty(change.Field), changeNullIfEmpty(change.OldValue),
			changeNullIfEmpty(change.NewValue))
	}

	return `INSERT INTO ` + models.BuildingChangeTable + `
		(building_id, external_building_id, building_name, actor_user_id, actor_role,
		 action, source, batch_id, field, old_value, new_value)
		VALUES ` + strings.Join(placeholders, ", "), args
}

func changeNullIfEmpty(v string) interface{} {
	if strings.TrimSpace(v) == "" {
		return nil
	}

	return v
}

func changeNullIfZero(v int64) interface{} {
	if v == 0 {
		return nil
	}

	return v
}

// FindChanges is the history of one building, newest first.
func (r *RepositoryBuildingImpl) FindChanges(ctx context.Context, tx *sql.Tx, buildingId int, take int, skip int) ([]models.BuildingChange, error) {
	SQL := `SELECT c.id, c.building_id, c.external_building_id, c.building_name,
		c.actor_user_id, u.name, c.actor_role, c.action, c.source, c.batch_id,
		c.field, c.old_value, c.new_value, c.created_at
		FROM ` + models.BuildingChangeTable + ` c
		LEFT JOIN ` + models.UserTable + ` u ON u.id = c.actor_user_id
		WHERE c.building_id = $1
		ORDER BY c.created_at DESC, c.id DESC LIMIT $2 OFFSET $3`

	rows, err := tx.QueryContext(ctx, SQL, buildingId, take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changes := []models.BuildingChange{}
	for rows.Next() {
		var n models.NullAbleBuildingChange
		if err := rows.Scan(&n.Id, &n.BuildingId, &n.ExternalBuildingId, &n.BuildingName,
			&n.ActorUserId, &n.ActorName, &n.ActorRole, &n.Action, &n.Source, &n.BatchId,
			&n.Field, &n.OldValue, &n.NewValue, &n.CreatedAt); err != nil {
			return nil, err
		}
		changes = append(changes, models.NullAbleBuildingChangeToChange(n))
	}

	return changes, rows.Err()
}

func (r *RepositoryBuildingImpl) CountChanges(ctx context.Context, tx *sql.Tx, buildingId int) (int, error) {
	var total int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM `+models.BuildingChangeTable+` WHERE building_id = $1`,
		buildingId).Scan(&total)

	return total, err
}

// FindAllForExport reads every building in key order, so an export is stable between
// runs and a diff between two exports shows real changes rather than reordering.
func (r *RepositoryBuildingImpl) FindAllForExport(ctx context.Context, tx *sql.Tx) ([]models.Building, error) {
	SQL := `SELECT b.id, b.external_building_id, b.iris_code, b.name, b.project_name,
		b.project_id, p.project_id_iris, b.audience,
		b.impression, b.cbd_area, b.building_status, b.competitor_location, b.competitor_exclusive,
		b.competitor_presence, b.sellable, b.connectivity, b.resource_type, b.subdistrict, b.citytown,
		b.province, b.grade_resource, b.building_type, b.completion_year, b.latitude, b.longitude,
		b.lcd_presence_status
		FROM ` + models.BuildingTable + ` b
		LEFT JOIN ` + models.BuildingProjectTable + ` p ON p.id = b.project_id
		ORDER BY b.external_building_id NULLS LAST, b.id`

	rows, err := tx.QueryContext(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buildings := []models.Building{}
	for rows.Next() {
		var n models.NullAbleBuilding
		if err := rows.Scan(&n.Id, &n.ExternalBuildingId, &n.IrisCode, &n.Name, &n.ProjectName,
			&n.ProjectId, &n.ProjectIdIris, &n.Audience, &n.Impression, &n.CbdArea, &n.BuildingStatus, &n.CompetitorLocation,
			&n.CompetitorExclusive, &n.CompetitorPresence, &n.Sellable, &n.Connectivity,
			&n.ResourceType, &n.Subdistrict, &n.Citytown, &n.Province, &n.GradeResource,
			&n.BuildingType, &n.CompletionYear, &n.Latitude, &n.Longitude,
			&n.LcdPresenceStatus); err != nil {
			return nil, err
		}

		buildings = append(buildings, models.Building{
			Id: int(n.Id.Int64), ExternalBuildingId: n.ExternalBuildingId.String,
			IrisCode: n.IrisCode.String, Name: n.Name.String, ProjectName: n.ProjectName.String,
			ProjectId: int(n.ProjectId.Int64), ProjectIdIris: n.ProjectIdIris.String,
			Audience: int(n.Audience.Int64), Impression: int(n.Impression.Int64),
			CbdArea: n.CbdArea.String, BuildingStatus: n.BuildingStatus.String,
			CompetitorLocation: n.CompetitorLocation.Bool, CompetitorExclusive: n.CompetitorExclusive.Bool,
			CompetitorPresence: n.CompetitorPresence.Bool, Sellable: n.Sellable.String,
			Connectivity: n.Connectivity.String, ResourceType: n.ResourceType.String,
			Subdistrict: n.Subdistrict.String, Citytown: n.Citytown.String,
			Province: n.Province.String, GradeResource: n.GradeResource.String,
			BuildingType: n.BuildingType.String, CompletionYear: int(n.CompletionYear.Int64),
			Latitude: n.Latitude.Float64, Longitude: n.Longitude.Float64,
			LcdPresenceStatus: n.LcdPresenceStatus.String,
		})
	}

	return buildings, rows.Err()
}

// importUpdateColumns is the set of columns the spreadsheet owns, in one place so a
// test can check it against the fields the change log tracks. The two must agree: a
// column tracked but not written means the audit trail claims a change the database
// never received, which is worse than not logging at all.
//
// Deliberately absent: images and synced_at, which the ERP photo sync owns;
// project_id, set by the project link rather than this sheet; and created_at.
var importUpdateColumns = []string{
	"external_building_id", "iris_code", "name", "project_name", "project_id",
	"latitude", "longitude", "subdistrict", "citytown", "province", "cbd_area",
	"building_type", "grade_resource", "completion_year", "building_status",
	"competitor_presence", "competitor_exclusive", "competitor_location",
	"audience", "impression", "sellable", "connectivity", "resource_type",
	"lcd_presence_status",
}

// ImportUpdateColumns exposes that list for the guard test.
func ImportUpdateColumns() []string {
	out := make([]string, len(importUpdateColumns))
	copy(out, importUpdateColumns)

	return out
}

// UpdateFromImport writes every column the spreadsheet owns.
//
// It exists because Update() is the building FORM's update and sets only sellable,
// connectivity and resource_type -- the three fields that form can edit. Calling that
// from the importer wrote nothing while the change log recorded everything, so an
// upload appeared to work and changed no data.
//
// location is rebuilt from the coordinates rather than taken from the caller, and is
// left NULL when either is missing or zero -- the same rule the ERP sync uses, so a
// building with no coordinates stays off the map instead of landing at 0,0.
func (repository *RepositoryBuildingImpl) UpdateFromImport(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error) {
	SQL := `UPDATE ` + models.BuildingTable + ` SET
		external_building_id = $1, iris_code = $2, name = $3, project_name = $4,
		project_id = $25,
		latitude = $5, longitude = $6,
		location = CASE
			WHEN $5::DOUBLE PRECISION IS NOT NULL AND $6::DOUBLE PRECISION IS NOT NULL
			 AND ($5::DOUBLE PRECISION) != 0 AND ($6::DOUBLE PRECISION) != 0
			THEN ST_SetSRID(ST_MakePoint($6::DOUBLE PRECISION, $5::DOUBLE PRECISION), 4326)::geography
			ELSE NULL END,
		subdistrict = $7, citytown = $8, province = $9, cbd_area = $10,
		building_type = $11, grade_resource = $12, completion_year = $13,
		building_status = $14, competitor_presence = $15, competitor_exclusive = $16,
		competitor_location = $17, audience = $18, impression = $19,
		sellable = $20, connectivity = $21, resource_type = $22,
		lcd_presence_status = $23, updated_at = CURRENT_TIMESTAMP
		WHERE id = $24`

	_, err := tx.ExecContext(ctx, SQL,
		changeNullIfEmpty(building.ExternalBuildingId), changeNullIfEmpty(building.IrisCode),
		building.Name, changeNullIfEmpty(building.ProjectName),
		nullIfZeroCoordinate(building.Latitude), nullIfZeroCoordinate(building.Longitude),
		changeNullIfEmpty(building.Subdistrict), changeNullIfEmpty(building.Citytown),
		changeNullIfEmpty(building.Province), changeNullIfEmpty(building.CbdArea),
		changeNullIfEmpty(building.BuildingType), changeNullIfEmpty(building.GradeResource),
		changeNullIfZero(int64(building.CompletionYear)),
		changeNullIfEmpty(building.BuildingStatus),
		building.CompetitorPresence, building.CompetitorExclusive, building.CompetitorLocation,
		changeNullIfZero(int64(building.Audience)), changeNullIfZero(int64(building.Impression)),
		changeNullIfEmpty(building.Sellable), changeNullIfEmpty(building.Connectivity),
		changeNullIfEmpty(building.ResourceType),
		changeNullIfEmpty(building.LcdPresenceStatus),
		building.Id,
		changeNullIfZero(int64(building.ProjectId)),
	)
	if err != nil {
		return models.Building{}, err
	}

	return building, nil
}

// nullIfZeroCoordinate keeps 0,0 out of the table: it is a real place in the Gulf of
// Guinea, and no TMN building is there.
func nullIfZeroCoordinate(v float64) interface{} {
	if v == 0 {
		return nil
	}

	return v
}
