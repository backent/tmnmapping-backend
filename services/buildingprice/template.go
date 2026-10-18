package buildingprice

import "github.com/malikabdulaziz/tmn-backend/spreadsheets"

// BuildingPriceColumns defines the price upload. It is the shape the rate card used,
// so a file exported from either imports here.
//
// The key is the IRIS Building ID, matching buildings.iris_code (B000006). Building
// Name is carried for the person reading the sheet and ignored on upload: tower names
// repeat across a project, so they cannot identify a building.
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
		Note:    "What the ADVERTISER PAYS for one week, in whole rupiah. A campaign of N weeks is charged price x N. Rows priced 0 are skipped, not sold for free.",
	},
}

// priceHeaderAliases are other names the price column arrives under. The business's
// own rate card workbook calls it "Round Up", and that workbook is the file people
// actually have; making them rename a column before every upload is friction with
// no benefit.
var priceHeaderAliases = []string{"Round Up"}
