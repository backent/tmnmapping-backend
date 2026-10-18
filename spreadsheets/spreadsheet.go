package spreadsheets

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// SheetColumn describes one column of an import template.
type SheetColumn struct {
	// Key is the normalized lookup name, e.g. "customer_code".
	Key string
	// Header is what the operator sees in row 1, e.g. "Customer Code".
	Header string
	// Required rejects the file when the column is absent.
	Required bool
	// Example fills the sample row of a downloaded template.
	Example string
	// Note explains the column on the template's Instructions sheet.
	Note string
}

// ParseSpreadsheet turns an uploaded xlsx or csv into rows of trimmed strings.
func ParseSpreadsheet(fileBytes []byte, fileType string) ([][]string, error) {
	switch strings.ToLower(fileType) {
	case "xlsx":
		return parseXLSX(fileBytes)
	case "csv":
		return parseCSV(fileBytes)
	default:
		return nil, ErrUnsupportedFileType
	}
}

func parseXLSX(fileBytes []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.GetRows(f.GetSheetName(0))
}

func parseCSV(fileBytes []byte) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(fileBytes))
	reader.FieldsPerRecord = -1

	var rows [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, record)
	}

	return rows, nil
}

// MapHeaderColumns resolves each column by its header text, never by position.
//
// Position cannot be trusted: operators reorder columns, and exported files do not
// always round-trip in the same order they were imported. Matching is
// case-insensitive and ignores surrounding whitespace, underscores and hyphens, so
// "Customer Code", "customer_code" and "CUSTOMER CODE" all resolve alike.
func MapHeaderColumns(header []string, columns []SheetColumn) map[string]int {
	normalized := make(map[string]int, len(header))
	for i, h := range header {
		normalized[normalizeHeader(h)] = i
	}

	colMap := make(map[string]int, len(columns))
	for _, col := range columns {
		if idx, ok := normalized[normalizeHeader(col.Header)]; ok {
			colMap[col.Key] = idx
			continue
		}
		if idx, ok := normalized[normalizeHeader(col.Key)]; ok {
			colMap[col.Key] = idx
		}
	}

	return colMap
}

func normalizeHeader(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", " ")
	value = strings.ReplaceAll(value, "-", " ")

	return strings.Join(strings.Fields(value), " ")
}

// MissingRequiredColumns lists the required headers absent from an uploaded file.
func MissingRequiredColumns(colMap map[string]int, columns []SheetColumn) []string {
	var missing []string
	for _, col := range columns {
		if !col.Required {
			continue
		}
		if _, ok := colMap[col.Key]; !ok {
			missing = append(missing, col.Header)
		}
	}

	return missing
}

// ColValue reads a cell by column key, trimmed. Missing columns and short rows read
// as empty rather than panicking, so a ragged file produces validation errors with
// row numbers instead of a 500.
func ColValue(row []string, colMap map[string]int, key string) string {
	idx, ok := colMap[key]
	if !ok || idx >= len(row) {
		return ""
	}

	return strings.TrimSpace(row[idx])
}

// BuildTemplate produces an empty import template: a header row, one example row,
// and a second sheet explaining every column.
func BuildTemplate(sheetName string, columns []SheetColumn) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	first := f.GetSheetName(0)
	if err := f.SetSheetName(first, sheetName); err != nil {
		return nil, err
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E8EEF7"}},
	})
	if err != nil {
		return nil, err
	}

	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, col.Header); err != nil {
			return nil, err
		}
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
		_ = f.SetColWidth(sheetName, columnLetter(i+1), columnLetter(i+1), 22)

		exampleCell, _ := excelize.CoordinatesToCellName(i+1, 2)
		if err := f.SetCellValue(sheetName, exampleCell, col.Example); err != nil {
			return nil, err
		}
	}

	if err := writeInstructions(f, sheetName, columns); err != nil {
		return nil, err
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeInstructions(f *excelize.File, sheetName string, columns []SheetColumn) error {
	const instructions = "Instructions"
	if _, err := f.NewSheet(instructions); err != nil {
		return err
	}

	rows := [][]interface{}{
		{"Column", "Required", "Notes"},
	}
	for _, col := range columns {
		required := "Optional"
		if col.Required {
			required = "Required"
		}
		rows = append(rows, []interface{}{col.Header, required, col.Note})
	}
	rows = append(rows,
		[]interface{}{"", "", ""},
		[]interface{}{"Row 2 of the " + sheetName + " sheet is an example.", "", "Delete it before uploading."},
		[]interface{}{"Column order does not matter.", "", "Columns are matched by their header text."},
		[]interface{}{"Nothing is saved if any row fails validation.", "", "Fix the reported rows and upload again."},
	)

	for i, row := range rows {
		for j, value := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
			if err := f.SetCellValue(instructions, cell, value); err != nil {
				return err
			}
		}
	}

	_ = f.SetColWidth(instructions, "A", "A", 26)
	_ = f.SetColWidth(instructions, "B", "B", 12)
	_ = f.SetColWidth(instructions, "C", "C", 80)

	return nil
}

// BuildExport writes a header row followed by data rows.
func BuildExport(sheetName string, headers []string, rows [][]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	first := f.GetSheetName(0)
	if err := f.SetSheetName(first, sheetName); err != nil {
		return nil, err
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, err
		}
		_ = f.SetColWidth(sheetName, columnLetter(i+1), columnLetter(i+1), 22)
	}

	for rowIdx, row := range rows {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			if err := f.SetCellValue(sheetName, cell, value); err != nil {
				return nil, err
			}
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func columnLetter(index int) string {
	name, err := excelize.ColumnNumberToName(index)
	if err != nil {
		return "A"
	}

	return name
}
