package buildingprice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePrice(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"1100000", 1100000, true},
		{"1,100,000", 1100000, true},
		{"1.100.000", 1100000, true},
		{"Rp 1.100.000", 1100000, true},
		{"IDR 1,100,000", 1100000, true},
		{"1100000.00", 1100000, true},
		{"1100000.0", 1100000, true},
		{" 1100000 ", 1100000, true},
		{"0", 0, true},
		{"1100000.5", 0, false},
		{"-5", 0, false},
		{"", 0, false},
		{"abc", 0, false},
	}

	for _, c := range cases {
		got, err := parsePrice(c.in)
		if c.ok {
			assert.NoError(t, err, c.in)
			assert.Equal(t, c.want, got, c.in)
		} else {
			assert.Error(t, err, c.in)
		}
	}
}

// The business's rate card workbook calls the price "Round Up". It is renamed to the
// canonical header only when the canonical header is not already there.
func TestApplyHeaderAliases(t *testing.T) {
	assert.Equal(t,
		[]string{"IRIS Building ID", "Pick Building Rate", "Price per Week (IDR)"},
		applyHeaderAliases([]string{"IRIS Building ID", "Pick Building Rate", "Round Up"}))

	both := []string{"IRIS Building ID", "Price per Week (IDR)", "Round Up"}
	assert.Equal(t, both, applyHeaderAliases(both), "a sheet with the real header keeps it")
}
