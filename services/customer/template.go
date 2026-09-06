package customer

import "github.com/malikabdulaziz/tmn-backend/spreadsheets"

// TemplateColumns defines the upload template and, by the same definition, what an
// uploaded file is parsed against. Template and importer cannot drift because they
// read the same list.
var TemplateColumns = []spreadsheets.SheetColumn{
	{
		Key: "code", Header: "Customer Code", Required: true, Example: "CUST-001",
		Note: "Unique identifier for the advertiser. Re-uploading the same code updates that customer instead of creating a duplicate.",
	},
	{
		Key: "name", Header: "Customer Name", Required: true, Example: "Kopi Nusantara Group",
		Note: "Registered or trading name of the advertiser.",
	},
	{
		Key: "industry", Header: "Industry", Required: false, Example: "Food & Beverage",
		Note: "Optional. Free text, used for reporting only.",
	},
	{
		Key: "status", Header: "Status", Required: false, Example: "active",
		Note: "active or inactive. Defaults to active when left blank. Inactive customers stay in history but are hidden from the quotation wizard.",
	},
}
