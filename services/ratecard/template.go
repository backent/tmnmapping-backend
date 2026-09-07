package ratecard

import "github.com/malikabdulaziz/tmn-backend/spreadsheets"

// BuildingPriceColumns defines the building price upload.
//
// The key is External Building ID because that is what the ERP sync writes and what
// stays stable across syncs. Building Name is carried for the human reading the
// sheet and is ignored on the way back in.
var BuildingPriceColumns = []spreadsheets.SheetColumn{
	{
		Key: "external_building_id", Header: "External Building ID", Required: true,
		Example: "BLDG-2025-09-02547",
		Note:    "Must match a building already in the system. Export this sheet first to get the exact values.",
	},
	{
		Key: "building_name", Header: "Building Name", Required: false,
		Example: "Menara BCA",
		Note:    "For your reference only. Ignored on upload — the External Building ID decides which building is priced.",
	},
	{
		Key: "price", Header: "Price per 4 Weeks (IDR)", Required: true,
		Example: "92000000",
		Note:    "What the ADVERTISER PAYS for four weeks, in whole rupiah. Not the landlord rent. A campaign of N weeks is charged price x N / 4.",
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
		Key: "price", Header: "Price per 4 Weeks (IDR)", Required: true,
		Example: "450000000",
		Note:    "What the ADVERTISER PAYS for four weeks of the whole package, in whole rupiah.",
	},
}
