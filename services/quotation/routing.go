package quotation

import (
	"errors"
	"os"
	"strconv"

	"github.com/malikabdulaziz/tmn-backend/models"
)

// Discount bands and the status each produces.
const (
	StatusDraft                  = "draft"
	StatusPendingManager         = "pending_manager"
	StatusPendingBusinessControl = "pending_business_control"
	StatusPendingCEO             = "pending_ceo"
	StatusReturned               = "returned"
	StatusApproved               = "approved"
)

var QuotationStatuses = []string{
	StatusDraft, StatusPendingManager, StatusPendingBusinessControl,
	StatusPendingCEO, StatusReturned, StatusApproved,
}

// Band names, used for the "who will approve this" hint shown in the wizard.
const (
	BandStandard  = "standard"
	BandElevated  = "elevated"
	BandExecutive = "executive"
)

var (
	ErrNoApproverForRole        = errors.New("no user holds the required approver role")
	ErrAmbiguousApproverForRole = errors.New("more than one user holds the required approver role")
)

// Thresholds are the discount boundaries between approval bands, in percent.
//
// Both are inclusive upper bounds: <= Standard routes to Head of Sales, <= Elevated
// routes to Head of Business Control, anything above goes to the CEO.
//
// Configurable rather than hardcoded because they are a commercial policy, not a
// system rule. Overridable per environment; the defaults are the values in the spec.
type Thresholds struct {
	Standard float64
	Elevated float64
}

// DefaultThresholds are the spec's 65% / 75%.
func DefaultThresholds() Thresholds {
	return Thresholds{
		Standard: envFloat("QUOTATION_DISCOUNT_THRESHOLD_STANDARD", 65),
		Elevated: envFloat("QUOTATION_DISCOUNT_THRESHOLD_ELEVATED", 75),
	}
}

func envFloat(key string, fallback float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value < 0 || value > 100 {
		return fallback
	}

	return value
}

// GetDiscountBand classifies a CUSTOMER discount.
//
// Not the effective discount. On the real quotation template the effective rate is
// 70.44% while the customer discount is 65%, and they fall in different bands -- the
// spec is explicit that only the customer discount routes.
func (t Thresholds) GetDiscountBand(customerDiscount float64) string {
	switch {
	case customerDiscount <= t.Standard:
		return BandStandard
	case customerDiscount <= t.Elevated:
		return BandElevated
	default:
		return BandExecutive
	}
}

// RoleForBand maps a band to the role that approves it.
func RoleForBand(band string) string {
	switch band {
	case BandStandard:
		return models.RoleHeadOfSales
	case BandElevated:
		return models.RoleHeadOfBusinessControl
	default:
		return models.RoleCEO
	}
}

// StatusForRole is the pending status a quotation takes while that role holds it.
func StatusForRole(role string) string {
	switch role {
	case models.RoleHeadOfSales:
		return StatusPendingManager
	case models.RoleHeadOfBusinessControl:
		return StatusPendingBusinessControl
	default:
		return StatusPendingCEO
	}
}

// ApproverDirectory answers "who holds this role". Implemented over the users table,
// which is the choice recorded in the Phase 3 decisions: exactly one user per
// approver role, resolved at submit time so a role change takes effect immediately.
type ApproverDirectory interface {
	FindUsersByRole(role string) ([]models.User, error)
}

// Route is the single atomic result of resolving an approval.
type Route struct {
	Band          string
	ApproverRole  string
	ApproverId    int
	Status        string
	EscalatedFrom string
}

// ResolveApprovalRoute decides who approves a quotation, in one pass.
//
// Two rules, in order:
//
//  1. The customer discount picks a band, and the band picks a role.
//  2. No self-approval. If the resolved approver is the quotation's own commercial
//     owner, it escalates one level -- Head of Sales to Head of Business Control,
//     Head of Business Control to CEO. The spec names this explicitly: "Ayu must
//     never approve her own quotation."
//
// Escalation can chain, so an owner who is also the CEO cannot approve their own
// quotation either; there is simply nobody above, and that is an error rather than a
// silent self-approval.
func ResolveApprovalRoute(
	customerDiscount float64,
	ownerUserId int,
	thresholds Thresholds,
	directory ApproverDirectory,
) (Route, error) {
	if customerDiscount < 0 || customerDiscount > 100 {
		return Route{}, ErrDiscountRange
	}

	band := thresholds.GetDiscountBand(customerDiscount)
	role := RoleForBand(band)
	escalatedFrom := ""

	for {
		approver, err := resolveSingleApprover(directory, role)
		if err != nil {
			return Route{}, err
		}

		if approver.Id != ownerUserId {
			return Route{
				Band:          band,
				ApproverRole:  role,
				ApproverId:    approver.Id,
				Status:        StatusForRole(role),
				EscalatedFrom: escalatedFrom,
			}, nil
		}

		next := escalate(role)
		if next == "" {
			// The owner is the CEO. There is no higher approver, so refuse rather
			// than let a quotation approve itself.
			return Route{}, ErrNoApproverForRole
		}

		escalatedFrom = role
		role = next
	}
}

func escalate(role string) string {
	switch role {
	case models.RoleHeadOfSales:
		return models.RoleHeadOfBusinessControl
	case models.RoleHeadOfBusinessControl:
		return models.RoleCEO
	default:
		return ""
	}
}

// resolveSingleApprover requires exactly one holder of the role. Zero means nobody
// can approve; more than one means the system would be guessing. Both are refused
// loudly at submit time rather than silently picking someone.
func resolveSingleApprover(directory ApproverDirectory, role string) (models.User, error) {
	users, err := directory.FindUsersByRole(role)
	if err != nil {
		return models.User{}, err
	}

	switch len(users) {
	case 0:
		return models.User{}, ErrNoApproverForRole
	case 1:
		return users[0], nil
	default:
		return models.User{}, ErrAmbiguousApproverForRole
	}
}
