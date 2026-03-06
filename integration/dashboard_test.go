//go:build integration

package integration_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardAcquisition_Empty(t *testing.T) {
	truncateTables(t, "acquisitions")

	resp := doWithAuth(t, http.MethodGet, "/dashboard/acquisition", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDashboardAcquisition_WithData(t *testing.T) {
	truncateTables(t, "acquisitions")

	now := time.Now()
	_, err := testSuite.db.Exec(`
		INSERT INTO acquisitions (external_id, workflow_state, acquisition_person, building_project, status, created_at_erp)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		"EXT-001", "active", "Person A", "Project Alpha", "in_progress", now,
	)
	require.NoError(t, err)

	resp := doWithAuth(t, http.MethodGet, "/dashboard/acquisition", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDashboardBuildingProposal_Empty(t *testing.T) {
	truncateTables(t, "building_proposals")

	resp := doWithAuth(t, http.MethodGet, "/dashboard/building-proposal", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDashboardBuildingProposal_WithData(t *testing.T) {
	truncateTables(t, "building_proposals")

	now := time.Now()
	_, err := testSuite.db.Exec(`
		INSERT INTO building_proposals (external_id, workflow_state, acquisition_person, building_project, status, number_of_screen, created_at_erp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		"BP-001", "submitted", "Person B", "Project Beta", "pending", 2, now,
	)
	require.NoError(t, err)

	resp := doWithAuth(t, http.MethodGet, "/dashboard/building-proposal", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDashboardLOI_Empty(t *testing.T) {
	truncateTables(t, "letters_of_intent")

	resp := doWithAuth(t, http.MethodGet, "/dashboard/loi", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDashboardLOI_WithData(t *testing.T) {
	truncateTables(t, "letters_of_intent")

	now := time.Now()
	_, err := testSuite.db.Exec(`
		INSERT INTO letters_of_intent (external_id, workflow_state, acquisition_person, building_project, status, number_of_screen, created_at_erp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		"LOI-001", "signed", "Person C", "Project Gamma", "approved", 5, now,
	)
	require.NoError(t, err)

	resp := doWithAuth(t, http.MethodGet, "/dashboard/loi", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDashboardUnauthenticated(t *testing.T) {
	endpoints := []string{
		"/dashboard/acquisition",
		"/dashboard/building-proposal",
		"/dashboard/loi",
		"/dashboard/building-lcd-presence",
	}
	for _, path := range endpoints {
		req := NewRequest(t, http.MethodGet, path, nil)
		resp := Do(t, req)
		resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 for unauthenticated %s", path)
	}
}
