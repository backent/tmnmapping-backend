package quotation_test

import (
	"testing"

	"github.com/malikabdulaziz/tmn-backend/services/quotation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A stored price -- a building's or a package's -- buys one base unit: 15 seconds,
// 180 spots a day, one week. Everything here is about what happens when a campaign
// asks for more than that.

func TestAllowedValuesAreTheBaseUnitAndItsMultiples(t *testing.T) {
	assert.Equal(t, []int{15, 30, 45, 60, 75}, quotation.AllowedTvcDurations())
	assert.Equal(t, []int{180, 360, 540, 720, 900}, quotation.AllowedSpots())
}

func TestUnitMultiplier(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		base     int
		expected int
		ok       bool
	}{
		{"the base unit itself", 15, 15, 1, true},
		{"double", 30, 15, 2, true},
		{"the cap", 75, 15, 5, true},
		{"spots base", 180, 180, 1, true},
		{"spots at the cap", 900, 180, 5, true},

		// There is no defensible price for 20 seconds when the rate card sells 15.
		{"not a whole multiple", 20, 15, 0, false},
		{"below the base", 10, 15, 0, false},
		{"beyond the cap", 90, 15, 0, false},
		{"zero", 0, 15, 0, false},
		{"negative", -15, 15, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multiplier, ok := quotation.UnitMultiplier(tt.value, tt.base)

			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.expected, multiplier)
		})
	}
}

// The rule the business asked for: the rate is for 15s / 180 spots / 1 week, and all
// three dimensions multiply through it.
func TestSelectionGross_MultipliesEveryCampaignDimension(t *testing.T) {
	const rate = 1_000_000

	tests := []struct {
		name     string
		weeks    int
		duration int
		spots    int
		expected int64
	}{
		{"the base unit, one week", 1, 15, 180, 1_000_000},
		{"four weeks", 4, 15, 180, 4_000_000},
		{"double duration", 1, 30, 180, 2_000_000},
		{"double spots", 1, 15, 360, 2_000_000},
		{"double both compounds to four", 1, 30, 360, 4_000_000},
		{"everything at the cap", 4, 75, 900, 100_000_000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gross, err := quotation.SelectionGross(rate, tt.weeks, tt.duration, tt.spots)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, gross)
		})
	}
}

// Guards the regression that matters most: the real template is a base-unit campaign,
// so introducing multipliers must not have moved its figures.
func TestSelectionGross_LeavesTheRealTemplateUnchanged(t *testing.T) {
	gross, err := quotation.SelectionGross(380_000_000, 4,
		quotation.BaseTvcDurationSeconds, quotation.BaseSpotsPerDay)

	require.NoError(t, err)
	assert.Equal(t, int64(1_520_000_000), gross, "Package A total gross on the signed document")
}

func TestSelectionGross_RefusesAnOffLadderCampaign(t *testing.T) {
	_, err := quotation.SelectionGross(1_000_000, 4, 20, 180)
	assert.ErrorIs(t, err, quotation.ErrTvcDurationUnit)

	_, err = quotation.SelectionGross(1_000_000, 4, 15, 200)
	assert.ErrorIs(t, err, quotation.ErrSpotsUnit)

	_, err = quotation.SelectionGross(1_000_000, 4, 90, 180)
	assert.ErrorIs(t, err, quotation.ErrTvcDurationUnit, "beyond the 5x cap")
}

// An unpriced resource stays unpriced however long the campaign runs: multiplying
// nothing must never produce something sellable.
func TestSelectionGross_IsZeroWithoutARateOrWeeks(t *testing.T) {
	gross, err := quotation.SelectionGross(0, 4, 60, 720)
	require.NoError(t, err)
	assert.Equal(t, int64(0), gross)

	gross, err = quotation.SelectionGross(1_000_000, 0, 60, 720)
	require.NoError(t, err)
	assert.Equal(t, int64(0), gross)
}

// The error names what the seller may choose, so the wizard can show it verbatim.
func TestUnitErrorsNameTheAllowedValues(t *testing.T) {
	assert.Contains(t, quotation.ErrTvcDurationUnit.Error(), "15 30 45 60 75")
	assert.Contains(t, quotation.ErrSpotsUnit.Error(), "180 360 540 720 900")
}
