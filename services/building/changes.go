package building

import (
	"strconv"

	"github.com/malikabdulaziz/tmn-backend/models"
)

// Actor is the caller, resolved by RequireAuth. Both id and role are stored on every
// change row; the role is recorded AS IT WAS, because roles change and an audit trail
// that changes with them is not an audit trail.
type Actor struct {
	UserId int
	Role   string
}

// trackedFields lists the columns a change is recorded for, in the order a reader
// wants them: identity, location, classification, status, audience, then the three
// local columns.
//
// Deliberately excluded:
//   - images and synced_at, written by the ERP photo sync on every cycle. Logging
//     them would bury every human edit under thousands of machine rows.
//   - lcd_presence_status, derived from building_status and the competitor flags.
//     Logging a derivation as though someone chose it would mislead; the three
//     columns it comes from are tracked instead.
//   - location, rebuilt from latitude and longitude, which are tracked.
//
// TestTrackedFieldsCoversTheImport keeps this in step with the spreadsheet columns.
var trackedFields = []string{
	"external_building_id",
	"iris_code",
	"name",
	"project_name",
	"project_id_iris",
	"latitude",
	"longitude",
	"subdistrict",
	"citytown",
	"province",
	"cbd_area",
	"building_type",
	"grade_resource",
	"completion_year",
	"building_status",
	"competitor_presence",
	"competitor_exclusive",
	"audience",
	"impression",
	"sellable",
	"connectivity",
	"resource_type",
}

// fieldValues flattens a building into the string form stored in the change log.
// Everything compares as a string because that is what old_value / new_value hold;
// comparing typed values first would mean two notions of equality to keep in step.
func fieldValues(b models.Building) map[string]string {
	return map[string]string{
		"external_building_id": b.ExternalBuildingId,
		"iris_code":            b.IrisCode,
		"name":                 b.Name,
		"project_name":         b.ProjectName,
		// The CODE, not the id: "PRJ-0001 -> PRJ-0002" is readable in a history
		// panel and "7 -> 8" is not.
		"project_id_iris":      b.ProjectIdIris,
		"latitude":             floatToLog(b.Latitude),
		"longitude":            floatToLog(b.Longitude),
		"subdistrict":          b.Subdistrict,
		"citytown":             b.Citytown,
		"province":             b.Province,
		"cbd_area":             b.CbdArea,
		"building_type":        b.BuildingType,
		"grade_resource":       b.GradeResource,
		"completion_year":      intToLog(b.CompletionYear),
		"building_status":      b.BuildingStatus,
		"competitor_presence":  boolToLog(b.CompetitorPresence),
		"competitor_exclusive": boolToLog(b.CompetitorExclusive),
		"audience":             intToLog(b.Audience),
		"impression":           intToLog(b.Impression),
		"sellable":             b.Sellable,
		"connectivity":         b.Connectivity,
		"resource_type":        b.ResourceType,
	}
}

// intToLog renders 0 as empty. The columns are nullable and the model has no null
// int, so logging "0" would report a change every time an empty field was saved
// untouched.
func intToLog(v int) string {
	if v == 0 {
		return ""
	}

	return strconv.Itoa(v)
}

// floatToLog trims trailing zeros so 106.80 and 106.8 are the same coordinate.
// A spreadsheet writes both, and they must not read as a move.
func floatToLog(v float64) string {
	if v == 0 {
		return ""
	}

	return strconv.FormatFloat(v, 'f', -1, 64)
}

// boolToLog writes yes/no, matching what the spreadsheet column holds, so the log
// reads in the same words as the file that produced it.
func boolToLog(v bool) string {
	if v {
		return "yes"
	}

	return "no"
}

// DiffBuildings returns one change row per field that actually differs. Unchanged
// fields produce nothing, which is what keeps the log readable and what makes
// re-importing an untouched export a no-op.
func DiffBuildings(before models.Building, after models.Building, actor Actor, source string, batchId string) []models.BuildingChange {
	oldValues := fieldValues(before)
	newValues := fieldValues(after)

	// The key can itself be edited, so the row is filed under where it ended up.
	externalId := after.ExternalBuildingId
	if externalId == "" {
		externalId = before.ExternalBuildingId
	}

	name := after.Name
	if name == "" {
		name = before.Name
	}

	changes := make([]models.BuildingChange, 0, len(trackedFields))
	for _, field := range trackedFields {
		if oldValues[field] == newValues[field] {
			continue
		}

		changes = append(changes, models.BuildingChange{
			BuildingId:         after.Id,
			ExternalBuildingId: externalId,
			BuildingName:       name,
			ActorUserId:        actor.UserId,
			ActorRole:          actor.Role,
			Action:             models.BuildingActionUpdated,
			Source:             source,
			BatchId:            batchId,
			Field:              field,
			OldValue:           oldValues[field],
			NewValue:           newValues[field],
		})
	}

	return changes
}

// CreationChange is the single row recording that a building came into existence.
// One row, not one per populated field: at creation every field is "new", so a
// per-field log would bury the one fact that matters.
func CreationChange(created models.Building, actor Actor, source string, batchId string) models.BuildingChange {
	return models.BuildingChange{
		BuildingId:         created.Id,
		ExternalBuildingId: created.ExternalBuildingId,
		BuildingName:       created.Name,
		ActorUserId:        actor.UserId,
		ActorRole:          actor.Role,
		Action:             models.BuildingActionCreated,
		Source:             source,
		BatchId:            batchId,
		NewValue:           created.Name,
	}
}

// ClearedFields lists the fields a replacement will blank, with what they hold today.
//
// Blank means clear on the import, so clearing is its destructive half. The preview
// has to name each one: "120 buildings updated" reads as routine, while
// "120 updated, 47 fields cleared" is something an operator stops at.
func ClearedFields(before models.Building, after models.Building) []ClearedField {
	oldValues := fieldValues(before)
	newValues := fieldValues(after)

	cleared := []ClearedField{}
	for _, field := range trackedFields {
		if oldValues[field] != "" && newValues[field] == "" {
			cleared = append(cleared, ClearedField{Field: field, Old: oldValues[field]})
		}
	}

	return cleared
}

// ClearedField is one value an upload is about to empty.
type ClearedField struct {
	Field string
	Old   string
}
