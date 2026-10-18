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
	{
		Key: "attention_to", Header: "Attention To", Required: true, Example: "Budi Santoso",
		Note: "The person a quotation for this brand is addressed to. Printed on the quotation document.",
	},
	{
		Key: "job_title", Header: "Job Title", Required: true, Example: "Marketing Director",
		Note: "Their job title, printed beneath their name on the quotation document.",
	},
	{
		Key: "contact_phone", Header: "Contact Phone", Required: true, Example: "+62 812 3456 7890",
		Note: "Contact number printed on the quotation document.",
	},
	{
		Key: "contact_email", Header: "Contact Email", Required: true, Example: "budi@example.com",
		Note: "Contact email printed on the quotation document.",
	},
}

// brandContactError reports what is missing from a brand's contact, or "" when it is
// complete.
//
// A quotation is addressed to a person, and the printed document carries all four
// fields, so a brand without them cannot be quoted usefully. Kept beside the template
// columns so the spreadsheet rule and the column list stay together.
func brandContactError(attentionTo, jobTitle, contactPhone, contactEmail string) string {
	switch {
	case attentionTo == "":
		return "Attention To is required"
	case jobTitle == "":
		return "Job Title is required"
	case contactPhone == "":
		return "Contact Phone is required"
	case contactEmail == "":
		return "Contact Email is required"
	}

	return ""
}
