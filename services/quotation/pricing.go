package quotation

import (
	"errors"
	"math"
)

// DefaultTaxRate is Indonesian PPN as it applies to TMN today: 11%, charged on the
// NETT amount, not on gross. Verified against the real quotation template
// (58,520,000 / 532,000,000 = 11.00%).
//
// Each quotation stores its own rate, so an approved quotation keeps the rate it was
// priced at even if this default later changes.
const DefaultTaxRate = 0.11

var (
	ErrDiscountRange = errors.New("discount must be between 0 and 100")
	ErrTaxRateRange  = errors.New("tax rate must be between 0 and 1")
	ErrAmountRange   = errors.New("amount is not a safe whole rupiah value")
)

// maxSafeIdr guards against overflow producing a silently wrong price. Rupiah amounts
// are whole numbers held in int64; anything beyond this is a data error, not a deal.
const maxSafeIdr = int64(1) << 53

// SelectionPricing is one side of a quotation — Placement or Bonus — reduced to what
// pricing needs.
type SelectionPricing struct {
	// GrossPricePerWeek is the sum of the selected resources' weekly rates.
	GrossPricePerWeek int64
	Weeks             int
}

// Gross is the selection's total before any discount.
//
// Rates are per week and multiply straight through: a 4-week campaign at
// 380,000,000/week is 1,520,000,000. There is no division by four anywhere -- that
// belonged to the reference prototype's 4-week rate card, not to this business.
func (s SelectionPricing) Gross() int64 {
	if s.Weeks <= 0 || s.GrossPricePerWeek <= 0 {
		return 0
	}

	return s.GrossPricePerWeek * int64(s.Weeks)
}

// PricingInput is everything CalculatePricing needs. It deliberately takes no
// database handle: pricing is a pure function so it can be tested against the real
// worked example without any infrastructure.
type PricingInput struct {
	Placement SelectionPricing
	Bonus     SelectionPricing

	// Discount is the customer discount in percent, applied to Placement only.
	Discount float64

	// TaxRate is a fraction, e.g. 0.11. Zero is valid and means no tax.
	TaxRate float64
}

// PricingSummary is the full set of figures a quotation stores and the document
// prints.
type PricingSummary struct {
	PlacementGross          int64
	PlacementDiscountAmount int64
	PlacementNet            int64

	BonusGross int64
	BonusNet   int64

	TotalGross int64
	TotalNet   int64

	// EffectiveDiscountAmount and EffectiveDiscountRate describe the whole deal
	// including the free Bonus. They are a commercial-risk and analytics metric.
	//
	// They NEVER determine the approver: routing uses the customer discount. On the
	// real template these read 1,268,000,000 and 70.44% while the customer discount
	// is 65%, which is exactly the pair that must not be confused.
	EffectiveDiscountAmount int64
	EffectiveDiscountRate   float64

	Tax               int64
	TotalIncludingTax int64
}

// CalculatePricing reproduces the model verified against the real quotation:
//
//	placement_gross = rate_per_week * weeks
//	placement_net   = gross - round(gross * discount%)
//	bonus_net       = 0                      // Bonus is always free
//	total_gross     = placement + bonus       // bonus still counts toward gross
//	total_net       = placement_net           // ... but contributes nothing to nett
//	tax             = round(total_net * rate) // on NETT, not gross
func CalculatePricing(input PricingInput) (PricingSummary, error) {
	if input.Discount < 0 || input.Discount > 100 || math.IsNaN(input.Discount) {
		return PricingSummary{}, ErrDiscountRange
	}
	if input.TaxRate < 0 || input.TaxRate > 1 || math.IsNaN(input.TaxRate) {
		return PricingSummary{}, ErrTaxRateRange
	}

	placementGross := input.Placement.Gross()
	bonusGross := input.Bonus.Gross()

	placementDiscount := roundHalfUp(float64(placementGross) * (input.Discount / 100))
	placementNet := placementGross - placementDiscount

	summary := PricingSummary{
		PlacementGross:          placementGross,
		PlacementDiscountAmount: placementDiscount,
		PlacementNet:            placementNet,
		BonusGross:              bonusGross,
		BonusNet:                0,
		TotalGross:              placementGross + bonusGross,
		TotalNet:                placementNet,
	}

	summary.EffectiveDiscountAmount = summary.TotalGross - summary.TotalNet
	if summary.TotalGross > 0 {
		summary.EffectiveDiscountRate = float64(summary.EffectiveDiscountAmount) / float64(summary.TotalGross) * 100
	}

	summary.Tax = roundHalfUp(float64(summary.TotalNet) * input.TaxRate)
	summary.TotalIncludingTax = summary.TotalNet + summary.Tax

	for _, amount := range []int64{
		summary.PlacementGross, summary.PlacementNet, summary.BonusGross,
		summary.TotalGross, summary.TotalNet, summary.Tax, summary.TotalIncludingTax,
	} {
		if amount < 0 || amount > maxSafeIdr {
			return PricingSummary{}, ErrAmountRange
		}
	}

	return summary, nil
}

// roundHalfUp keeps rupiah whole. math.Round is half-away-from-zero, which for the
// non-negative amounts here is half-up.
func roundHalfUp(value float64) int64 {
	return int64(math.Round(value))
}
