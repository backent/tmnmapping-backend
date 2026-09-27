package building

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"

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

// parseRow turns one spreadsheet row into a building, collecting every problem
// rather than stopping at the first: someone fixing a file wants the whole list.
//
// It returns the project code separately -- the building model has no field for it,
// and resolving it to a project id needs a database lookup the parser does not do.
func parseRow(row []string, colMap map[string]int) (models.Building, string, []rowError, []rowError) {
	var errs []rowError
	var warnings []rowError

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

	number := func(key string) int {
		raw := text(key)
		if raw == "" {
			return 0
		}

		cleaned := strings.ReplaceAll(strings.TrimSuffix(strings.TrimSuffix(raw, ".0"), ".00"), ",", "")

		parsed, err := strconv.Atoi(cleaned)
		if err != nil || parsed < 0 {
			errs = append(errs, rowError{headerFor(key), raw, headerFor(key) + " must be a whole number"})

			return 0
		}

		return parsed
	}

	// A coordinate is optional, but one that is present and unreadable is a mistake
	// worth naming: storing zero would put the building in the Gulf of Guinea.
	coordinate := func(key string, limit float64) float64 {
		raw := text(key)
		if raw == "" {
			return 0
		}

		parsed, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", "."), 64)
		if err != nil {
			errs = append(errs, rowError{headerFor(key), raw, headerFor(key) + " must be a decimal number"})

			return 0
		}
		if parsed < -limit || parsed > limit {
			errs = append(errs, rowError{headerFor(key), raw,
				fmt.Sprintf("%s must be between -%g and %g", headerFor(key), limit, limit)})

			return 0
		}

		return parsed
	}

	// yes/no arrives as yes, no, TRUE, FALSE, 1 or 0 depending on who wrote the file
	// and which tool produced it.
	yesNo := func(key string) bool {
		raw := text(key)

		canonicalised, ok := CanonicalYesNo(raw)
		if !ok {
			errs = append(errs, rowError{headerFor(key), raw,
				headerFor(key) + " must be yes or no"})

			return false
		}

		return canonicalised == "yes"
	}

	// A status this application does not know is ACCEPTED and flagged. ERP adds
	// statuses without warning, and rejecting a row for a value ERP itself produced
	// would block a legitimate file; only the map's progress filter is affected.
	buildingStatus := func() string {
		raw := text("building_status")

		canonicalised, known := CanonicalBuildingStatus(raw)
		if !known {
			warnings = append(warnings, rowError{headerFor("building_status"), raw,
				"\"" + canonicalised + "\" is not a status the map filters on. It is stored as written."})
		}

		return canonicalised
	}()

	building := models.Building{
		ExternalBuildingId: text("external_building_id"),
		Name:               text("name"),
		IrisCode:           text("iris_code"),
		Latitude:           coordinate("latitude", 90),
		Longitude:          coordinate("longitude", 180),
		Subdistrict:        text("subdistrict"),
		Citytown:           text("citytown"),
		Province:           text("province"),
		CbdArea:            text("cbd_area"),
		// Canonicalised on the way in, as the ERP sync does, so the map filter and
		// the quotation print one spelling rather than five. A BLANK cell is left
		// blank: CanonicalizeBuildingType turns "" into "Other", which is right for
		// the sync (every ERP building has a type) and wrong here, where a blank
		// means "clear this" and inventing "Other" would both contradict that and
		// make a re-imported export look edited.
		BuildingType:        canonicalTypeOrBlank(text("building_type")),
		GradeResource:       text("grade_resource"),
		CompletionYear:      number("completion_year"),
		BuildingStatus:      buildingStatus,
		CompetitorPresence:  yesNo("competitor_presence"),
		CompetitorExclusive: yesNo("competitor_exclusive"),
		Audience:            number("audience"),
		Impression:          number("impression"),
		Sellable:            vocabulary("sellable"),
		Connectivity:        vocabulary("connectivity"),
		ResourceType:        text("resource_type"),
	}

	// competitor_location has never been anything but a duplicate of presence, and
	// the map filters on it. Kept in step here rather than leaving them to diverge.
	building.CompetitorLocation = building.CompetitorPresence

	// Derived, never taken from the file: one source of truth, so the sheet cannot
	// disagree with the status and competitor columns it is calculated from.
	building.LcdPresenceStatus = calculateLcdPresenceStatus(
		building.CompetitorPresence, building.CompetitorExclusive, building.BuildingStatus)

	if building.ExternalBuildingId == "" {
		errs = append(errs, rowError{headerFor("external_building_id"), "", "Building Code is required"})
	}
	if building.Name == "" {
		errs = append(errs, rowError{headerFor("name"), "", "Building Name is required"})
	}

	return building, text("project_id_iris"), errs, warnings
}

// canonicalTypeOrBlank canonicalises a building type, leaving an empty cell empty.
func canonicalTypeOrBlank(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}

	return CanonicalizeBuildingType(raw)
}

// isBlankRow reports a row with nothing in it. Spreadsheets carry trailing empties
// and blank separator lines; neither is a building.
func isBlankRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}

	return true
}

func addRowErrors(result *web.ImportResult, rowNumber int, errs []rowError) {
	for _, err := range errs {
		result.AddError(rowNumber, err.column, err.value, err.message)
	}
}

// newBatchId groups every change row written by one upload, so the history reads as
// "one import by Dara" rather than thousands of unrelated edits.
func newBatchId() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		// A batch id that cannot be generated must not silently become empty: the
		// database rejects an import row without one, which is the right outcome.
		panic(err)
	}

	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:16])
}
