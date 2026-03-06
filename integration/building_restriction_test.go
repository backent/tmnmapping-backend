//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildingRestrictionCreate_HappyPath(t *testing.T) {
	truncateTables(t, "building_restrictions")
	bID := insertBuilding(t, "BR Building A")

	body := map[string]interface{}{
		"name":         "Restriction Alpha",
		"building_ids": []int{bID},
	}
	resp := doWithAuth(t, http.MethodPost, "/building-restrictions", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)

	assert.Equal(t, "Restriction Alpha", data["name"])
	buildings, ok := data["buildings"].([]interface{})
	require.True(t, ok)
	assert.Len(t, buildings, 1)
}

func TestBuildingRestrictionCreate_Unauthenticated(t *testing.T) {
	req := NewRequest(t, http.MethodPost, "/building-restrictions", map[string]interface{}{
		"name":         "X",
		"building_ids": []int{1},
	})
	resp := Do(t, req)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestBuildingRestrictionFindAll(t *testing.T) {
	truncateTables(t, "building_restrictions")
	bID := insertBuilding(t, "BR Building B")

	for i := 1; i <= 2; i++ {
		resp := doWithAuth(t, http.MethodPost, "/building-restrictions", map[string]interface{}{
			"name": fmt.Sprintf("Restriction %d", i), "building_ids": []int{bID},
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	resp := doWithAuth(t, http.MethodGet, "/building-restrictions", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var list []interface{}
	DecodeData(t, wr, &list)
	assert.Len(t, list, 2)
}

func TestBuildingRestrictionFindById_HappyPath(t *testing.T) {
	truncateTables(t, "building_restrictions")
	bID := insertBuilding(t, "BR Building C")

	createResp := doWithAuth(t, http.MethodPost, "/building-restrictions", map[string]interface{}{
		"name": "FindableRestriction", "building_ids": []int{bID},
	})
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	resp := doWithAuth(t, http.MethodGet, fmt.Sprintf("/building-restrictions/%d", id), nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)
	assert.Equal(t, "FindableRestriction", data["name"])
}

func TestBuildingRestrictionFindById_NotFound(t *testing.T) {
	resp := doWithAuth(t, http.MethodGet, "/building-restrictions/9999999", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestBuildingRestrictionUpdate_HappyPath(t *testing.T) {
	truncateTables(t, "building_restrictions")
	bID := insertBuilding(t, "BR Building D")

	createResp := doWithAuth(t, http.MethodPost, "/building-restrictions", map[string]interface{}{
		"name": "OriginalRestriction", "building_ids": []int{bID},
	})
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	updateResp := doWithAuth(t, http.MethodPut, fmt.Sprintf("/building-restrictions/%d", id), map[string]interface{}{
		"name": "RenamedRestriction", "building_ids": []int{bID},
	})
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	wr := DecodeWebResponse(t, updateResp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)
	assert.Equal(t, "RenamedRestriction", data["name"])
}

func TestBuildingRestrictionDelete_HappyPath(t *testing.T) {
	truncateTables(t, "building_restrictions")
	bID := insertBuilding(t, "BR Building E")

	createResp := doWithAuth(t, http.MethodPost, "/building-restrictions", map[string]interface{}{
		"name": "ToDeleteRestriction", "building_ids": []int{bID},
	})
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	delResp := doWithAuth(t, http.MethodDelete, fmt.Sprintf("/building-restrictions/%d", id), nil)
	defer delResp.Body.Close()
	assert.Equal(t, http.StatusOK, delResp.StatusCode)

	getResp := doWithAuth(t, http.MethodGet, fmt.Sprintf("/building-restrictions/%d", id), nil)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}
