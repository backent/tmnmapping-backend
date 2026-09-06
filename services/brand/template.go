package brand

import "github.com/malikabdulaziz/tmn-backend/spreadsheets"

var TemplateColumns = []spreadsheets.SheetColumn{
	{
		Key: "code", Header: "Brand Code", Required: true, Example: "BRND-001",
		Note: "Unique identifier for the brand. Re-uploading the same code updates that brand instead of creating a duplicate.",
	},
	{
		Key: "customer_code", Header: "Customer Code", Required: true, Example: "CUST-001",
		Note: "Must match a Customer Code that already exists. Upload the customer file first.",
	},
	{
		Key: "name", Header: "Brand Name", Required: true, Example: "Kopi Kenangan",
		Note: "The advertiser brand a quotation is raised for.",
	},
	{
		Key: "category", Header: "Category", Required: false, Example: "Ready-to-drink coffee",
		Note: "Optional. Free text, used for reporting only.",
	},
	{
		Key: "status", Header: "Status", Required: false, Example: "active",
		Note: "active or inactive. Defaults to active when left blank.",
	},
}
