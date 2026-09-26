package quotation_test

import (
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/services/quotation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubDirectory answers "who holds this role" from a fixed map.
type stubDirectory struct {
	byRole map[string][]models.User
	err    error
}

func (s stubDirectory) FindUsersByRole(role string) ([]models.User, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.byRole[role], nil
}

// A full directory: one person per approver role, which is the model chosen for
// Phase 3.
func fullDirectory() stubDirectory {
	return stubDirectory{byRole: map[string][]models.User{
		models.RoleHeadOfSales:           {{Id: 10, Username: "ayu", Role: models.RoleHeadOfSales}},
		models.RoleHeadOfBusinessControl: {{Id: 20, Username: "april", Role: models.RoleHeadOfBusinessControl}},
		models.RoleCEO:                   {{Id: 30, Username: "thomas", Role: models.RoleCEO}},
	}}
}

func TestGetDiscountBand_Boundaries(t *testing.T) {
	th := quotation.DefaultThresholds()
	require.Equal(t, float64(65), th.Standard)
	require.Equal(t, float64(75), th.Elevated)

	tests := []struct {
		discount float64
		band     string
	}{
		{0, quotation.BandStandard},
		{64.99, quotation.BandStandard},
		{65, quotation.BandStandard},    // inclusive upper bound
		{65.01, quotation.BandElevated}, // just over
		{75, quotation.BandElevated},    // inclusive upper bound
		{75.01, quotation.BandExecutive},
		{100, quotation.BandExecutive},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.band, th.GetDiscountBand(tt.discount), "discount %.2f%%", tt.discount)
		})
	}
}

func TestResolveApprovalRoute_RoutesByBand(t *testing.T) {
	th := quotation.DefaultThresholds()
	dir := fullDirectory()

	tests := []struct {
		discount float64
		role     string
		approver int
		status   string
	}{
		{50, models.RoleHeadOfSales, 10, quotation.StatusPendingManager},
		{65, models.RoleHeadOfSales, 10, quotation.StatusPendingManager},
		{70, models.RoleHeadOfBusinessControl, 20, quotation.StatusPendingBusinessControl},
		{75, models.RoleHeadOfBusinessControl, 20, quotation.StatusPendingBusinessControl},
		{80, models.RoleCEO, 30, quotation.StatusPendingCEO},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			route, err := quotation.ResolveApprovalRoute(tt.discount, 999, th, dir)
			require.NoError(t, err)
			assert.Equal(t, tt.role, route.ApproverRole)
			assert.Equal(t, tt.approver, route.ApproverId)
			assert.Equal(t, tt.status, route.Status)
			assert.Empty(t, route.EscalatedFrom)
		})
	}
}

// The real quotation carries a 65% customer discount and a 70.44% effective rate.
// Routing must use 65% and go to Head of Sales -- the effective figure is analytics.
func TestResolveApprovalRoute_UsesTheCustomerDiscountNotTheEffectiveOne(t *testing.T) {
	th := quotation.DefaultThresholds()

	onCustomerDiscount, err := quotation.ResolveApprovalRoute(65, 999, th, fullDirectory())
	require.NoError(t, err)
	assert.Equal(t, models.RoleHeadOfSales, onCustomerDiscount.ApproverRole)

	// Had it routed on the effective 70.44%, it would have landed here instead.
	onEffective, err := quotation.ResolveApprovalRoute(70.44, 999, th, fullDirectory())
	require.NoError(t, err)
	assert.Equal(t, models.RoleHeadOfBusinessControl, onEffective.ApproverRole)
	assert.NotEqual(t, onCustomerDiscount.ApproverRole, onEffective.ApproverRole,
		"the two figures route differently, which is why only one of them may be used")
}

// "Ayu must never approve her own quotation."
func TestResolveApprovalRoute_EscalatesRatherThanSelfApprove(t *testing.T) {
	th := quotation.DefaultThresholds()

	// Head of Sales owns a 50% quotation: escalates to Business Control.
	route, err := quotation.ResolveApprovalRoute(50, 10, th, fullDirectory())
	require.NoError(t, err)
	assert.Equal(t, models.RoleHeadOfBusinessControl, route.ApproverRole)
	assert.Equal(t, 20, route.ApproverId)
	assert.Equal(t, quotation.StatusPendingBusinessControl, route.Status)
	assert.Equal(t, models.RoleHeadOfSales, route.EscalatedFrom)

	// Business Control owns a 70% quotation: escalates to CEO.
	route, err = quotation.ResolveApprovalRoute(70, 20, th, fullDirectory())
	require.NoError(t, err)
	assert.Equal(t, models.RoleCEO, route.ApproverRole)
	assert.Equal(t, 30, route.ApproverId)
}

// Escalation chains: if the owner holds both lower roles' seats it keeps climbing.
func TestResolveApprovalRoute_EscalationChains(t *testing.T) {
	dir := stubDirectory{byRole: map[string][]models.User{
		models.RoleHeadOfSales:           {{Id: 10}},
		models.RoleHeadOfBusinessControl: {{Id: 10}}, // same person holds both
		models.RoleCEO:                   {{Id: 30}},
	}}

	route, err := quotation.ResolveApprovalRoute(50, 10, quotation.DefaultThresholds(), dir)
	require.NoError(t, err)
	assert.Equal(t, models.RoleCEO, route.ApproverRole)
	assert.Equal(t, 30, route.ApproverId)
}

// There is nobody above the CEO, so a CEO-owned executive-band quotation must fail
// rather than quietly approve itself.
func TestResolveApprovalRoute_CeoCannotApproveTheirOwn(t *testing.T) {
	_, err := quotation.ResolveApprovalRoute(90, 30, quotation.DefaultThresholds(), fullDirectory())
	assert.ErrorIs(t, err, quotation.ErrNoApproverForRole)
}

// Nobody holds the role: refuse at submit rather than create an unapprovable
// quotation. This is the state the system is in today, with every user an admin.
func TestResolveApprovalRoute_NoApproverForRole(t *testing.T) {
	empty := stubDirectory{byRole: map[string][]models.User{}}

	_, err := quotation.ResolveApprovalRoute(50, 999, quotation.DefaultThresholds(), empty)
	assert.ErrorIs(t, err, quotation.ErrNoApproverForRole)
}

// Two holders means the system would be guessing. Refuse loudly.
func TestResolveApprovalRoute_AmbiguousApprover(t *testing.T) {
	dir := stubDirectory{byRole: map[string][]models.User{
		models.RoleHeadOfSales: {{Id: 10}, {Id: 11}},
	}}

	_, err := quotation.ResolveApprovalRoute(50, 999, quotation.DefaultThresholds(), dir)
	assert.ErrorIs(t, err, quotation.ErrAmbiguousApproverForRole)
}

func TestResolveApprovalRoute_RejectsOutOfRangeDiscount(t *testing.T) {
	_, err := quotation.ResolveApprovalRoute(-1, 1, quotation.DefaultThresholds(), fullDirectory())
	assert.ErrorIs(t, err, quotation.ErrDiscountRange)

	_, err = quotation.ResolveApprovalRoute(101, 1, quotation.DefaultThresholds(), fullDirectory())
	assert.ErrorIs(t, err, quotation.ErrDiscountRange)
}

// Thresholds are commercial policy, not a system rule, so they are configurable.
func TestThresholds_AreConfigurable(t *testing.T) {
	custom := quotation.Thresholds{Standard: 50, Elevated: 60}

	assert.Equal(t, quotation.BandStandard, custom.GetDiscountBand(50))
	assert.Equal(t, quotation.BandElevated, custom.GetDiscountBand(55))
	assert.Equal(t, quotation.BandExecutive, custom.GetDiscountBand(65),
		"65%% is standard under the default thresholds but executive under these")
}

func TestDefaultThresholds_ReadFromEnvironment(t *testing.T) {
	t.Setenv("QUOTATION_DISCOUNT_THRESHOLD_STANDARD", "55")
	t.Setenv("QUOTATION_DISCOUNT_THRESHOLD_ELEVATED", "70")

	th := quotation.DefaultThresholds()
	assert.Equal(t, float64(55), th.Standard)
	assert.Equal(t, float64(70), th.Elevated)

	// A nonsense value falls back rather than silently routing everything to the CEO.
	t.Setenv("QUOTATION_DISCOUNT_THRESHOLD_STANDARD", "not-a-number")
	assert.Equal(t, float64(65), quotation.DefaultThresholds().Standard)
}
