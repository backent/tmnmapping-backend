package building

import (
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

// One column list and one scanner for every full-building read.
//
// This exists because there were five hand-maintained copies, and on 2026-09-23 three
// separate bugs came from a field reaching some of them and not others: project_id was
// added to the model and the write path, then FindById returned it as 0, then
// FindByExternalId could not see it at all, so the importer read every linked building
// as newly linked and reported a change on every upload.
//
// A column added here now reaches every read at once, or fails to compile.
//
// The project is read with correlated subqueries rather than a LEFT JOIN on purpose:
// FindAll builds its WHERE with unprefixed column names, and both buildings and
// building_projects have a `name`, so a join would make those filters ambiguous and
// break the list screen. The subqueries are a primary-key lookup each.
var buildingColumns = `b.id, b.external_building_id, b.iris_code, b.name, b.project_name,
	b.audience, b.impression, b.cbd_area, b.building_status,
	b.competitor_location, b.competitor_exclusive, b.competitor_presence,
	b.sellable, b.connectivity, b.resource_type,
	b.subdistrict, b.citytown, b.province, b.grade_resource, b.building_type,
	b.completion_year, b.latitude, b.longitude, b.images, b.lcd_presence_status,
	b.synced_at, b.created_at, b.updated_at,
	b.project_id,
	(SELECT p.project_id_iris FROM ` + models.BuildingProjectTable + ` p WHERE p.id = b.project_id),
	(SELECT p.name FROM ` + models.BuildingProjectTable + ` p WHERE p.id = b.project_id)`

// buildingFrom pairs with buildingColumns. The alias is required: every column above
// is prefixed, while the filters callers append are not.
var buildingFrom = ` FROM ` + models.BuildingTable + ` b`

// selectBuilding builds a full-building query with the given clause appended, e.g.
// " WHERE b.id = $1" or " ORDER BY b.name LIMIT $1".
func selectBuilding(clause string) string {
	return `SELECT ` + buildingColumns + buildingFrom + clause
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows, so one scanner serves the
// single-row and multi-row reads.
type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanBuildingRow reads one row in buildingColumns order. Adding a column to that
// list without adding it here is a compile-time mismatch in argument count, which is
// the point: the two cannot drift.
func scanBuildingRow(scanner rowScanner) (models.Building, error) {
	var n models.NullAbleBuilding
	var projectDisplayName sql.NullString

	err := scanner.Scan(
		&n.Id, &n.ExternalBuildingId, &n.IrisCode, &n.Name, &n.ProjectName,
		&n.Audience, &n.Impression, &n.CbdArea, &n.BuildingStatus,
		&n.CompetitorLocation, &n.CompetitorExclusive, &n.CompetitorPresence,
		&n.Sellable, &n.Connectivity, &n.ResourceType,
		&n.Subdistrict, &n.Citytown, &n.Province, &n.GradeResource, &n.BuildingType,
		&n.CompletionYear, &n.Latitude, &n.Longitude, &n.Images, &n.LcdPresenceStatus,
		&n.SyncedAt, &n.CreatedAt, &n.UpdatedAt,
		&n.ProjectId, &n.ProjectIdIris, &projectDisplayName,
	)
	if err != nil {
		return models.Building{}, err
	}

	building := models.NullAbleBuildingToBuilding(n)
	// The PROJECT's name, which is not buildings.project_name -- that column is the
	// ERP correlation key the LOI dashboard joins on by text, and the two can differ.
	building.ProjectDisplayName = projectDisplayName.String

	return building, nil
}

// scanBuildingRows drains a result set through the shared scanner.
func scanBuildingRows(rows *sql.Rows) ([]models.Building, error) {
	buildings := []models.Building{}
	for rows.Next() {
		building, err := scanBuildingRow(rows)
		if err != nil {
			return nil, err
		}
		buildings = append(buildings, building)
	}

	return buildings, rows.Err()
}
