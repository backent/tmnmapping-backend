package building

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/malikabdulaziz/tmn-backend/services/erp"
)

// ERP supplies two IRIS codes with a trailing tab. "B003027\t" is a different string
// from "B003027" everywhere it is compared: the price import looks buildings up by
// this value, and the spreadsheet round trip reported both rows as changed on every
// upload because the sheet trims and the database did not.
//
// Trimming in the database does not hold -- the next sync writes the untrimmed value
// straight back, which is exactly what happened on 2026-09-23. It has to happen on
// the way in.
func TestTrimERPBuilding_StripsWhitespaceFromEveryTextField(t *testing.T) {
	building := erp.ERPBuilding{
		BuildingId:      " BLDG-1 ",
		IrisCode:        "B003027\t",
		BuildingName:    "\tGading Tower 1 ",
		BuildingProject: " Gading Resort Residence\n",
		CbdArea:         " Non-CBD ",
		Subdistrict:     " Kelapa Gading ",
		Citytown:        "Jakarta Utara\t",
		Province:        " DKI Jakarta",
		GradeResource:   "Grade A ",
		BuildingType:    " Apartment ",
	}

	trimERPBuilding(&building)

	assert.Equal(t, "BLDG-1", building.BuildingId)
	assert.Equal(t, "B003027", building.IrisCode)
	assert.Equal(t, "Gading Tower 1", building.BuildingName)
	assert.Equal(t, "Gading Resort Residence", building.BuildingProject)
	assert.Equal(t, "Non-CBD", building.CbdArea)
	assert.Equal(t, "Kelapa Gading", building.Subdistrict)
	assert.Equal(t, "Jakarta Utara", building.Citytown)
	assert.Equal(t, "DKI Jakarta", building.Province)
	assert.Equal(t, "Grade A", building.GradeResource)
	assert.Equal(t, "Apartment", building.BuildingType)
}

// A clean record must come through untouched, so trimming cannot be blamed for a
// change the sync reports.
func TestTrimERPBuilding_LeavesCleanValuesAlone(t *testing.T) {
	building := erp.ERPBuilding{BuildingId: "BLDG-1", IrisCode: "B003027", BuildingName: "Tower 1"}
	original := building

	trimERPBuilding(&building)

	assert.Equal(t, original, building)
}
