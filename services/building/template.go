package building

import (
	"strings"

	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
)

// BuildingSheetName is the tab the importer reads.
const BuildingSheetName = "Buildings"

// Vocabularies for the columns that have one.
//
// building_status is CLOSED because LCD presence is derived from it: "BAST Signed"
// must keep that exact spelling or a building silently stops counting as TMN's. The
// other two are closed because the building form already restricts them.
//
// Grade stays open -- a new one must not need a deploy. Building type is not open
// either, but it is not listed here: unknown values collapse to "Other" rather than
// rejecting the row, because the map's filter chips are built from a fixed list.
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
	"building_status":      BuildingStatuses,
	"sellable":             SellableValues,
	"connectivity":         ConnectivityValues,
	"competitor_presence":  YesNoValues,
	"competitor_exclusive": YesNoValues,
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
		Note: "One of: " + strings.Join(BuildingStatuses, ", ") +
			". IMPORTANT: LCD presence is derived from this plus the two competitor columns, so 'BAST Signed' must keep that exact spelling or the building stops counting as TMN's.",
	},
	{
		Key: "competitor_presence", Header: "Competitor Presence", Required: false,
		Example: "no", Note: "yes or no. Feeds the derived LCD presence.",
	},
	{
		Key: "competitor_exclusive", Header: "Competitor Exclusive", Required: false,
		Example: "no", Note: "yes or no. Feeds the derived LCD presence.",
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
