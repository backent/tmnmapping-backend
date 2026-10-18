package quotation

import (
	"strconv"
	"strings"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/assert"
)

func items(n int) []models.QuotationSelectionItem {
	list := make([]models.QuotationSelectionItem, n)
	for i := range list {
		list[i] = models.QuotationSelectionItem{
			BuildingId: i + 1, BuildingName: "B" + strconv.Itoa(i+1),
			UnitPriceIdr: 1_000_000,
		}
	}

	return list
}

// The placeholder numbering is the hazard here. Getting it wrong on the list scope
// broke every scoped query with a 42P18; a batch insert repeats the same risk nine
// columns at a time.
func TestBuildSelectionItemsInsertNumbersEveryPlaceholderContiguously(t *testing.T) {
	SQL, args := buildSelectionItemsInsert(42, items(3))

	assert.Len(t, args, 3*itemColumns)

	for i := 1; i <= len(args); i++ {
		assert.Contains(t, SQL, "$"+strconv.Itoa(i), "placeholder $%d is missing", i)
	}
	assert.NotContains(t, SQL, "$"+strconv.Itoa(len(args)+1), "numbered past the arguments")

	// Every row repeats the selection id as its first column.
	assert.Equal(t, 42, args[0])
	assert.Equal(t, 42, args[itemColumns])
	assert.Equal(t, 42, args[2*itemColumns])
}

func TestBuildSelectionItemsInsertWritesOneTuplePerItem(t *testing.T) {
	SQL, _ := buildSelectionItemsInsert(1, items(4))

	assert.Equal(t, 4, strings.Count(SQL, "("+"$"), "one VALUES tuple per item")
	assert.Equal(t, 1, strings.Count(SQL, "INSERT INTO"), "one statement, not one per row")
}

// A full "select all filtered" is the case this exists for: it must chunk rather
// than send one statement with more parameters than Postgres accepts.
func TestSelectionItemBatchStaysInsidePostgresParameterLimit(t *testing.T) {
	assert.LessOrEqual(t, itemsPerInsert*itemColumns, 65535,
		"a batch must fit Postgres' 65535 parameter cap")

	_, args := buildSelectionItemsInsert(1, items(itemsPerInsert))
	assert.LessOrEqual(t, len(args), 65535)
}

func TestBuildSelectionItemsInsertHandlesAnEmptyBatch(t *testing.T) {
	_, args := buildSelectionItemsInsert(1, nil)

	assert.Empty(t, args)
}
