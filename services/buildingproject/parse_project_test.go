package buildingproject

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
)

// rowFrom builds a spreadsheet row in template order, so a test reads like the sheet.
func rowFrom(values map[string]string) ([]string, map[string]int) {
	row := make([]string, len(BuildingProjectColumns))
	colMap := map[string]int{}
	for i, column := range BuildingProjectColumns {
		colMap[column.Key] = i
		row[i] = values[column.Key]
	}

	return row, colMap
}

func TestParseRow_ReadsAFullRow(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"project_id_iris": "PRJ-0001", "name": "Gading Resort Residence",
		"building_type": "Apartment", "grade": "Grade A", "pic": "Dara",
		"tmn_project_status": "Active", "no_of_tower": "31", "no_of_screen": "180",
		"created_date": "2026-01-05", "contract_type": "Initial", "contract_no": "CON/1",
		"contract_start": "2026-02-01", "contract_end": "2027-01-31", "period_month": "12",
		"annual_rental": "24000000", "payment_term": "Monthly", "company_name": "PT Arunika",
		"exclusivity": "Non-Exclusive", "doc_type": "PKS", "contract_status": "Signed",
	})

	project, errs := parseRow(row, colMap)

	assert.Empty(t, errs)
	assert.Equal(t, "PRJ-0001", project.ProjectIdIris)
	assert.Equal(t, 31, project.NoOfTower)
	assert.Equal(t, int64(24000000), project.AnnualRental)
	assert.Equal(t, "2026-02-01", project.ContractStart)
}

// A sheet filled in by a person is not tidy. Accepting these costs nothing and
// rejecting them would send a correct file back for reformatting.
func TestParseRow_AcceptsTheShapesASpreadsheetProduces(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"project_id_iris": "PRJ-1", "name": "X",
		"no_of_screen":   "180.00",
		"no_of_tower":    "1,000",
		"annual_rental":  "Rp 24.000.000",
		"created_date":   "05/01/2026",
		"contract_start": "2026-02-01 00:00:00",
		"doc_type":       "  pks  ",
	})

	project, errs := parseRow(row, colMap)

	assert.Empty(t, errs)
	assert.Equal(t, 180, project.NoOfScreen)
	assert.Equal(t, 1000, project.NoOfTower)
	assert.Equal(t, int64(24000000), project.AnnualRental)
	assert.Equal(t, "2026-01-05", project.CreatedDate)
	assert.Equal(t, "2026-02-01", project.ContractStart)
	assert.Equal(t, "PKS", project.DocType, "stored canonically whatever the casing")
}

// Every problem on a row is collected, not just the first: someone fixing a file
// wants the whole list in one pass.
func TestParseRow_CollectsEveryProblem(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"no_of_screen": "many", "doc_type": "Handshake", "contract_date": "last tuesday",
	})

	_, errs := parseRow(row, colMap)

	messages := map[string]bool{}
	for _, err := range errs {
		messages[err.column] = true
	}

	assert.True(t, messages["Project ID IRIS"], "missing key")
	assert.True(t, messages["Project Name"], "missing name")
	assert.True(t, messages["No of Screen"], "not a number")
	assert.True(t, messages["Doc. Type"], "outside its vocabulary")
	assert.True(t, messages["Contract Date"], "not a date")
	assert.Len(t, errs, 5)
}

func TestParseRow_RejectsEndBeforeStart(t *testing.T) {
	row, colMap := rowFrom(map[string]string{
		"project_id_iris": "PRJ-1", "name": "X",
		"contract_start": "2027-01-01", "contract_end": "2026-01-01",
	})

	_, errs := parseRow(row, colMap)

	assert.Len(t, errs, 1)
	assert.Equal(t, "Contract End", errs[0].column)
}

// Blank means clear, so a preview must be able to name what it is about to blank.
func TestClearedFields_NamesWhatWillBeBlanked(t *testing.T) {
	before := models.BuildingProject{
		ProjectIdIris: "PRJ-1", Name: "X",
		CompanyName: "PT Arunika", Pic: "Dara", AnnualRental: 24000000,
	}
	after := models.BuildingProject{ProjectIdIris: "PRJ-1", Name: "X"}

	cleared := clearedFields(before, after)

	fields := map[string]string{}
	for _, item := range cleared {
		fields[item.Field] = item.Old
	}

	assert.Equal(t, "PT Arunika", fields["company_name"])
	assert.Equal(t, "Dara", fields["pic"])
	assert.Equal(t, "24000000", fields["annual_rental"])
	assert.Len(t, cleared, 3)
}

// Filling a previously empty field is not a clear -- only the reverse is.
func TestClearedFields_FillingIsNotClearing(t *testing.T) {
	before := models.BuildingProject{ProjectIdIris: "PRJ-1", Name: "X"}
	after := models.BuildingProject{ProjectIdIris: "PRJ-1", Name: "X", Pic: "Dara"}

	assert.Empty(t, clearedFields(before, after))
}

func TestIsBlankRow(t *testing.T) {
	assert.True(t, isBlankRow([]string{"", "  ", ""}))
	assert.False(t, isBlankRow([]string{"", "PRJ-1", ""}))
}

// Every import row must carry a batch id -- the database rejects one without it --
// and two uploads must never share one.
func TestNewBatchId_IsAUniqueUUID(t *testing.T) {
	first := newBatchId()
	second := newBatchId()

	assert.Len(t, first, 36)
	assert.NotEqual(t, first, second)
	assert.Equal(t, "4", string(first[14]), "version 4")
}

// The round trip the whole editing workflow rests on: export, re-import unchanged,
// and nothing should be seen as a change. A formatting bug in either direction --
// dates, separators, empty versus zero -- shows up here instead of in front of a user.
func TestExportRoundTrip_ReImportingAnUnchangedExportChangesNothing(t *testing.T) {
	original := models.BuildingProject{
		ProjectIdIris: "PRJ-0001", Name: "Gading Resort Residence",
		BuildingType: "Apartment", Grade: "Grade A", Pic: "Dara",
		TmnProjectStatus: "Active", NoOfTower: 31, NoOfScreen: 180,
		CreatedDate: "2026-01-05", Remark: "note",
		ContractType: "Initial", ContractNo: "CON/2026/001",
		ContractDate: "2026-01-18", ContractStart: "2026-02-01", ContractEnd: "2027-01-31",
		PeriodMonth: 12, AnnualRental: 24000000, PaymentTerm: "Monthly",
		CompanyName: "PT Arunika", Exclusivity: "Non-Exclusive", DocType: "PKS",
		ContractStatus: "Signed", CancelledAt: "2026-08-01",
		CancelLastStatus: "Negotiation", CancelReason: "budget",
	}

	// Exactly the row the exporter writes.
	exported := [][]interface{}{{
		original.ProjectIdIris, original.Name, original.BuildingType, original.Grade,
		original.Pic, original.TmnProjectStatus,
		blankIfZero(original.NoOfTower), blankIfZero(original.NoOfScreen),
		original.CreatedDate, original.Remark,
		original.ContractType, original.ContractNo, original.ContractDate,
		original.ContractStart, original.ContractEnd, blankIfZero(original.PeriodMonth),
		blankIfZero64(original.AnnualRental), original.PaymentTerm, original.CompanyName,
		original.Exclusivity, original.DocType, original.ContractStatus,
		original.CancelledAt, original.CancelLastStatus, original.CancelReason,
	}}

	fileBytes, err := spreadsheets.BuildExport(BuildingProjectSheetName, TemplateHeaders(), exported)
	assert.NoError(t, err)

	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, "xlsx")
	assert.NoError(t, err)
	assert.Len(t, rows, 2)

	colMap := spreadsheets.MapHeaderColumns(rows[0], BuildingProjectColumns)
	assert.Empty(t, spreadsheets.MissingRequiredColumns(colMap, BuildingProjectColumns))

	reparsed, errs := parseRow(rows[1], colMap)
	assert.Empty(t, errs)

	changes := DiffProjects(original, reparsed, testActorInternal, models.BuildingProjectSourceImport, "batch")
	assert.Empty(t, changes, "re-importing an untouched export must log no changes")
	assert.Empty(t, clearedFields(original, reparsed), "and must clear nothing")
}

// An empty count must survive the round trip as empty, not become a real zero.
func TestExportRoundTrip_EmptyCountsStayEmpty(t *testing.T) {
	original := models.BuildingProject{ProjectIdIris: "PRJ-2", Name: "Sparse"}

	exported := [][]interface{}{{
		original.ProjectIdIris, original.Name, "", "", "", "",
		blankIfZero(0), blankIfZero(0), "", "", "", "", "", "", "",
		blankIfZero(0), blankIfZero64(0), "", "", "", "", "", "", "", "",
	}}

	fileBytes, err := spreadsheets.BuildExport(BuildingProjectSheetName, TemplateHeaders(), exported)
	assert.NoError(t, err)

	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, "xlsx")
	assert.NoError(t, err)

	colMap := spreadsheets.MapHeaderColumns(rows[0], BuildingProjectColumns)
	reparsed, errs := parseRow(rows[1], colMap)

	assert.Empty(t, errs)
	assert.Equal(t, 0, reparsed.NoOfTower)
	assert.Empty(t, DiffProjects(original, reparsed, testActorInternal, models.BuildingProjectSourceImport, "b"))
}

var testActorInternal = Actor{UserId: 1, Role: models.RoleAdmin}
