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
	SQL := `SELECT id, external_building_id, iris_code, name, project_name, audience,
		impression, cbd_area, building_status, competitor_location, competitor_exclusive,
		competitor_presence, sellable, connectivity, resource_type, subdistrict, citytown,
		province, grade_resource, building_type, completion_year, latitude, longitude,
		lcd_presence_status
		FROM ` + models.BuildingTable + `
		ORDER BY external_building_id NULLS LAST, id`

	rows, err := tx.QueryContext(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buildings := []models.Building{}
	for rows.Next() {
		var n models.NullAbleBuilding
		if err := rows.Scan(&n.Id, &n.ExternalBuildingId, &n.IrisCode, &n.Name, &n.ProjectName,
			&n.Audience, &n.Impression, &n.CbdArea, &n.BuildingStatus, &n.CompetitorLocation,
			&n.CompetitorExclusive, &n.CompetitorPresence, &n.Sellable, &n.Connectivity,
			&n.ResourceType, &n.Subdistrict, &n.Citytown, &n.Province, &n.GradeResource,
			&n.BuildingType, &n.CompletionYear, &n.Latitude, &n.Longitude,
			&n.LcdPresenceStatus); err != nil {
			return nil, err
		}

		buildings = append(buildings, models.Building{
			Id: int(n.Id.Int64), ExternalBuildingId: n.ExternalBuildingId.String,
			IrisCode: n.IrisCode.String, Name: n.Name.String, ProjectName: n.ProjectName.String,
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
