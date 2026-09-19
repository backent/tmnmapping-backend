package quotation

import (
	"errors"
	"fmt"
	"math"
)

// DefaultTaxRate is Indonesian PPN as it applies to TMN today: 11%, charged on the
// NETT amount, not on gross. Verified against the real quotation template
// (58,520,000 / 532,000,000 = 11.00%).
//
// Each quotation stores its own rate, so an approved quotation keeps the rate it was
// priced at even if this default later changes.
const DefaultTaxRate = 0.11

// Campaign base units.
//
// Every price in the system -- a building's weekly rate and a sales package's alike
// -- buys ONE unit: a 15-second spot, shown 180 times per day per screen, for one
// week. A longer spot or a higher frequency is a multiple of that unit and multiplies
// the rate, which is why a 30-second / 360-spot campaign costs four times a
// 15-second / 180-spot one over the same weeks.
//
// These are constants rather than settings because they define what a stored price
// MEANS. Changing one would silently reprice every building and package already in
// the database -- a data migration, not a configuration change.
const (
	BaseTvcDurationSeconds = 15
	BaseSpotsPerDay        = 180

	// MaxUnitMultiplier caps how far a campaign can scale off the base unit.
	MaxUnitMultiplier = 5
)

var (
	ErrDiscountRange = errors.New("discount must be between 0 and 100")
	ErrTaxRateRange  = errors.New("tax rate must be between 0 and 1")
	ErrAmountRange   = errors.New("amount is not a safe whole rupiah value")

	ErrTvcDurationUnit = fmt.Errorf("tvc duration must be one of %v seconds", AllowedTvcDurations())
	ErrSpotsUnit       = fmt.Errorf("spots per day must be one of %v", AllowedSpots())
)

// AllowedTvcDurations and AllowedSpots are the values a campaign may take: the base
// unit and its multiples, up to MaxUnitMultiplier. The wizard offers exactly these,
// and the server refuses anything else rather than pricing a value nobody quoted.
func AllowedTvcDurations() []int { return unitLadder(BaseTvcDurationSeconds) }

func AllowedSpots() []int { return unitLadder(BaseSpotsPerDay) }

func unitLadder(base int) []int {
	ladder := make([]int, 0, MaxUnitMultiplier)
	for multiplier := 1; multiplier <= MaxUnitMultiplier; multiplier++ {
		ladder = append(ladder, base*multiplier)
	}

	return ladder
}

// UnitMultiplier reports how many base units a campaign value buys.
//
// A value that is not a whole multiple of the base, or that exceeds the cap, buys
// nothing: there is no defensible price for 20 seconds when the rate card sells 15.
func UnitMultiplier(value int, base int) (int, bool) {
	if value <= 0 || base <= 0 || value%base != 0 {
		return 0, false
	}

	multiplier := value / base
	if multiplier > MaxUnitMultiplier {
		return 0, false
	}

	return multiplier, true
}

// SelectionGross is one selection's gross, before any discount:
//
//	gross = rate_per_week × weeks × (duration / 15) × (spots / 180)
//
// The rate is what the building or package costs for one base unit for one week, so
// all three campaign dimensions multiply through it.
func SelectionGross(baseRatePerWeek int64, weeks int, durationSeconds int, spots int) (int64, error) {
	durationMultiplier, ok := UnitMultiplier(durationSeconds, BaseTvcDurationSeconds)
	if !ok {
		return 0, ErrTvcDurationUnit
	}

	spotsMultiplier, ok := UnitMultiplier(spots, BaseSpotsPerDay)
	if !ok {
		return 0, ErrSpotsUnit
	}

	if baseRatePerWeek <= 0 || weeks <= 0 {
		return 0, nil
	}

	gross := baseRatePerWeek * int64(weeks) * int64(durationMultiplier) * int64(spotsMultiplier)
	if gross < 0 || gross > maxSafeIdr {
		return 0, ErrAmountRange
	}

	return gross, nil
}

// maxSafeIdr guards against overflow producing a silently wrong price. Rupiah amounts
// are whole numbers held in int64; anything beyond this is a data error, not a deal.
const maxSafeIdr = int64(1) << 53

// SelectionPricing is one side of a quotation — Placement or Bonus — reduced to what
// the totals need: the gross SelectionGross already worked out.
//
// It holds the finished figure rather than the parts, so the campaign multipliers are
// applied in exactly one place. Rebuilding a per-week rate here to multiply again is
// what would double-count them.
type SelectionPricing struct {
	Gross int64
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

	placementGross := input.Placement.Gross
	bonusGross := input.Bonus.Gross

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
