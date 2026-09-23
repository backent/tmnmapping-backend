package building

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
)

func rowFrom(values map[string]string) ([]string, map[string]int) {
	row := make([]string, len(BuildingColumns))
	colMap := map[string]int{}
	for i, column := range BuildingColumns {
		colMap[column.Key] = i
		row[i] = values[column.Key]
	}

	return row, colMap
}

var importActor = Actor{UserId: 1, Role: models.RoleAdmin}

func TestParseRow_ReadsAFullRow(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"external_building_id": "BLDG-1", "name": "Gading Tower 1", "iris_code": "B000006",
		"latitude": "-6.2088", "longitude": "106.8456", "citytown": "Jakarta Utara",
		"building_type": "Apartment", "completion_year": "2019",
		"building_status": "BAST Signed", "competitor_presence": "no",
		"competitor_exclusive": "no", "audience": "4200", "sellable": "sell",
	})

	building, _, errs, _ := parseRow(row, colMap)

	assert.Empty(t, errs)
	assert.Equal(t, "BLDG-1", building.ExternalBuildingId)
	assert.InDelta(t, -6.2088, building.Latitude, 0.00001)
	assert.Equal(t, 2019, building.CompletionYear)
	assert.Equal(t, "sell", building.Sellable)
}

// LCD presence is DERIVED, never read from the file. One source of truth means the
// sheet cannot disagree with the columns it is calculated from.
func TestParseRow_DerivesLcdPresence(t *testing.T) {
	tests := []struct {
		status, presence, exclusive, want string
	}{
		{"BAST Signed", "no", "no", "TMN"},
		{"BAST Signed", "yes", "no", "CoExist"},
		{"BAST Signed", "no", "yes", ""},
		{"Surveyed", "yes", "no", "Competitor"},
		{"Surveyed", "no", "no", "Opportunity"},
	}

	for _, test := range tests {
		row, colMap := rowFrom(map[string]string{
			"external_building_id": "B", "name": "N", "building_status": test.status,
			"competitor_presence": test.presence, "competitor_exclusive": test.exclusive,
		})

		building, _, errs, _ := parseRow(row, colMap)

		assert.Empty(t, errs)
		assert.Equal(t, test.want, building.LcdPresenceStatus,
			"%s / presence %s / exclusive %s", test.status, test.presence, test.exclusive)
	}
}

// "BAST Signed" must survive whatever casing a person types, or the building
// silently stops counting as TMN's.
func TestParseRow_BastSignedSurvivesCasing(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"external_building_id": "B", "name": "N", "building_status": "  bast signed  ",
		"competitor_presence": "no", "competitor_exclusive": "no",
	})

	building, _, errs, _ := parseRow(row, colMap)

	assert.Empty(t, errs)
	assert.Equal(t, "BAST Signed", building.BuildingStatus, "stored canonically")
	assert.Equal(t, "TMN", building.LcdPresenceStatus)
}

func TestParseRow_CollectsEveryProblem(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"latitude": "north", "audience": "many",
		"sellable": "maybe",
	})

	_, _, errs, _ := parseRow(row, colMap)

	columns := map[string]bool{}
	for _, err := range errs {
		columns[err.column] = true
	}

	assert.True(t, columns["Building Code"])
	assert.True(t, columns["Building Name"])
	assert.True(t, columns["Latitude"])
	assert.True(t, columns["Audience"])
	assert.True(t, columns["Sellable"])
}

// A coordinate out of range would put the building somewhere it is not, which is
// worse than leaving it off the map.
func TestParseRow_RejectsImpossibleCoordinates(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"external_building_id": "B", "name": "N", "latitude": "200", "longitude": "500",
	})

	_, _, errs, _ := parseRow(row, colMap)

	assert.Len(t, errs, 2)
}

func TestClearedFields_NamesWhatWillBeBlanked(t *testing.T) {
	before := models.Building{ExternalBuildingId: "B", Name: "N", Citytown: "Jakarta", Audience: 4200}
	after := models.Building{ExternalBuildingId: "B", Name: "N"}

	cleared := ClearedFields(before, after)

	fields := map[string]string{}
	for _, item := range cleared {
		fields[item.Field] = item.Old
	}

	assert.Equal(t, "Jakarta", fields["citytown"])
	assert.Equal(t, "4200", fields["audience"])
}

// A coordinate written as 106.80 and 106.8 is the same place. Logging it as a move
// would fill the history with edits nobody made.
func TestDiffBuildings_CoordinateFormattingIsNotAChange(t *testing.T) {
	before := models.Building{ExternalBuildingId: "B", Name: "N", Longitude: 106.80}
	after := models.Building{ExternalBuildingId: "B", Name: "N", Longitude: 106.8}

	assert.Empty(t, DiffBuildings(before, after, importActor, models.BuildingSourceImport, "b"))
}

func TestDiffBuildings_NoChangesLogsNothing(t *testing.T) {
	b := models.Building{
		ExternalBuildingId: "B", Name: "N", Citytown: "Jakarta",
		BuildingStatus: "BAST Signed", Audience: 4200, Latitude: -6.2088,
	}

	assert.Empty(t, DiffBuildings(b, b, importActor, models.BuildingSourceImport, "b"))
}

func TestDiffBuildings_RecordsTheOldValue(t *testing.T) {
	before := models.Building{ExternalBuildingId: "B", Name: "N", Audience: 4200}
	after := models.Building{ExternalBuildingId: "B", Name: "N", Audience: 5000}

	changes := DiffBuildings(before, after, importActor, models.BuildingSourceForm, "")

	assert.Len(t, changes, 1)
	assert.Equal(t, "audience", changes[0].Field)
	assert.Equal(t, "4200", changes[0].OldValue)
	assert.Equal(t, "5000", changes[0].NewValue)
	assert.Empty(t, changes[0].BatchId, "a form edit carries no batch id")
}

// Photos and synced_at are written by the ERP sync on every cycle. Tracking them
// would bury every human edit under thousands of machine rows.
func TestDiffBuildings_IgnoresTheSyncedColumns(t *testing.T) {
	before := models.Building{ExternalBuildingId: "B", Name: "N",
		Images: []models.BuildingImage{{Name: "front", Path: "/a"}}, SyncedAt: "2026-01-01"}
	after := models.Building{ExternalBuildingId: "B", Name: "N",
		Images: []models.BuildingImage{{Name: "front", Path: "/b"}}, SyncedAt: "2026-09-23"}

	assert.Empty(t, DiffBuildings(before, after, importActor, models.BuildingSourceSync, ""))
}

// Every column the spreadsheet offers must be tracked, or edits to it go unrecorded
// -- and the change log is the only undo trail a destructive upload has.
func TestTrackedFieldsCoversTheImport(t *testing.T) {
	tracked := map[string]bool{}
	for _, field := range trackedFields {
		tracked[field] = true
	}

	for _, column := range BuildingColumns {
		// project_id_iris is resolved to a link, not stored on the building row.
		if column.Key == "project_id_iris" {
			continue
		}

		assert.True(t, tracked[column.Key],
			"spreadsheet column %q is not tracked: add it to trackedFields", column.Key)
	}
}

func TestTemplate_HeadersMapBackToTheirKeys(t *testing.T) {
	colMap := spreadsheets.MapHeaderColumns(TemplateHeaders(), BuildingColumns)

	for i, column := range BuildingColumns {
		index, ok := colMap[column.Key]
		assert.True(t, ok, "header %q did not map back", column.Header)
		assert.Equal(t, i, index)
	}

	assert.Empty(t, spreadsheets.MissingRequiredColumns(colMap, BuildingColumns))
}

// The guarantee the whole edit-and-reupload workflow rests on, and the one that makes
// blank-means-clear safe: a re-imported untouched export must change nothing.
func TestExportRoundTrip_ReImportingAnUnchangedExportChangesNothing(t *testing.T) {
	original := models.Building{
		ExternalBuildingId: "BLDG-2025-09-02547", Name: "Gading Resort Residence - Tower 1",
		IrisCode: "B000006", Latitude: -6.2088, Longitude: 106.8456,
		Subdistrict: "Kelapa Gading", Citytown: "Jakarta Utara", Province: "DKI Jakarta",
		CbdArea: "Non-CBD", BuildingType: "Apartment", GradeResource: "Grade A",
		CompletionYear: 2019, BuildingStatus: "BAST Signed",
		CompetitorPresence: false, CompetitorExclusive: false,
		Audience: 4200, Impression: 12600, Sellable: "sell", Connectivity: "online",
		ResourceType: "LCD",
	}
	original.LcdPresenceStatus = calculateLcdPresenceStatus(false, false, "BAST Signed")

	exported := [][]interface{}{{
		original.ExternalBuildingId, original.Name, original.IrisCode, "",
		blankIfZeroFloat(original.Latitude), blankIfZeroFloat(original.Longitude),
		original.Subdistrict, original.Citytown, original.Province, original.CbdArea,
		original.BuildingType, original.GradeResource, blankIfZero(original.CompletionYear),
		original.BuildingStatus, yesNoCell(original.CompetitorPresence), yesNoCell(original.CompetitorExclusive),
		blankIfZero(original.Audience), blankIfZero(original.Impression),
		original.Sellable, original.Connectivity, original.ResourceType,
	}}

	fileBytes, err := spreadsheets.BuildExport(BuildingSheetName, TemplateHeaders(), exported)
	assert.NoError(t, err)

	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, "xlsx")
	assert.NoError(t, err)

	colMap := spreadsheets.MapHeaderColumns(rows[0], BuildingColumns)
	reparsed, _, errs, _ := parseRow(rows[1], colMap)
	assert.Empty(t, errs)

	assert.Empty(t, DiffBuildings(original, reparsed, importActor, models.BuildingSourceImport, "b"),
		"re-importing an untouched export must log no changes")
	assert.Empty(t, ClearedFields(original, reparsed), "and must clear nothing")
	assert.Equal(t, original.LcdPresenceStatus, reparsed.LcdPresenceStatus,
		"the derived status must survive the round trip")
}

func TestExportRoundTrip_EmptyValuesStayEmpty(t *testing.T) {
	original := models.Building{ExternalBuildingId: "BLDG-2", Name: "Sparse"}
	original.LcdPresenceStatus = calculateLcdPresenceStatus(false, false, "")

	exported := [][]interface{}{{
		original.ExternalBuildingId, original.Name, "", "",
		blankIfZeroFloat(0), blankIfZeroFloat(0), "", "", "", "", "", "",
		blankIfZero(0), "", yesNoCell(false), yesNoCell(false),
		blankIfZero(0), blankIfZero(0), "", "", "",
	}}

	fileBytes, err := spreadsheets.BuildExport(BuildingSheetName, TemplateHeaders(), exported)
	assert.NoError(t, err)

	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, "xlsx")
	assert.NoError(t, err)

	colMap := spreadsheets.MapHeaderColumns(rows[0], BuildingColumns)
	reparsed, _, errs, _ := parseRow(rows[1], colMap)

	assert.Empty(t, errs)
	assert.Equal(t, 0, reparsed.Audience)
	assert.Zero(t, reparsed.Latitude, "a blank coordinate is off the map, not at the equator")
	assert.Empty(t, DiffBuildings(original, reparsed, importActor, models.BuildingSourceImport, "b"))
}

// A blank Building Type must stay blank. CanonicalizeBuildingType turns "" into
// "Other", which is right for the ERP sync -- every ERP building has a type -- and
// wrong here, where a blank cell means "clear this". Inventing "Other" would also
// make a re-imported export look edited.
func TestParseRow_BlankBuildingTypeStaysBlank(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"external_building_id": "B", "name": "N", "building_type": "   ",
	})

	building, _, errs, _ := parseRow(row, colMap)

	assert.Empty(t, errs)
	assert.Equal(t, "", building.BuildingType)
}

// An unrecognised type is collapsed, not rejected: the map's filter chips are built
// from the canonical list, so a stray value would create a chip nobody asked for.
func TestParseRow_UnknownBuildingTypeCollapsesToOther(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"external_building_id": "B", "name": "N", "building_type": "Submarine",
	})

	building, _, errs, _ := parseRow(row, colMap)

	assert.Empty(t, errs)
	assert.Equal(t, "Other", building.BuildingType)
}

// ERP adds statuses without warning: "Building Onboarded" appeared on 9 live
// buildings and was unknown to this application until a real export was re-imported.
// Rejecting a row for a value ERP itself produced would block a legitimate file, so
// an unknown status is accepted, stored as written, and flagged.
func TestParseRow_UnknownStatusIsAcceptedAndFlagged(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"external_building_id": "B", "name": "N", "building_status": "Building Onboarded",
	})

	building, _, errs, warnings := parseRow(row, colMap)

	assert.Empty(t, errs, "an unknown status must not reject the row")
	assert.Len(t, warnings, 1)
	assert.Equal(t, "Building Onboarded", building.BuildingStatus, "stored as written")
	assert.Contains(t, warnings[0].message, "not a status the map filters on")

	// Not BAST Signed, so it is not TMN's.
	assert.Equal(t, "Opportunity", building.LcdPresenceStatus)
}

func TestParseRow_KnownStatusIsCanonicalisedWithoutWarning(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"external_building_id": "B", "name": "N", "building_status": "  surveyed ",
	})

	building, _, errs, warnings := parseRow(row, colMap)

	assert.Empty(t, errs)
	assert.Empty(t, warnings)
	assert.Equal(t, "Surveyed", building.BuildingStatus)
}

// Excel writes TRUE/FALSE for a checkbox and 1/0 for a formula. None of those should
// reject a row, and none should silently become "no" -- a wrong competitor flag
// changes the building's LCD presence.
func TestParseRow_ReadsEveryBooleanSpelling(t *testing.T) {
	for _, yes := range []string{"yes", "Yes", "YES", "y", "true", "TRUE", "1"} {
		row, colMap := rowFrom(map[string]string{
			"external_building_id": "B", "name": "N", "competitor_presence": yes,
		})

		building, _, errs, _ := parseRow(row, colMap)

		assert.Empty(t, errs, "%q should be read as yes", yes)
		assert.True(t, building.CompetitorPresence, "%q should be read as yes", yes)
	}

	for _, no := range []string{"no", "N", "false", "FALSE", "0", ""} {
		row, colMap := rowFrom(map[string]string{
			"external_building_id": "B", "name": "N", "competitor_presence": no,
		})

		building, _, errs, _ := parseRow(row, colMap)

		assert.Empty(t, errs, "%q should be read as no", no)
		assert.False(t, building.CompetitorPresence, "%q should be read as no", no)
	}
}

func TestParseRow_RejectsAnUnreadableBoolean(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"external_building_id": "B", "name": "N", "competitor_presence": "maybe",
	})

	_, _, errs, _ := parseRow(row, colMap)

	assert.Len(t, errs, 1)
	assert.Equal(t, "Competitor Presence", errs[0].column)
}
