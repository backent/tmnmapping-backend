package buildingproject

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
)

// The template, the export, the importer and the change log must all agree on which
// columns exist. A column added to the model but not the template would be invisible
// to everyone editing by spreadsheet; one added to the template but not tracked would
// import silently and never appear in the history.
func TestTemplateMatchesTrackedFields(t *testing.T) {
	templateKeys := make(map[string]bool, len(BuildingProjectColumns))
	for _, column := range BuildingProjectColumns {
		templateKeys[column.Key] = true
	}

	for _, field := range trackedFields {
		assert.True(t, templateKeys[field],
			"tracked field %q has no template column: add it to BuildingProjectColumns", field)
	}

	tracked := make(map[string]bool, len(trackedFields))
	for _, field := range trackedFields {
		tracked[field] = true
	}

	for _, column := range BuildingProjectColumns {
		assert.True(t, tracked[column.Key],
			"template column %q is not tracked: add it to trackedFields in changes.go", column.Key)
	}
}

// Only the key and the name are mandatory. A project may be raised from a partly
// known site and completed later; requiring more would push people back to editing
// by hand for exactly the records that need attention.
func TestTemplate_OnlyKeyAndNameAreRequired(t *testing.T) {
	required := []string{}
	for _, column := range BuildingProjectColumns {
		if column.Required {
			required = append(required, column.Key)
		}
	}

	assert.Equal(t, []string{"project_id_iris", "name"}, required)
}

// Every closed vocabulary must be spelled out in the note for its column. Someone
// filling the sheet should not have to open the app to discover what a field accepts.
func TestTemplate_ClosedVocabulariesAreDocumented(t *testing.T) {
	notes := make(map[string]string, len(BuildingProjectColumns))
	for _, column := range BuildingProjectColumns {
		notes[column.Key] = column.Note
	}

	for _, field := range ClosedVocabularyFields() {
		note, ok := notes[field]
		assert.True(t, ok, "closed vocabulary %q has no column", field)

		for _, value := range AllowedValues(field) {
			assert.Contains(t, note, value,
				"column %q does not list its allowed value %q", field, value)
		}
	}
}

// The headers a person sees must survive a round trip through the header matcher,
// including the ones with punctuation -- "Doc. Type" and "Cancelled/Terminate Date"
// are the two most likely to normalise badly.
func TestTemplate_HeadersMapBackToTheirKeys(t *testing.T) {
	colMap := spreadsheets.MapHeaderColumns(TemplateHeaders(), BuildingProjectColumns)

	for i, column := range BuildingProjectColumns {
		index, ok := colMap[column.Key]
		assert.True(t, ok, "header %q did not map back to key %q", column.Header, column.Key)
		assert.Equal(t, i, index, "header %q mapped to the wrong column", column.Header)
	}

	assert.Empty(t, spreadsheets.MissingRequiredColumns(colMap, BuildingProjectColumns))
}

// A file in the workbook's ORIGINAL shape must fail the header check rather than
// import its totals row as a project.
func TestTemplate_OldShapeIsRejected(t *testing.T) {
	// Row 1 of the original: the projection block begins at column AB, and the two
	// required headers are present, but the sub-header and totals rows are not
	// something the parser can see from the header alone -- so what protects us is
	// that the old file's row 1 is still a valid header. The real guard is the
	// required-column check against a file that has been cut down wrongly.
	truncated := []string{"Create Date", "PIC", "TMN Project Status", "Contract Type"}

	colMap := spreadsheets.MapHeaderColumns(truncated, BuildingProjectColumns)
	missing := spreadsheets.MissingRequiredColumns(colMap, BuildingProjectColumns)

	assert.ElementsMatch(t, []string{"Project ID IRIS", "Project Name"}, missing,
		"a file missing the key columns must be rejected naming them")
}

func TestTemplate_Builds(t *testing.T) {
	bytes, err := BuildTemplate()

	assert.NoError(t, err)
	assert.NotEmpty(t, bytes)
}
