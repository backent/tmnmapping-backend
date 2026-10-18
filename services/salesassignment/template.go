package salesassignment

import "github.com/malikabdulaziz/tmn-backend/spreadsheets"

var TemplateColumns = []spreadsheets.SheetColumn{
	{
		Key: "customer_code", Header: "Customer Code", Required: true, Example: "CUST-001",
		Note: "Must match an existing Customer Code.",
	},
	{
		Key: "brand_code", Header: "Brand Code", Required: true, Example: "BRND-001",
		Note: "Must match an existing Brand Code, and that brand must belong to the customer above.",
	},
	{
		Key: "sales_username", Header: "Sales Username", Required: true, Example: "sales1",
		Note: "Username of the sales PIC. The account must already exist under Administration > Users.",
	},
	{
		Key: "status", Header: "Status", Required: false, Example: "active",
		Note: "active or inactive. Defaults to active when left blank.",
	},
	{
		Key: "registration_date", Header: "Registration Date", Required: false, Example: "2026-01-15",
		Note: "Optional. Format YYYY-MM-DD.",
	},
	{
		Key: "expiry_date", Header: "Expiry Date", Required: false, Example: "2026-12-31",
		Note: "Optional. Format YYYY-MM-DD. Must not be earlier than the registration date.",
	},
}
