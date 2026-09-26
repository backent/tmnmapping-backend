package spreadsheets_test

import (
	"testing"

	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var columns = []spreadsheets.SheetColumn{
	{Key: "code", Header: "Customer Code", Required: true, Example: "CUST-001"},
	{Key: "name", Header: "Customer Name", Required: true, Example: "Acme"},
	{Key: "industry", Header: "Industry", Required: false, Example: "Retail"},
}

// Operators reorder and re-case columns constantly; matching must survive it.
func TestMapHeaderColumns_IgnoresOrderCaseAndSeparators(t *testing.T) {
	header := []string{"  INDUSTRY ", "customer_name", "Customer-Code"}

	colMap := spreadsheets.MapHeaderColumns(header, columns)

	assert.Equal(t, 2, colMap["code"])
	assert.Equal(t, 1, colMap["name"])
	assert.Equal(t, 0, colMap["industry"])
	assert.Empty(t, spreadsheets.MissingRequiredColumns(colMap, columns))
}

func TestMissingRequiredColumns_ReportsHeadersNotKeys(t *testing.T) {
	colMap := spreadsheets.MapHeaderColumns([]string{"Customer Code"}, columns)

	assert.Equal(t, []string{"Customer Name"}, spreadsheets.MissingRequiredColumns(colMap, columns))
}

// A ragged row must read as empty rather than panic, so the importer can report it
// with a row number instead of returning a 500.
func TestColValue_HandlesShortRowsAndMissingColumns(t *testing.T) {
	colMap := map[string]int{"code": 0, "name": 1}

	assert.Equal(t, "CUST-001", spreadsheets.ColValue([]string{" CUST-001 "}, colMap, "code"))
	assert.Equal(t, "", spreadsheets.ColValue([]string{"CUST-001"}, colMap, "name"))
	assert.Equal(t, "", spreadsheets.ColValue([]string{"CUST-001"}, colMap, "absent"))
}

func TestParseSpreadsheet_CSV(t *testing.T) {
	rows, err := spreadsheets.ParseSpreadsheet([]byte("a,b\n1,2\n"), "csv")

	require.NoError(t, err)
	assert.Equal(t, [][]string{{"a", "b"}, {"1", "2"}}, rows)
}

func TestParseSpreadsheet_RejectsOtherTypes(t *testing.T) {
	_, err := spreadsheets.ParseSpreadsheet([]byte("{}"), "json")

	assert.ErrorIs(t, err, spreadsheets.ErrUnsupportedFileType)
}

// The template must be readable by the parser that consumes uploads, or the file we
// hand the operator would fail on the way back in.
func TestBuildTemplate_IsParseableAndCarriesEveryColumn(t *testing.T) {
	fileBytes, err := spreadsheets.BuildTemplate("Customers", columns)
	require.NoError(t, err)

	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, "xlsx")
	require.NoError(t, err)
	require.Len(t, rows, 2, "a header row and one example row")

	colMap := spreadsheets.MapHeaderColumns(rows[0], columns)
	assert.Empty(t, spreadsheets.MissingRequiredColumns(colMap, columns))
	assert.Equal(t, "CUST-001", spreadsheets.ColValue(rows[1], colMap, "code"))
}

// Export reuses the template headers so a downloaded file can be edited and
// uploaded straight back.
func TestBuildExport_RoundTripsThroughTheParser(t *testing.T) {
	headers := []string{"Customer Code", "Customer Name", "Industry"}
	data := [][]interface{}{{"CUST-001", "Acme", "Retail"}}

	fileBytes, err := spreadsheets.BuildExport("Customers", headers, data)
	require.NoError(t, err)

	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, "xlsx")
	require.NoError(t, err)

	colMap := spreadsheets.MapHeaderColumns(rows[0], columns)
	assert.Empty(t, spreadsheets.MissingRequiredColumns(colMap, columns))
	assert.Equal(t, "Acme", spreadsheets.ColValue(rows[1], colMap, "name"))
}
