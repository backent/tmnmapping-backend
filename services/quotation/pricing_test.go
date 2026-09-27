package quotation_test

import (
	"testing"

	"github.com/malikabdulaziz/tmn-backend/services/quotation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// baseUnit prices a selection at the base campaign unit: 15 seconds, 180 spots a day,
// for the given weeks. Every expectation written before campaign multipliers existed
// is a base-unit campaign, so these must come out unchanged.
func baseUnit(t *testing.T, ratePerWeek int64, weeks int) quotation.SelectionPricing {
	t.Helper()

	gross, err := quotation.SelectionGross(ratePerWeek, weeks,
		quotation.BaseTvcDurationSeconds, quotation.BaseSpotsPerDay)
	require.NoError(t, err)

	return quotation.SelectionPricing{Gross: gross}
}

// The real TMN quotation template, reproduced exactly.
// See docs/QUOTATION_DOCUMENT_ANALYSIS.md §1. Every figure below is read off that
// document, not invented, which makes this the most valuable test in the package.
func TestCalculatePricing_MatchesTheRealQuotationTemplate(t *testing.T) {
	summary, err := quotation.CalculatePricing(quotation.PricingInput{
		Placement: baseUnit(t, 380_000_000, 4),
		Bonus:     baseUnit(t, 70_000_000, 4),
		Discount:  65,
		TaxRate:   quotation.DefaultTaxRate,
	})
	require.NoError(t, err)

	assert.Equal(t, int64(1_520_000_000), summary.PlacementGross, "Package A total gross")
	assert.Equal(t, int64(988_000_000), summary.PlacementDiscountAmount)
	assert.Equal(t, int64(532_000_000), summary.PlacementNet, "Package A total nett")
	assert.Equal(t, int64(280_000_000), summary.BonusGross, "Bonus total gross")
	assert.Equal(t, int64(0), summary.BonusNet, "Bonus is FREE")
	assert.Equal(t, int64(1_800_000_000), summary.TotalGross, "TOTAL gross")
	assert.Equal(t, int64(532_000_000), summary.TotalNet, "Total Nett")
	assert.Equal(t, int64(1_268_000_000), summary.EffectiveDiscountAmount, "Saving Value")
	assert.InDelta(t, 70.44, summary.EffectiveDiscountRate, 0.01, "Total Discount")
	assert.Equal(t, int64(58_520_000), summary.Tax, "VAT")
	assert.Equal(t, int64(590_520_000), summary.TotalIncludingTax, "Total (VAT included)")
}

// Rates are per week and multiply straight through. The prototype divided by four
// because its rate card was a 4-week rate; this business's is not.
func TestCalculatePricing_GrossIsRatePerWeekTimesWeeks(t *testing.T) {
	tests := []struct {
		name     string
		rate     int64
		weeks    int
		expected int64
	}{
		{"one week", 1_000_000, 1, 1_000_000},
		{"four weeks", 380_000_000, 4, 1_520_000_000},
		{"eight weeks", 250_000, 8, 2_000_000},
		{"zero weeks is zero", 1_000_000, 0, 0},
		{"negative weeks is zero", 1_000_000, -3, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, err := quotation.CalculatePricing(quotation.PricingInput{
				Placement: baseUnit(t, tt.rate, tt.weeks),
			})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, summary.PlacementGross)
		})
	}
}

// The discount applies to Placement only. Bonus stays free and still inflates gross,
// which is what drags the effective rate above the customer discount.
func TestCalculatePricing_BonusIsFreeButCountsTowardGross(t *testing.T) {
	summary, err := quotation.CalculatePricing(quotation.PricingInput{
		Placement: baseUnit(t, 100_000_000, 4),
		Bonus:     baseUnit(t, 100_000_000, 4),
		Discount:  50,
	})
	require.NoError(t, err)

	assert.Equal(t, int64(200_000_000), summary.PlacementNet)
	assert.Equal(t, int64(0), summary.BonusNet)
	assert.Equal(t, int64(800_000_000), summary.TotalGross, "bonus counts toward gross")
	assert.Equal(t, int64(200_000_000), summary.TotalNet, "bonus contributes nothing to nett")

	// 50% customer discount, but 75% effective once the free bonus is included.
	assert.InDelta(t, 75.0, summary.EffectiveDiscountRate, 0.001)
}

func TestCalculatePricing_WithoutBonus(t *testing.T) {
	summary, err := quotation.CalculatePricing(quotation.PricingInput{
		Placement: baseUnit(t, 100_000_000, 4),
		Discount:  25,
	})
	require.NoError(t, err)

	assert.Equal(t, int64(400_000_000), summary.TotalGross)
	assert.Equal(t, int64(300_000_000), summary.TotalNet)
	// With no bonus the effective rate equals the customer discount.
	assert.InDelta(t, 25.0, summary.EffectiveDiscountRate, 0.001)
}

// Tax is charged on NETT, not on gross. Getting this backwards on the real example
// would overcharge by 139,480,000.
func TestCalculatePricing_TaxIsChargedOnNett(t *testing.T) {
	summary, err := quotation.CalculatePricing(quotation.PricingInput{
		Placement: baseUnit(t, 100_000_000, 4),
		Discount:  50,
		TaxRate:   0.11,
	})
	require.NoError(t, err)

	assert.Equal(t, int64(200_000_000), summary.TotalNet)
	assert.Equal(t, int64(22_000_000), summary.Tax, "11% of nett, not of the 400m gross")
}

func TestCalculatePricing_ZeroTaxRateIsValid(t *testing.T) {
	summary, err := quotation.CalculatePricing(quotation.PricingInput{
		Placement: baseUnit(t, 1_000_000, 1),
		TaxRate:   0,
	})
	require.NoError(t, err)

	assert.Equal(t, int64(0), summary.Tax)
	assert.Equal(t, summary.TotalNet, summary.TotalIncludingTax)
}

// The boundaries that decide an approval band must be exact, so the arithmetic at
// each one is pinned.
func TestCalculatePricing_DiscountBoundaries(t *testing.T) {
	tests := []struct {
		discount float64
		net      int64
	}{
		{0, 1_000_000_000},
		{65, 350_000_000},
		{65.01, 349_900_000},
		{75, 250_000_000},
		{75.01, 249_900_000},
		{100, 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			summary, err := quotation.CalculatePricing(quotation.PricingInput{
				Placement: baseUnit(t, 250_000_000, 4),
				Discount:  tt.discount,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.net, summary.PlacementNet, "discount %.2f%%", tt.discount)
		})
	}
}

func TestCalculatePricing_RejectsOutOfRangeInput(t *testing.T) {
	_, err := quotation.CalculatePricing(quotation.PricingInput{Discount: -1})
	assert.ErrorIs(t, err, quotation.ErrDiscountRange)

	_, err = quotation.CalculatePricing(quotation.PricingInput{Discount: 101})
	assert.ErrorIs(t, err, quotation.ErrDiscountRange)

	_, err = quotation.CalculatePricing(quotation.PricingInput{TaxRate: 1.5})
	assert.ErrorIs(t, err, quotation.ErrTaxRateRange)

	_, err = quotation.CalculatePricing(quotation.PricingInput{TaxRate: -0.1})
	assert.ErrorIs(t, err, quotation.ErrTaxRateRange)
}

// An empty quotation prices to zero rather than dividing by zero on the effective
// rate.
func TestCalculatePricing_EmptyQuotation(t *testing.T) {
	summary, err := quotation.CalculatePricing(quotation.PricingInput{Discount: 50})
	require.NoError(t, err)

	assert.Equal(t, int64(0), summary.TotalGross)
	assert.Equal(t, float64(0), summary.EffectiveDiscountRate)
}

// A price beyond safe integer range is a data error, not a deal worth billions of
// billions.
func TestCalculatePricing_RejectsOverflowingAmounts(t *testing.T) {
	_, err := quotation.CalculatePricing(quotation.PricingInput{
		Placement: quotation.SelectionPricing{Gross: 1 << 54},
	})
	assert.ErrorIs(t, err, quotation.ErrAmountRange)
}

// The multipliers are the new way to overflow: a large rate times weeks times both
// campaign multipliers has to be caught where it is computed.
func TestSelectionGross_RejectsOverflowingAmounts(t *testing.T) {
	_, err := quotation.SelectionGross(1<<52, 52, 75, 900)

	assert.ErrorIs(t, err, quotation.ErrAmountRange)
}

// ---------------------------------------------------------------------------
// Package pricing moved onto the package (migration 020)
// ---------------------------------------------------------------------------

// A package priced at zero is unpriced, not free. Quoting it would sell the whole
// package for nothing, which is exactly the mistake the building path already
// refuses -- see "has no price in the published rate card".
func TestPackageGrossUsesTheWeeklyRateTimesWeeks(t *testing.T) {
	// 112,500,000/week over 4 weeks, the shape a real package quotation takes.
	selection := baseUnit(t, 112_500_000, 4)

	assert.Equal(t, int64(450_000_000), selection.Gross)
}

func TestPackageGrossIsZeroWhenUnpriced(t *testing.T) {
	selection := baseUnit(t, 0, 4)

	assert.Equal(t, int64(0), selection.Gross,
		"an unpriced package must not produce a sellable gross")
}
