//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// polygonPoints returns a minimal valid set of polygon points (min 3 required by validation).
func polygonPoints() []map[string]interface{} {
	return []map[string]interface{}{
		{"lat": -6.1, "lng": 106.7},
		{"lat": -6.2, "lng": 106.8},
		{"lat": -6.3, "lng": 106.7},
	}
}

func TestSavedPolygonCreate_HappyPath(t *testing.T) {
	truncateTables(t, "saved_polygons")

	body := map[string]interface{}{
		"name":   "Test Polygon",
		"points": polygonPoints(),
	}
	resp := doWithAuth(t, http.MethodPost, "/saved-polygons", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)

	assert.Equal(t, "Test Polygon", data["name"])
	points, ok := data["points"].([]interface{})
	require.True(t, ok)
	assert.Len(t, points, 3)

	// Verify ordering (ord field) is present
	pt0 := points[0].(map[string]interface{})
	assert.NotNil(t, pt0["ord"])
}

func TestSavedPolygonCreate_TooFewPoints(t *testing.T) {
	body := map[string]interface{}{
		"name": "Bad Polygon",
		"points": []map[string]interface{}{
			{"lat": -6.1, "lng": 106.7},
			{"lat": -6.2, "lng": 106.8},
			// only 2 points — min is 3
		},
	}
	resp := doWithAuth(t, http.MethodPost, "/saved-polygons", body)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSavedPolygonCreate_Unauthenticated(t *testing.T) {
	req := NewRequest(t, http.MethodPost, "/saved-polygons", map[string]interface{}{
		"name": "X", "points": polygonPoints(),
	})
	resp := Do(t, req)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSavedPolygonFindAll(t *testing.T) {
	truncateTables(t, "saved_polygons")

	for i := 1; i <= 2; i++ {
		resp := doWithAuth(t, http.MethodPost, "/saved-polygons", map[string]interface{}{
			"name": fmt.Sprintf("Polygon %d", i), "points": polygonPoints(),
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	resp := doWithAuth(t, http.MethodGet, "/saved-polygons", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var list []interface{}
	DecodeData(t, wr, &list)
	assert.Len(t, list, 2)
}

func TestSavedPolygonFindById_HappyPath(t *testing.T) {
	truncateTables(t, "saved_polygons")

	createResp := doWithAuth(t, http.MethodPost, "/saved-polygons", map[string]interface{}{
		"name": "FindablePolygon", "points": polygonPoints(),
	})
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	resp := doWithAuth(t, http.MethodGet, fmt.Sprintf("/saved-polygons/%d", id), nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)
	assert.Equal(t, "FindablePolygon", data["name"])
}

func TestSavedPolygonFindById_NotFound(t *testing.T) {
	resp := doWithAuth(t, http.MethodGet, "/saved-polygons/9999999", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestSavedPolygonUpdate_ReplacesPoints tests the delete-then-reinsert pattern
// which exercises the UNIQUE(saved_polygon_id, ord) constraint in saved_polygon_points.
func TestSavedPolygonUpdate_ReplacesPoints(t *testing.T) {
	truncateTables(t, "saved_polygons")

	createResp := doWithAuth(t, http.MethodPost, "/saved-polygons", map[string]interface{}{
		"name": "OriginalPolygon", "points": polygonPoints(),
	})
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	newPoints := []map[string]interface{}{
		{"lat": -7.0, "lng": 107.0},
		{"lat": -7.1, "lng": 107.1},
		{"lat": -7.2, "lng": 107.0},
		{"lat": -7.15, "lng": 106.9}, // 4 points in the update
	}
	updateResp := doWithAuth(t, http.MethodPut, fmt.Sprintf("/saved-polygons/%d", id), map[string]interface{}{
		"name": "UpdatedPolygon", "points": newPoints,
	})
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	wr := DecodeWebResponse(t, updateResp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)
	assert.Equal(t, "UpdatedPolygon", data["name"])

	pts, ok := data["points"].([]interface{})
	require.True(t, ok)
	assert.Len(t, pts, 4, "update should replace with 4 new points")
	pt0 := pts[0].(map[string]interface{})
	assert.InDelta(t, -7.0, pt0["lat"], 0.001)
}

func TestSavedPolygonDelete_HappyPath(t *testing.T) {
	truncateTables(t, "saved_polygons")

	createResp := doWithAuth(t, http.MethodPost, "/saved-polygons", map[string]interface{}{
		"name": "ToDeletePolygon", "points": polygonPoints(),
	})
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	delResp := doWithAuth(t, http.MethodDelete, fmt.Sprintf("/saved-polygons/%d", id), nil)
	defer delResp.Body.Close()
	assert.Equal(t, http.StatusOK, delResp.StatusCode)

	getResp := doWithAuth(t, http.MethodGet, fmt.Sprintf("/saved-polygons/%d", id), nil)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}
