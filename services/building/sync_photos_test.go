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
