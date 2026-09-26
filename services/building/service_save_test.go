package building

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/malikabdulaziz/tmn-backend/models"
	webBuilding "github.com/malikabdulaziz/tmn-backend/web/building"
)

// The form and the spreadsheet must agree about every derived value, or the same
// building would read differently depending on which way it was edited.
func TestRequestToBuilding_DerivesWhatMustNotBeTyped(t *testing.T) {
	building := requestToBuilding(webBuilding.SaveBuildingRequest{
		ExternalBuildingId: "  BLDG-1 ", Name: " Tower 1 ",
		BuildingStatus:     "bast signed",
		CompetitorPresence: false, CompetitorExclusive: false,
		BuildingType: "apartment",
	})

	assert.Equal(t, "BLDG-1", building.ExternalBuildingId, "trimmed")
	assert.Equal(t, "Tower 1", building.Name)
	assert.Equal(t, "BAST Signed", building.BuildingStatus, "canonicalised, whatever the casing")
	assert.Equal(t, "Apartment", building.BuildingType)
	assert.Equal(t, "TMN", building.LcdPresenceStatus, "derived, never typed")
	assert.False(t, building.CompetitorLocation, "kept in step with presence")
}

func TestRequestToBuilding_CompetitorLocationMirrorsPresence(t *testing.T) {
	building := requestToBuilding(webBuilding.SaveBuildingRequest{
		ExternalBuildingId: "B", Name: "N", CompetitorPresence: true,
	})

	assert.True(t, building.CompetitorLocation)
	assert.Equal(t, "Competitor", building.LcdPresenceStatus)
}

// A blank type stays blank rather than becoming "Other" -- same rule as the import,
// where a blank cell means "clear this".
func TestRequestToBuilding_BlankTypeStaysBlank(t *testing.T) {
	building := requestToBuilding(webBuilding.SaveBuildingRequest{
		ExternalBuildingId: "B", Name: "N", BuildingType: "   ",
	})

	assert.Equal(t, "", building.BuildingType)
}

// An unfamiliar status is stored as written here too: ERP adds statuses without
// warning, and the form must not be stricter than the file.
func TestRequestToBuilding_UnknownStatusIsKept(t *testing.T) {
	building := requestToBuilding(webBuilding.SaveBuildingRequest{
		ExternalBuildingId: "B", Name: "N", BuildingStatus: "Building Onboarded",
	})

	assert.Equal(t, "Building Onboarded", building.BuildingStatus)
	assert.Equal(t, "Opportunity", building.LcdPresenceStatus, "not BAST Signed, so not TMN's")
}

// Every column the form owns must be one the change log tracks, or an edit made here
// would go unrecorded while the same edit by spreadsheet is logged.
func TestSaveRequest_CoversOnlyTrackedFields(t *testing.T) {
	building := requestToBuilding(webBuilding.SaveBuildingRequest{
		ExternalBuildingId: "B", Name: "N", IrisCode: "I", Subdistrict: "S",
		Citytown: "C", Province: "P", CbdArea: "A", GradeResource: "G",
		CompletionYear: 2019, Audience: 1, Impression: 2,
		Sellable: "sell", Connectivity: "online", ResourceType: "R",
		Latitude: -6.2, Longitude: 106.8,
	})

	values := fieldValues(building)
	for _, field := range []string{
		"external_building_id", "iris_code", "name", "subdistrict", "citytown",
		"province", "cbd_area", "grade_resource", "completion_year",
		"audience", "impression", "sellable", "connectivity", "resource_type",
		"latitude", "longitude",
	} {
		assert.NotEmpty(t, values[field], "%q should be carried from the form request", field)
	}
}

// A stub project raised while saving the building FORM must be recorded as a form
// change, not an import. The database refuses the alternative -- a form row carries
// no batch id, and building_project_changes pairs source and batch_id in a check
// constraint -- so getting this wrong is a 500 on save, which is how it was found.
func TestResolveProjectSource_FormAndImportAreDistinct(t *testing.T) {
	assert.Equal(t, "form", models.BuildingProjectSourceForm)
	assert.Equal(t, "import", models.BuildingProjectSourceImport)
	assert.NotEqual(t, models.BuildingProjectSourceForm, models.BuildingProjectSourceImport,
		"the two sources must stay distinct: the check constraint keys off them")
}
