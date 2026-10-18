package building

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// orderBy and orderDirection come straight from the query string and are spliced into
// ORDER BY. Before buildingSafeOrder, GET /buildings?orderBy=<sql> ran that SQL for any
// signed-in role.
func TestBuildingSafeOrderRefusesAnythingNotWhitelisted(t *testing.T) {
	payloads := []string{
		"(CASE WHEN (SELECT 1)=1 THEN name ELSE id::text END)",
		"name; DROP TABLE buildings",
		"name--",
		"b.name",
		"NAME",
		"password",
		"",
	}

	for _, payload := range payloads {
		column, direction := buildingSafeOrder(payload, "ASC")
		assert.Equal(t, "b.created_at", column, "orderBy %q", payload)
		assert.Equal(t, "ASC", direction)
	}
}

func TestBuildingSafeOrderRefusesAnInjectedDirection(t *testing.T) {
	for _, direction := range []string{"ASC; DROP TABLE buildings", "asc", "DESC NULLS FIRST", ""} {
		_, got := buildingSafeOrder("name", direction)
		assert.Equal(t, "DESC", got, "orderDirection %q", direction)
	}
}

// The values the frontend actually sends must keep working: the buildings page
// defaults to created_at and the building picker sorts by name.
func TestBuildingSafeOrderKeepsTheSortsTheAppUses(t *testing.T) {
	column, direction := buildingSafeOrder("name", "ASC")
	assert.Equal(t, "b.name", column)
	assert.Equal(t, "ASC", direction)

	column, direction = buildingSafeOrder("created_at", "DESC")
	assert.Equal(t, "b.created_at", column)
	assert.Equal(t, "DESC", direction)
}

// Every whitelisted column must be one the select actually reads, or sorting by it
// would be a 500 rather than a sort.
func TestBuildingAllowedOrderByNamesSelectedColumns(t *testing.T) {
	for key, column := range buildingAllowedOrderBy {
		assert.Contains(t, buildingColumns, column, "orderBy %q", key)
	}
}
