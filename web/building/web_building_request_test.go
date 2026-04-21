package building

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildMappingRequestFromBody_Filters(t *testing.T) {
	body := &MappingByFilterRequest{
		Filters: ExportMappingFilters{
			DistrictSubdistrict: []string{"Jakarta Pusat", "Menteng"},
			BuildingType:        []string{"Office", "Hotel"},
			LcdPresence:         []string{"TMN"},
		},
		MapCenter: &struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		}{Lat: -6.2, Lng: 106.8166},
	}

	req := BuildMappingRequestFromBody(body)

	assert.Equal(t, "Jakarta Pusat,Menteng", req.GetSubdistrict())
	assert.Equal(t, "Office,Hotel", req.GetBuildingType())
	assert.Equal(t, "TMN", req.GetLCDPresence())
	assert.Equal(t, "-6.2", req.GetLat())
	assert.Equal(t, "106.8166", req.GetLng())
}

func TestBuildMappingRequestFromBody_PolygonThreeOrMorePoints(t *testing.T) {
	body := &MappingByFilterRequest{
		Filters: ExportMappingFilters{
			Polygon: []struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			}{
				{Lat: -6.0, Lng: 106.0},
				{Lat: -6.1, Lng: 106.1},
				{Lat: -6.2, Lng: 106.2},
			},
		},
	}

	req := BuildMappingRequestFromBody(body)

	polygon := req.GetPolygon()
	assert.NotEmpty(t, polygon)
	assert.True(t, strings.Contains(polygon, "-6.1"))
	assert.True(t, strings.Contains(polygon, "106.2"))
}

func TestBuildMappingRequestFromBody_PolygonUnderThreePointsIgnored(t *testing.T) {
	body := &MappingByFilterRequest{
		Filters: ExportMappingFilters{
			Polygon: []struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			}{
				{Lat: -6.0, Lng: 106.0},
				{Lat: -6.1, Lng: 106.1},
			},
		},
	}

	req := BuildMappingRequestFromBody(body)

	assert.Empty(t, req.GetPolygon())
}

func TestBuildMappingRequestFromBody_BoundsSet(t *testing.T) {
	body := &MappingByFilterRequest{
		Bounds: &MappingBounds{
			MinLat: -7.0,
			MaxLat: -6.0,
			MinLng: 106.0,
			MaxLng: 107.0,
		},
	}

	req := BuildMappingRequestFromBody(body)

	assert.Equal(t, "-7", req.GetMinLat())
	assert.Equal(t, "-6", req.GetMaxLat())
	assert.Equal(t, "106", req.GetMinLng())
	assert.Equal(t, "107", req.GetMaxLng())
}

func TestBuildMappingRequestFromBody_BoundsNil(t *testing.T) {
	body := &MappingByFilterRequest{}

	req := BuildMappingRequestFromBody(body)

	assert.Empty(t, req.GetMinLat())
	assert.Empty(t, req.GetMaxLat())
	assert.Empty(t, req.GetMinLng())
	assert.Empty(t, req.GetMaxLng())
}

// Regression: the frontend MapBounds interface uses camelCase keys (minLat, minLng,
// maxLat, maxLng). If the Go struct tags drift to snake_case, json.Decode succeeds
// but leaves Bounds as a non-nil pointer with zero values — which then gets applied
// as a bounding box around (0,0), silently returning an empty building list.
func TestMappingByFilterRequest_DecodesFrontendBoundsPayload(t *testing.T) {
	raw := []byte(`{
		"filters": {"lcd_presence": ["TMN"]},
		"map_center": {"lat": -6.2, "lng": 106.8166},
		"bounds": {"minLat": -6.22, "minLng": 106.77, "maxLat": -6.17, "maxLng": 106.86}
	}`)

	var body MappingByFilterRequest
	err := json.Unmarshal(raw, &body)
	assert.NoError(t, err)

	assert.NotNil(t, body.Bounds)
	assert.InDelta(t, -6.22, body.Bounds.MinLat, 1e-9)
	assert.InDelta(t, 106.77, body.Bounds.MinLng, 1e-9)
	assert.InDelta(t, -6.17, body.Bounds.MaxLat, 1e-9)
	assert.InDelta(t, 106.86, body.Bounds.MaxLng, 1e-9)

	req := BuildMappingRequestFromBody(&body)
	assert.Equal(t, "-6.22", req.GetMinLat())
	assert.Equal(t, "-6.17", req.GetMaxLat())
	assert.Equal(t, "106.77", req.GetMinLng())
	assert.Equal(t, "106.86", req.GetMaxLng())
}

// Regression: the frontend MappingFilters sends `poi_ids: number[]` (plural,
// array of category ids), not `poi_id: int`. A field-name/type mismatch would
// silently drop the filter and fall back to a map_center radius search —
// visibly wrong markers on the map for the user.
func TestMappingByFilterRequest_DecodesPoiIdsPayload(t *testing.T) {
	raw := []byte(`{
		"filters": {"lcd_presence": ["TMN"], "poi_ids": [90, 91], "radius": 2},
		"map_center": {"lat": -6.2, "lng": 106.8166}
	}`)

	var body MappingByFilterRequest
	err := json.Unmarshal(raw, &body)
	assert.NoError(t, err)

	assert.Equal(t, []int{90, 91}, body.Filters.PoiIDs)

	req := BuildMappingRequestFromBody(&body)
	assert.Equal(t, "90,91", req.GetPOIId())
	assert.Equal(t, "2000", req.GetRadius())
}

func TestBuildMappingRequestFromBody_RadiusKmToMeters(t *testing.T) {
	radius := 2.5 // km
	body := &MappingByFilterRequest{
		Filters: ExportMappingFilters{
			Radius: &radius,
		},
	}

	req := BuildMappingRequestFromBody(body)

	// 2.5 km -> 2500 m
	assert.Equal(t, "2500", req.GetRadius())
}
