//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSalesPackageCreate_HappyPath(t *testing.T) {
	truncateTables(t, "sales_packages")
	bID := insertBuilding(t, "SP Building A")

	body := map[string]interface{}{
		"name":         "Package Alpha",
		"building_ids": []int{bID},
	}
	resp := doWithAuth(t, http.MethodPost, "/sales-packages", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)

	assert.Equal(t, "Package Alpha", data["name"])
	buildings, ok := data["buildings"].([]interface{})
	require.True(t, ok)
	assert.Len(t, buildings, 1)
}

func TestSalesPackageCreate_Unauthenticated(t *testing.T) {
	req := NewRequest(t, http.MethodPost, "/sales-packages", map[string]interface{}{
		"name":         "X",
		"building_ids": []int{1},
	})
	resp := Do(t, req)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSalesPackageFindAll(t *testing.T) {
	truncateTables(t, "sales_packages")
	bID := insertBuilding(t, "SP Building B")

	for i := 1; i <= 2; i++ {
		resp := doWithAuth(t, http.MethodPost, "/sales-packages", map[string]interface{}{
			"name":         fmt.Sprintf("Package %d", i),
			"building_ids": []int{bID},
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	resp := doWithAuth(t, http.MethodGet, "/sales-packages", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var list []interface{}
	DecodeData(t, wr, &list)
	assert.Len(t, list, 2)
}

func TestSalesPackageFindById_HappyPath(t *testing.T) {
	truncateTables(t, "sales_packages")
	bID := insertBuilding(t, "SP Building C")

	createResp := doWithAuth(t, http.MethodPost, "/sales-packages", map[string]interface{}{
		"name": "FindablePackage", "building_ids": []int{bID},
	})
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	resp := doWithAuth(t, http.MethodGet, fmt.Sprintf("/sales-packages/%d", id), nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)
	assert.Equal(t, "FindablePackage", data["name"])
}

func TestSalesPackageFindById_NotFound(t *testing.T) {
	resp := doWithAuth(t, http.MethodGet, "/sales-packages/9999999", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSalesPackageUpdate_HappyPath(t *testing.T) {
	truncateTables(t, "sales_packages")
	bID := insertBuilding(t, "SP Building D")

	createResp := doWithAuth(t, http.MethodPost, "/sales-packages", map[string]interface{}{
		"name": "OriginalPackage", "building_ids": []int{bID},
	})
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	updateResp := doWithAuth(t, http.MethodPut, fmt.Sprintf("/sales-packages/%d", id), map[string]interface{}{
		"name": "RenamedPackage", "building_ids": []int{bID},
	})
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	wr := DecodeWebResponse(t, updateResp)
	var data map[string]interface{}
	DecodeData(t, wr, &data)
	assert.Equal(t, "RenamedPackage", data["name"])
}

func TestSalesPackageDelete_HappyPath(t *testing.T) {
	truncateTables(t, "sales_packages")
	bID := insertBuilding(t, "SP Building E")

	createResp := doWithAuth(t, http.MethodPost, "/sales-packages", map[string]interface{}{
		"name": "ToDeletePackage", "building_ids": []int{bID},
	})
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	createWr := DecodeWebResponse(t, createResp)
	var created map[string]interface{}
	DecodeData(t, createWr, &created)
	id := int(created["id"].(float64))

	delResp := doWithAuth(t, http.MethodDelete, fmt.Sprintf("/sales-packages/%d", id), nil)
	defer delResp.Body.Close()
	assert.Equal(t, http.StatusOK, delResp.StatusCode)

	getResp := doWithAuth(t, http.MethodGet, fmt.Sprintf("/sales-packages/%d", id), nil)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}
