package buildingproject

import (
	"strings"

	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
)

// BuildingProjectSheetName is the tab the importer reads. A workbook may carry other
// sheets; only this one is parsed.
const BuildingProjectSheetName = "Projects"

// BuildingProjectColumns defines the upload, and is the single source of truth for
// the template, the export and the importer. TestTemplateMatchesTrackedFields keeps it
// in step with the model, so a column cannot exist in one and not the other.
//
// Headers match the business's own workbook wherever it already had one, so adapting
// an existing file is a matter of deleting columns rather than renaming them.
//
// Format: row 1 is the header, row 2 onward is data. No sub-header row, no totals row,
// nothing beyond these columns. A file in the old shape -- sub-headers in row 2, a
// SUM row in row 3, 224 projection columns -- is rejected on the header check rather
// than half-imported.
var BuildingProjectColumns = []spreadsheets.SheetColumn{
	{
		Key: "project_id_iris", Header: "Project ID IRIS", Required: true,
		Example: "PRJ-0001",
		Note:    "The key. Required, and must be unique. Decides whether a row creates a project or updates one, so it must match the existing value exactly when editing. Export first rather than typing it.",
	},
	{
		Key: "name", Header: "Project Name", Required: true,
		Example: "Gading Resort Residence",
		Note:    "Required. The site, not one tower -- towers are buildings and belong to this project. Keep the spelling identical to ERP: the Letter of Intent dashboard matches buildings to projects by this text.",
	},
	{
		Key: "building_type", Header: "Building Type", Required: false,
		Example: "Apartment",
		Note:    "Free text. Office, Apartment, Mall, Hotel, Hospital, University and so on. A new category is accepted, not rejected.",
	},
	{
		Key: "grade", Header: "Grade", Required: false,
		Example: "Grade A",
		Note:    "Free text. Usually Premium, Grade A, Grade B or Standard. A new grade is accepted.",
	},
	{
		Key: "pic", Header: "PIC", Required: false,
		Example: "Dara Maheswari",
		Note:    "The person who owns this acquisition.",
	},
	{
		Key: "tmn_project_status", Header: "TMN Project Status", Required: false,
		Example: "Active",
		Note:    "One of: " + strings.Join(TmnProjectStatuses, ", ") + ". Case does not matter; anything else is rejected with the row.",
	},
	{
		Key: "no_of_tower", Header: "No of Tower", Required: false,
		Example: "4",
		Note:    "Whole number. Declared by the business and NOT checked against how many buildings actually carry this project -- the two are maintained independently.",
	},
	{
		Key: "no_of_screen", Header: "No of Screen", Required: false,
		Example: "21",
		Note:    "Whole number. Independent of the screen counts on individual buildings; neither is derived from the other.",
	},
	{
		Key: "created_date", Header: "Create Date", Required: false,
		Example: "2026-01-05",
		Note:    "YYYY-MM-DD. When the project was opened, not when this row was uploaded.",
	},
	{
		Key: "remark", Header: "Remark", Required: false,
		Example: "",
		Note:    "Free text.",
	},
	{
		Key: "contract_type", Header: "Contract Type", Required: false,
		Example: "Initial",
		Note:    "One of: " + strings.Join(ContractTypes, ", ") + ". Describes the CURRENT contract. Previous contracts are not kept as rows; the change history records what each value used to be.",
	},
	{
		Key: "contract_no", Header: "Contract No", Required: false,
		Example: "CON/2026/001",
		Note:    "Restricted. Only finance roles can see this column in the app or in an export.",
	},
	{
		Key: "contract_date", Header: "Contract Date", Required: false,
		Example: "2026-01-18",
		Note:    "YYYY-MM-DD. When the contract was signed, which is not necessarily when it starts.",
	},
	{
		Key: "contract_start", Header: "Contract Start", Required: false,
		Example: "2026-02-01",
		Note:    "YYYY-MM-DD.",
	},
	{
		Key: "contract_end", Header: "Contract End", Required: false,
		Example: "2027-01-31",
		Note:    "YYYY-MM-DD. Must not be earlier than Contract Start.",
	},
	{
		Key: "period_month", Header: "Period Month", Required: false,
		Example: "12",
		Note:    "Whole number of months, greater than zero. Usually 12, 24, 36 or 60.",
	},
	{
		Key: "annual_rental", Header: "Annual Rental", Required: false,
		Example: "24000000",
		Note:    "Whole rupiah per YEAR, no separators or currency symbol. This is what TMN PAYS the landlord -- it is cost, not the price an advertiser pays. Restricted to finance roles. Price Per Screen and Contract Value are calculated from it and are not columns here.",
	},
	{
		Key: "payment_term", Header: "Payment Term", Required: false,
		Example: "Monthly",
		Note:    "One of: " + strings.Join(PaymentTerms, ", ") + ".",
	},
	{
		Key: "company_name", Header: "Company Name", Required: false,
		Example: "PT Arunika Properti",
		Note:    "The landlord or building owner. Restricted to finance roles.",
	},
	{
		Key: "exclusivity", Header: "Exclusivity", Required: false,
		Example: "Non-Exclusive",
		Note:    "One of: " + strings.Join(Exclusivities, ", ") + ".",
	},
	{
		Key: "doc_type", Header: "Doc. Type", Required: false,
		Example: "PKS",
		Note:    "One of: " + strings.Join(DocTypes, ", ") + ".",
	},
	{
		Key: "contract_status", Header: "Contract Status", Required: false,
		Example: "Signed",
		Note:    "One of: " + strings.Join(ContractStatuses, ", ") + ".",
	},
	{
		Key: "cancelled_at", Header: "Cancelled/Terminate Date", Required: false,
		Example: "",
		Note:    "YYYY-MM-DD. Leave empty unless the contract ended early. Setting it does not change any earlier month: history is not rewritten by a cancellation.",
	},
	{
		Key: "cancel_last_status", Header: "Last Status", Required: false,
		Example: "",
		Note:    "The status the project held before it was cancelled or terminated. Only meaningful alongside Cancelled/Terminate Date.",
	},
	{
		Key: "cancel_reason", Header: "Reason", Required: false,
		Example: "",
		Note:    "Why it was cancelled or terminated.",
	},
}

// BuildTemplate produces the empty workbook: header row, one example row, and a
// second sheet explaining every column. The service exposes it as Template().
func BuildTemplate() ([]byte, error) {
	return spreadsheets.BuildTemplate(BuildingProjectSheetName, BuildingProjectColumns)
}

// TemplateHeaders is the header row in order, used by the export so template and
// export cannot drift into different column lists.
func TemplateHeaders() []string {
	headers := make([]string, len(BuildingProjectColumns))
	for i, column := range BuildingProjectColumns {
		headers[i] = column.Header
	}

	return headers
}
