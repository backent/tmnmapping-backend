package buildingproject

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
)

// rowError is a rejection carrying the column it belongs to, so the operator is
// pointed at a cell rather than a row.
type rowError struct {
	column  string
	value   string
	message string
}

// headerFor maps a column key back to the header the operator sees, so an error names
// what is printed in the sheet rather than the internal key.
func headerFor(key string) string {
	for _, column := range BuildingProjectColumns {
		if column.Key == key {
			return column.Header
		}
	}

	return key
}

// parseRow turns one spreadsheet row into a project, collecting every problem rather
// than stopping at the first: someone fixing a file wants the whole list in one pass.
func parseRow(row []string, colMap map[string]int) (models.BuildingProject, []rowError) {
	var errs []rowError

	text := func(key string) string {
		return spreadsheets.ColValue(row, colMap, key)
	}

	vocabulary := func(key string) string {
		raw := text(key)
		canonicalised, ok := CanonicalVocabularyValue(key, raw)
		if !ok {
			errs = append(errs, rowError{headerFor(key), raw, VocabularyError(key, raw)})

			return ""
		}

		return canonicalised
	}

	// A count is optional, but a count that is present and not a whole number is a
	// mistake worth naming -- silently storing zero would look like a real answer.
	number := func(key string) int {
		raw := text(key)
		if raw == "" {
			return 0
		}

		// Spreadsheets hand back "21" as "21.00" often enough to be worth allowing.
		raw = strings.TrimSuffix(strings.TrimSuffix(raw, ".0"), ".00")
		raw = strings.ReplaceAll(raw, ",", "")

		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			errs = append(errs, rowError{headerFor(key), text(key), headerFor(key) + " must be a whole number"})

			return 0
		}

		return parsed
	}

	money := func(key string) int64 {
		raw := text(key)
		if raw == "" {
			return 0
		}

		cleaned := strings.NewReplacer(",", "", ".", "", " ", "", "Rp", "", "rp", "").Replace(raw)
		parsed, err := strconv.ParseInt(cleaned, 10, 64)
		if err != nil || parsed < 0 {
			errs = append(errs, rowError{headerFor(key), raw,
				headerFor(key) + " must be whole rupiah, with no currency symbol"})

			return 0
		}

		return parsed
	}

	// Dates are accepted in the two shapes a spreadsheet actually produces, and
	// stored in one. Rejecting DD/MM/YYYY would send people back to reformat a file
	// that is otherwise correct.
	date := func(key string) string {
		raw := text(key)
		if raw == "" {
			return ""
		}

		for _, layout := range []string{"2006-01-02", "02/01/2006", "2/1/2006", "01/02/2006 15:04:05", "2006-01-02 15:04:05"} {
			if parsed, err := time.Parse(layout, raw); err == nil {
				return parsed.Format("2006-01-02")
			}
		}

		errs = append(errs, rowError{headerFor(key), raw, headerFor(key) + " must be a date, written as YYYY-MM-DD"})

		return ""
	}

	project := models.BuildingProject{
		ProjectIdIris:    text("project_id_iris"),
		Name:             text("name"),
		BuildingType:     text("building_type"),
		Grade:            text("grade"),
		Pic:              text("pic"),
		TmnProjectStatus: vocabulary("tmn_project_status"),
		NoOfTower:        number("no_of_tower"),
		NoOfScreen:       number("no_of_screen"),
		CreatedDate:      date("created_date"),
		Remark:           text("remark"),
		ContractType:     vocabulary("contract_type"),
		ContractNo:       text("contract_no"),
		ContractDate:     date("contract_date"),
		ContractStart:    date("contract_start"),
		ContractEnd:      date("contract_end"),
		PeriodMonth:      number("period_month"),
		AnnualRental:     money("annual_rental"),
		PaymentTerm:      vocabulary("payment_term"),
		CompanyName:      text("company_name"),
		Exclusivity:      vocabulary("exclusivity"),
		DocType:          vocabulary("doc_type"),
		ContractStatus:   vocabulary("contract_status"),
		CancelledAt:      date("cancelled_at"),
		CancelLastStatus: text("cancel_last_status"),
		CancelReason:     text("cancel_reason"),
	}

	if project.ProjectIdIris == "" {
		errs = append(errs, rowError{headerFor("project_id_iris"), "", "Project ID IRIS is required"})
	}
	if project.Name == "" {
		errs = append(errs, rowError{headerFor("name"), "", "Project Name is required"})
	}

	// The database enforces this too, but as a 500 naming a constraint. Catching it
	// here points at the two cells involved.
	if project.ContractStart != "" && project.ContractEnd != "" && project.ContractEnd < project.ContractStart {
		errs = append(errs, rowError{headerFor("contract_end"), project.ContractEnd,
			"Contract End is before Contract Start"})
	}

	return project, errs
}

// isBlankRow reports a row with nothing in it. Spreadsheets carry trailing empties
// and a blank line between sections; neither is a project.
func isBlankRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}

	return true
}

// clearedFields lists the fields this row will blank, with what they hold today.
//
// Blank means clear on this import, so clearing is its destructive half. The preview
// has to name each one: "120 projects updated" reads as routine, while "120 updated,
// 47 fields cleared" is a thing an operator stops and looks at.
func clearedFields(before models.BuildingProject, after models.BuildingProject) []struct {
	Field string
	Old   string
} {
	oldValues := fieldValues(before)
	newValues := fieldValues(after)

	cleared := []struct {
		Field string
		Old   string
	}{}

	for _, field := range trackedFields {
		if oldValues[field] != "" && newValues[field] == "" {
			cleared = append(cleared, struct {
				Field string
				Old   string
			}{Field: field, Old: oldValues[field]})
		}
	}

	return cleared
}

// addRowErrors files every problem found on one row against that row.
func addRowErrors(result *web.ImportResult, rowNumber int, errs []rowError) {
	for _, err := range errs {
		result.AddError(rowNumber, err.column, err.value, err.message)
	}
}

// newBatchId groups every change row written by one upload, so the history reads as
// "one import by Dara" rather than as hundreds of unrelated edits.
func newBatchId() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		// A batch id that cannot be generated must not silently become empty: the
		// database rejects an import row without one, which is the correct outcome.
		panic(err)
	}

	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:16])
}
