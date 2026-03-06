//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Buildings come from ERP sync — there is no POST /buildings endpoint.
// All building tests seed data directly via insertBuilding().

func TestBuildingFindAll_Empty(t *testing.T) {
	truncateTables(t, "buildings")

	resp := doWithAuth(t, http.MethodGet, "/buildings", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var list []interface{}
	DecodeData(t, wr, &list)
	assert.Empty(t, list)
}

func TestBuildingFindAll_WithData(t *testing.T) {
	truncateTables(t, "buildings")
	insertBuilding(t, "Building One")
	insertBuilding(t, "Building Two")

	resp := doWithAuth(t, http.MethodGet, "/buildings", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var list []interface{}
	DecodeData(t, wr, &list)
	assert.Len(t, list, 2)
}

func TestBuildingFindAll_FilteredByBuildingType(t *testing.T) {
	truncateTables(t, "buildings")

	// insertBuilding seeds building_type='Office'; insert a Retail one manually
	insertBuilding(t, "Office Building")
	_, err := testSuite.db.Exec(`
		INSERT INTO buildings (name, building_type, grade_resource, latitude, longitude)
		VALUES ('Retail Building', 'Retail', 'B', -6.3, 106.9)`)
	require.NoError(t, err)

	resp := doWithAuth(t, http.MethodGet, "/buildings?building_type=Office", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var list []interface{}
	DecodeData(t, wr, &list)
	assert.Len(t, list, 1)
	b := list[0].(map[string]interface{})
	assert.Equal(t, "Office", b["building_type"])
}

func TestBuildingFindById_HappyPath(t *testing.T) {
	truncateTables(t, "buildings")
	id := insertBuilding(t, "Specific Building")

	resp := doWithAuth(t, http.MethodGet, fmt.Sprintf("/buildings/%d", id), nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)

	assert.Equal(t, "Specific Building", data["name"])
	assert.InDelta(t, -6.2, data["latitude"], 0.001)
	assert.InDelta(t, 106.8, data["longitude"], 0.001)
}

func TestBuildingFindById_NotFound(t *testing.T) {
	resp := doWithAuth(t, http.MethodGet, "/buildings/9999999", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestBuildingGetFilterOptions(t *testing.T) {
	truncateTables(t, "buildings")
	insertBuilding(t, "Filter Test Building")

	resp := doWithAuth(t, http.MethodGet, "/building-filter-options", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBuildingGetDropdown(t *testing.T) {
	truncateTables(t, "buildings")
	insertBuilding(t, "Dropdown Building")

	resp := doWithAuth(t, http.MethodGet, "/building-dropdown", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var list []interface{}
	DecodeData(t, wr, &list)
	assert.NotEmpty(t, list)

	// Each item should have id, name, and building_type
	b := list[0].(map[string]interface{})
	assert.NotNil(t, b["id"])
	assert.NotNil(t, b["name"])
	assert.NotNil(t, b["building_type"])
}

func TestBuildingFindAllForMapping(t *testing.T) {
	truncateTables(t, "buildings")
	insertBuilding(t, "Mapping Building")

	resp := doWithAuth(t, http.MethodGet, "/mapping-buildings", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBuildingUpdate_HappyPath(t *testing.T) {
	truncateTables(t, "buildings")
	id := insertBuilding(t, "Updatable Building")

	updateBody := map[string]interface{}{
		"sellable":     "not_sell",
		"connectivity": "manual",
	}
	resp := doWithAuth(t, http.MethodPut, fmt.Sprintf("/buildings/%d", id), updateBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)
	assert.Equal(t, "not_sell", data["sellable"])
	assert.Equal(t, "manual", data["connectivity"])
}

func TestGetLCDPresenceSummary(t *testing.T) {
	truncateTables(t, "buildings")
	insertBuilding(t, "LCD Building")

	resp := doWithAuth(t, http.MethodGet, "/dashboard/building-lcd-presence", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
