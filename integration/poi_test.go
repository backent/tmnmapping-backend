//go:build integration

package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// poiCreateBody returns a valid CreatePOIRequest payload.
func poiCreateBody(brand, color string) map[string]interface{} {
	return map[string]interface{}{
		"brand": brand,
		"color": color,
		"points": []map[string]interface{}{
			{
				"poi_name":  "Test Location",
				"address":   "Jl. Test No. 1, Jakarta",
				"latitude":  -6.2,
				"longitude": 106.8,
				"category":  "Food",
			},
		},
	}
}

func TestPOICreate_HappyPath(t *testing.T) {
	truncateTables(t, "pois")

	resp := doWithAuth(t, http.MethodPost, "/pois", poiCreateBody("Starbucks", "#00704A"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)

	assert.Equal(t, "Starbucks", data["brand"])
	assert.Equal(t, "#00704A", data["color"])

	points, ok := data["points"].([]interface{})
	require.True(t, ok, "data.points should be a list")
	require.Len(t, points, 1)

	pt := points[0].(map[string]interface{})
	assert.Equal(t, "Test Location", pt["poi_name"])
	assert.InDelta(t, -6.2, pt["latitude"], 0.001)
	assert.InDelta(t, 106.8, pt["longitude"], 0.001)
}

// TestPOICreate_WithGeographyPoint verifies that the PostGIS ST_SetSRID(ST_MakePoint(...))
// path actually works against a real database — this is the primary value of integration tests
// over unit tests with sqlmock.
func TestPOICreate_WithGeographyPoint(t *testing.T) {
	truncateTables(t, "pois")

	body := map[string]interface{}{
		"brand": "GeoTest",
		"color": "#FF0000",
		"points": []map[string]interface{}{
			{
				"poi_name":  "PostGIS Point",
				"address":   "Lat/Lng test",
				"latitude":  1.2897,  // Singapore-ish
				"longitude": 103.8501,
			},
		},
	}

	resp := doWithAuth(t, http.MethodPost, "/pois", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Round-trip: verify the coordinates survived the PostGIS round-trip
	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)

	points := data["points"].([]interface{})
	pt := points[0].(map[string]interface{})
	assert.InDelta(t, 1.2897, pt["latitude"], 0.0001)
	assert.InDelta(t, 103.8501, pt["longitude"], 0.0001)
}

func TestPOICreate_ValidationError_MissingBrand(t *testing.T) {
	body := map[string]interface{}{
		"color": "#FF0000",
		"points": []map[string]interface{}{
			{"poi_name": "x", "address": "y", "latitude": -6.2, "longitude": 106.8},
		},
	}
	resp := doWithAuth(t, http.MethodPost, "/pois", body)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPOICreate_Unauthenticated(t *testing.T) {
	req := NewRequest(t, http.MethodPost, "/pois", poiCreateBody("X", "#000"))
	resp := Do(t, req)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPOIFindAll_Empty(t *testing.T) {
	truncateTables(t, "pois")

	resp := doWithAuth(t, http.MethodGet, "/pois", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var list []interface{}
	DecodeData(t, wr, &list)
	assert.Empty(t, list)
}

func TestPOIFindAll_WithPagination(t *testing.T) {
	truncateTables(t, "pois")

	// Insert 3 POIs
	for i := 1; i <= 3; i++ {
		resp := doWithAuth(t, http.MethodPost, "/pois", poiCreateBody(
			fmt.Sprintf("Brand%d", i), "#FFFFFF",
		))
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	resp := doWithAuth(t, http.MethodGet, "/pois?take=2&skip=0", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var list []interface{}
	DecodeData(t, wr, &list)
	assert.Len(t, list, 2)

	// Extras should contain pagination metadata
	if len(wr.Extras) > 0 {
		var extras map[string]interface{}
		err := json.Unmarshal(wr.Extras, &extras)
		require.NoError(t, err)
		assert.NotNil(t, extras["total"])
	}
}

func TestPOIFindById_HappyPath(t *testing.T) {
	truncateTables(t, "pois")

	createResp := doWithAuth(t, http.MethodPost, "/pois", poiCreateBody("FindMe", "#123456"))
	require.Equal(t, http.StatusOK, createResp.StatusCode)

	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	resp := doWithAuth(t, http.MethodGet, fmt.Sprintf("/pois/%d", id), nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)
	assert.Equal(t, "FindMe", data["brand"])
}

func TestPOIFindById_NotFound(t *testing.T) {
	resp := doWithAuth(t, http.MethodGet, "/pois/9999999", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPOIUpdate_HappyPath(t *testing.T) {
	truncateTables(t, "pois")

	createResp := doWithAuth(t, http.MethodPost, "/pois", poiCreateBody("Original", "#AAAAAA"))
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	updateBody := map[string]interface{}{
		"brand": "Updated",
		"color": "#BBBBBB",
		"points": []map[string]interface{}{
			{"poi_name": "Updated Loc", "address": "New Address", "latitude": -6.9, "longitude": 107.6},
		},
	}
	resp := doWithAuth(t, http.MethodPut, fmt.Sprintf("/pois/%d", id), updateBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)
	assert.Equal(t, "Updated", data["brand"])
	assert.Equal(t, "#BBBBBB", data["color"])
}

func TestPOIDelete_HappyPath(t *testing.T) {
	truncateTables(t, "pois")

	createResp := doWithAuth(t, http.MethodPost, "/pois", poiCreateBody("ToDelete", "#CCCCCC"))
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	// Delete
	delResp := doWithAuth(t, http.MethodDelete, fmt.Sprintf("/pois/%d", id), nil)
	defer delResp.Body.Close()
	assert.Equal(t, http.StatusOK, delResp.StatusCode)

	// Subsequent GET must return 404
	getResp := doWithAuth(t, http.MethodGet, fmt.Sprintf("/pois/%d", id), nil)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}
