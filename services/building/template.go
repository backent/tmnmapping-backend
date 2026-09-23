package building

import (
	"strings"

	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
)

// BuildingSheetName is the tab the importer reads.
const BuildingSheetName = "Buildings"

// Vocabularies for the columns that have one.
//
// building_status is NOT closed, despite LCD presence being derived from it. ERP
// introduces statuses without warning -- "Building Onboarded" appeared on 9 buildings
// and was unknown to this list until a real export was re-imported on 2026-09-23 --
// and rejecting a row for a value ERP itself produced would block a legitimate file.
// A known status is canonicalised so "bast signed" still reads as "BAST Signed";
// anything else passes through with a notice. Only the progress filter is affected.
//
// sellable and connectivity ARE closed: they come from this application, not ERP.
// The yes/no pair is closed because a mistyped value silently becoming "no" would
// change a building's LCD presence.
//
// Grade stays open. Building type is neither: unknown values collapse to "Other",
// because the map's filter chips are built from a fixed list.
var (
	BuildingStatuses = []string{
		"BAST Signed", "Surveyed", "Building Proposal Approved",
		"Building Information Updated", "Planning", "Cancelled", "Work Order Submitted",
	}
	SellableValues     = []string{"sell", "not_sell"}
	ConnectivityValues = []string{"online", "manual", "not_yet_checked"}
	YesNoValues        = []string{"yes", "no"}
)

var closedVocabularies = map[string][]string{
	"sellable":             SellableValues,
	"connectivity":         ConnectivityValues,
	"competitor_presence":  YesNoValues,
	"competitor_exclusive": YesNoValues,
}

// yesNoAliases are the other spellings a spreadsheet produces for a boolean. Excel
// writes TRUE/FALSE for a checkbox and 1/0 for a formula, and a person writes Y or N;
// none of those should be a rejected row.
var yesNoAliases = map[string]string{
	"yes": "yes", "y": "yes", "true": "yes", "1": "yes",
	"no": "no", "n": "no", "false": "no", "0": "no",
}

// CanonicalYesNo reads whatever a sheet holds for a boolean column.
func CanonicalYesNo(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "no", true
	}

	canonical, ok := yesNoAliases[strings.ToLower(trimmed)]

	return canonical, ok
}

// CanonicalBuildingStatus canonicalises a known status and passes anything else
// through unchanged, reporting whether it was recognised so the caller can warn.
//
// "BAST Signed" must survive whatever casing a person types: LCD presence keys off
// that exact spelling, and a building whose status stops matching silently stops
// counting as TMN's.
func CanonicalBuildingStatus(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", true
	}

	for _, candidate := range BuildingStatuses {
		if strings.EqualFold(candidate, trimmed) {
			return candidate, true
		}
	}

	return trimmed, false
}

// BuildingColumns defines the upload, and is the single source of truth for the
// template, the export and the importer.
//
// Headers match the export produced on 2026-09-20, so a file already being filled in
// by hand still loads. Row 1 is the header, row 2 onward is data.
var BuildingColumns = []spreadsheets.SheetColumn{
	{
		Key: "external_building_id", Header: "Building Code", Required: true,
		Example: "BLDG-2025-09-02547",
		Note:    "The key. Required and unique; it decides whether a row creates a building or updates one. Export first rather than typing it.",
	},
	{
		Key: "name", Header: "Building Name", Required: true,
		Example: "Gading Resort Residence - Tower 1",
		Note:    "Required. One tower, not the whole site -- the site is the Project.",
	},
	{
		Key: "iris_code", Header: "IRIS Code", Required: false,
		Example: "B000006",
		Note:    "Optional, but must be unique when present. The price import matches buildings by this, so a building without one cannot be priced by spreadsheet.",
	},
	{
		Key: "project_id_iris", Header: "Project ID IRIS", Required: false,
		Example: "PRJ-0001",
		Note:    "Links this building to a project. An unknown code creates a stub project rather than rejecting the row, so the two files can be uploaded in any order.",
	},
	{
		Key: "latitude", Header: "Latitude", Required: false,
		Example: "-6.2088",
		Note:    "Decimal degrees. Required for the building to appear on the map. Blank or zero leaves it off the map rather than at the equator.",
	},
	{
		Key: "longitude", Header: "Longitude", Required: false,
		Example: "106.8456",
		Note:    "Decimal degrees, same.",
	},
	{
		Key: "subdistrict", Header: "Subdistrict", Required: false,
		Example: "Kelapa Gading", Note: "Map filter.",
	},
	{
		Key: "citytown", Header: "City / Town", Required: false,
		Example: "Jakarta Utara", Note: "Map filter. Copied onto a quotation when the building is sold.",
	},
	{
		Key: "province", Header: "Province", Required: false,
		Example: "DKI Jakarta", Note: "Map filter.",
	},
	{
		Key: "cbd_area", Header: "CBD Area", Required: false,
		Example: "Non-CBD", Note: "Map filter.",
	},
	{
		Key: "building_type", Header: "Building Type", Required: false,
		Example: "Apartment",
		Note: "One of: " + strings.Join(CanonicalBuildingTypes, ", ") +
			". Anything else is collapsed to Other rather than rejected, because the map's filter chips are built from this list. Printed on the quotation.",
	},
	{
		Key: "grade_resource", Header: "Grade Resource", Required: false,
		Example: "Grade A", Note: "Free text. Map filter.",
	},
	{
		Key: "completion_year", Header: "Completion Year", Required: false,
		Example: "2019", Note: "Whole number. The map's year filter.",
	},
	{
		Key: "building_status", Header: "Building Status", Required: false,
		Example: "BAST Signed",
		Note: "Usually one of: " + strings.Join(BuildingStatuses, ", ") +
			". Another value is accepted and flagged -- only the map's progress filter is affected. IMPORTANT: LCD presence is derived from this plus the two competitor columns, so a building must keep 'BAST Signed' to count as TMN's.",
	},
	{
		Key: "competitor_presence", Header: "Competitor Presence", Required: false,
		Example: "no", Note: "yes or no. TRUE/FALSE and 1/0 are read too. Blank counts as no. Feeds the derived LCD presence.",
	},
	{
		Key: "competitor_exclusive", Header: "Competitor Exclusive", Required: false,
		Example: "no", Note: "yes or no. TRUE/FALSE and 1/0 are read too. Blank counts as no. Feeds the derived LCD presence.",
	},
	{
		Key: "audience", Header: "Audience", Required: false,
		Example: "4200", Note: "Whole number. Traffic through the building.",
	},
	{
		Key: "impression", Header: "Impressions", Required: false,
		Example: "12600", Note: "Whole number. ERP called this audience_projection.",
	},
	{
		Key: "sellable", Header: "Sellable", Required: false,
		Example: "sell", Note: "One of: " + strings.Join(SellableValues, ", ") + ". Empty on every building today.",
	},
	{
		Key: "connectivity", Header: "Connectivity", Required: false,
		Example: "online", Note: "One of: " + strings.Join(ConnectivityValues, ", ") + ". Empty on every building today.",
	},
	{
		Key: "resource_type", Header: "Resource Type", Required: false,
		Example: "", Note: "Free text. Empty on every building today.",
	},
}

// CanonicalVocabularyValue matches loosely and returns the canonical spelling, so a
// sheet written by a person holding "Sell" or " online " still loads while what is
// STORED stays exact -- the filters compare strings.
func CanonicalVocabularyValue(field string, value string) (string, bool) {
	allowed, known := closedVocabularies[field]
	if !known {
		return strings.TrimSpace(value), true
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", true
	}

	for _, candidate := range allowed {
		if strings.EqualFold(candidate, trimmed) {
			return candidate, true
		}
	}

	return "", false
}

// VocabularyError names the field and lists what it accepts; "invalid" alone sends
// the reader back to the spec.
func VocabularyError(field string, value string) string {
	allowed, known := closedVocabularies[field]
	if !known {
		return field + ": unexpected value " + value
	}

	return headerFor(field) + ": \"" + strings.TrimSpace(value) + "\" is not one of " + strings.Join(allowed, ", ")
}

// AllowedValues returns a copy of the vocabulary for field, or nil when it is open.
func AllowedValues(field string) []string {
	allowed, known := closedVocabularies[field]
	if !known {
		return nil
	}

	out := make([]string, len(allowed))
	copy(out, allowed)

	return out
}

// headerFor maps a column key back to the header the operator sees, so an error
// names what is printed in the sheet rather than the internal key.
func headerFor(key string) string {
	for _, column := range BuildingColumns {
		if column.Key == key {
			return column.Header
		}
	}

	return key
}

// BuildTemplate produces the empty workbook: header row, one example row, and a
// second sheet explaining every column.
func BuildTemplate() ([]byte, error) {
	return spreadsheets.BuildTemplate(BuildingSheetName, BuildingColumns)
}

// TemplateHeaders is the header row in order, so template and export cannot drift
// into different column lists.
func TemplateHeaders() []string {
	headers := make([]string, len(BuildingColumns))
	for i, column := range BuildingColumns {
		headers[i] = column.Header
	}

	return headers
}
