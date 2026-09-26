package building

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/services/erp"
)

func TestErpImages_DropsTheEmptyPaths(t *testing.T) {
	images := erpImages(erp.ERPBuilding{
		FrontSidePhoto: "/files/front.jpg",
		BackSidePhoto:  "",
		LeftSidePhoto:  "/files/left.jpg",
		RightSidePhoto: "",
	})

	assert.Equal(t, []models.BuildingImage{
		{Name: "front", Path: "/files/front.jpg"},
		{Name: "left", Path: "/files/left.jpg"},
	}, images)
}

func TestErpImages_NoPhotosIsAnEmptySliceNotNil(t *testing.T) {
	images := erpImages(erp.ERPBuilding{})

	assert.NotNil(t, images)
	assert.Empty(t, images)
}

// The sync runs on a timer against a table people now edit by hand. Writing a row
// whose photos have not moved would bump updated_at on all 3,747 buildings every run
// and make "when did this last change" meaningless.
func TestSameImages_OnlyWritesWhenThePhotosMoved(t *testing.T) {
	a := []models.BuildingImage{{Name: "front", Path: "/a.jpg"}, {Name: "back", Path: "/b.jpg"}}

	assert.True(t, sameImages(a, []models.BuildingImage{
		{Name: "front", Path: "/a.jpg"}, {Name: "back", Path: "/b.jpg"},
	}))

	assert.False(t, sameImages(a, []models.BuildingImage{
		{Name: "front", Path: "/CHANGED.jpg"}, {Name: "back", Path: "/b.jpg"},
	}), "a moved path is a change")

	assert.False(t, sameImages(a, []models.BuildingImage{{Name: "front", Path: "/a.jpg"}}),
		"a removed photo is a change")

	assert.True(t, sameImages(nil, []models.BuildingImage{}),
		"no photos either way is not a change")
}

// ERP holds more than one Building row for the same building_id -- three of them,
// each with a different set of photos. Processing both wrote A, then B, then A again
// next run: six rows rewritten on every sync forever, with updated_at bouncing on a
// table people now edit by hand. Verified against the live ERP on 2026-09-23.
func TestDedupeERPBuildings_KeepsTheFirstRecordPerId(t *testing.T) {
	unique, duplicates := dedupeERPBuildings([]erp.ERPBuilding{
		{BuildingId: "BLDG-1", FrontSidePhoto: "/first.jpg"},
		{BuildingId: "BLDG-2", FrontSidePhoto: "/other.jpg"},
		{BuildingId: "BLDG-1", FrontSidePhoto: "/second.jpg"},
	})

	assert.Equal(t, 1, duplicates)
	assert.Len(t, unique, 2)
	assert.Equal(t, "/first.jpg", unique[0].FrontSidePhoto,
		"first wins, and must win the same way every run")
	assert.Equal(t, "BLDG-2", unique[1].BuildingId)
}

// Whitespace must not make one record look like two.
func TestDedupeERPBuildings_MatchesOnTrimmedId(t *testing.T) {
	unique, duplicates := dedupeERPBuildings([]erp.ERPBuilding{
		{BuildingId: "BLDG-1"},
		{BuildingId: " BLDG-1 "},
	})

	assert.Equal(t, 1, duplicates)
	assert.Len(t, unique, 1)
}

// A record with no id cannot be deduped and is left for the sync to skip, rather than
// collapsing every such record into one.
func TestDedupeERPBuildings_KeepsUnidentifiedRecords(t *testing.T) {
	unique, duplicates := dedupeERPBuildings([]erp.ERPBuilding{
		{BuildingId: ""}, {BuildingId: ""},
	})

	assert.Equal(t, 0, duplicates)
	assert.Len(t, unique, 2)
}
