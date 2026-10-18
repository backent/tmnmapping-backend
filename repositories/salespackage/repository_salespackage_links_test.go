package salespackage

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ids(n int) []int {
	list := make([]int, n)
	for i := range list {
		list[i] = i + 1
	}

	return list
}

// Placeholder numbering is the hazard in any multi-row insert: number it wrong and
// Postgres refuses the statement with 42P18, taking the whole save with it.
func TestBuildBuildingLinksInsertNumbersEveryPlaceholderContiguously(t *testing.T) {
	SQL, args := buildBuildingLinksInsert(42, ids(3))

	assert.Len(t, args, 3*linkColumns)

	for i := 1; i <= len(args); i++ {
		assert.Contains(t, SQL, "$"+strconv.Itoa(i), "placeholder $%d is missing", i)
	}
	assert.NotContains(t, SQL, "$"+strconv.Itoa(len(args)+1), "numbered past the arguments")
}

func TestBuildBuildingLinksInsertPairsPackageWithEachBuilding(t *testing.T) {
	_, args := buildBuildingLinksInsert(42, []int{7, 8})

	assert.Equal(t, []interface{}{42, 7, 42, 8}, args)
}

func TestBuildBuildingLinksInsertWritesOneStatement(t *testing.T) {
	SQL, _ := buildBuildingLinksInsert(1, ids(4))

	assert.Equal(t, 1, strings.Count(SQL, "INSERT INTO"), "one statement, not one per row")
	assert.Equal(t, 4, strings.Count(SQL, "($"), "one tuple per building")
}

// The case this exists for: a package filled from a filtered "select all".
func TestBuildingLinkBatchStaysInsidePostgresParameterLimit(t *testing.T) {
	assert.LessOrEqual(t, linksPerInsert*linkColumns, 65535,
		"a batch must fit Postgres' 65535 parameter cap")

	_, args := buildBuildingLinksInsert(1, ids(linksPerInsert))
	assert.LessOrEqual(t, len(args), 65535)
}

// UNIQUE (sales_package_id, building_id) means a repeated id would fail the whole
// insert, so duplicates are dropped before they reach SQL.
func TestDedupeBuildingIdsKeepsFirstOccurrenceAndOrder(t *testing.T) {
	assert.Equal(t, []int{3, 1, 2}, dedupeBuildingIds([]int{3, 1, 3, 2, 1, 3}))
}

func TestDedupeBuildingIdsHandlesEmpty(t *testing.T) {
	assert.Empty(t, dedupeBuildingIds(nil))
}
