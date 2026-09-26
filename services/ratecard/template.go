package ratecard

import "github.com/malikabdulaziz/tmn-backend/spreadsheets"

// BuildingPriceColumns defines the building price upload.
//
// The key is the IRIS Building ID, matching buildings.iris_code. That is what the
// rate card source spreadsheet uses (B000006). external_building_id is a different
// identifier entirely (BLDG-2025-09-02547) and matched none of the 1,620 rows in the
// sample file, so keying on it would reject every row.
//
// Building Name is carried for the human reading the sheet and ignored on upload:
// tower names repeat across a project, so they cannot identify a building.
var BuildingPriceColumns = []spreadsheets.SheetColumn{
	{
		Key: "iris_building_id", Header: "IRIS Building ID", Required: true,
		Example: "B000006",
		Note:    "Must match a building already in the system. Export this sheet first to get the exact values.",
	},
	{
		Key: "building_name", Header: "Building Name", Required: false,
		Example: "Kubikahomy Apartment - Tower A",
		Note:    "For your reference only. Ignored on upload -- the IRIS Building ID decides which building is priced.",
	},
	{
		Key: "price", Header: "Price per Week (IDR)", Required: true,
		Example: "1100000",
		Note:    "What the ADVERTISER PAYS for one week, in whole rupiah. Not the landlord rent. A campaign of N weeks is charged price x N. Rows priced 0 are skipped, not sold for free.",
	},
}

// PackagePriceColumns defines the sales package price upload.
//
// There is no column for which buildings a package contains: that is taken from the
// sales package master at publish time and frozen then, so it cannot drift.
var PackagePriceColumns = []spreadsheets.SheetColumn{
	{
		Key: "package_name", Header: "Sales Package", Required: true,
		Example: "Jakarta CBD Premium",
		Note:    "Must match an existing sales package name exactly.",
	},
	{
		Key: "price", Header: "Price per Week (IDR)", Required: true,
		Example: "112500000",
		Note:    "What the ADVERTISER PAYS for one week of the whole package, in whole rupiah.",
	},
}
