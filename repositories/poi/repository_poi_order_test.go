package poi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// orderBy and orderDirection come straight from the query string and are spliced into
// ORDER BY. Before poiSafeOrder, GET /pois?orderBy=<sql> ran that SQL for any
// signed-in role.
func TestPOISafeOrderRefusesAnythingNotWhitelisted(t *testing.T) {
	payloads := []string{
		"brand, (SELECT pg_sleep(5))::text",
		"id; DROP TABLE poi_points",
		"p.brand",
		"",
	}

	for _, payload := range payloads {
		column, _ := poiSafeOrder(payload, "ASC")
		assert.Equal(t, "p.created_at", column, "orderBy %q", payload)
	}

	_, direction := poiSafeOrder("brand", "ASC, (SELECT 1)")
	assert.Equal(t, "DESC", direction)
}

func TestPOISafeOrderKeepsTheSortsTheAppUses(t *testing.T) {
	column, direction := poiSafeOrder("created_at", "DESC")
	assert.Equal(t, "p.created_at", column)
	assert.Equal(t, "DESC", direction)

	column, direction = poiSafeOrder("brand", "ASC")
	assert.Equal(t, "p.brand", column)
	assert.Equal(t, "ASC", direction)
}

func TestPOIAllowedOrderByNamesSelectedColumns(t *testing.T) {
	for key, column := range poiAllowedOrderBy {
		assert.Contains(t, poiSelectCols, column, "orderBy %q", key)
	}
}
