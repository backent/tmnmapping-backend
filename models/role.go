package models

// Role vocabulary for the sales quotation workflow.
// See docs/QUOTATION_FEATURE_ANALYSIS.md §5.3 (in the tmn-mapping root).
const (
	RoleAdmin                 = "admin"
	RoleSales                 = "sales"
	RoleHeadOfSales           = "head_of_sales"
	RoleHeadOfBusinessControl = "head_of_business_control"
	RoleCEO                   = "ceo"
)

// Sales group values, used by the quotation wizard to scope customer visibility.
const (
	SalesGroupSalesTeam     = "sales_team"
	SalesGroupEveryoneSales = "everyone_sales"
	SalesGroupFreelancer    = "freelancer"
)

// Roles is every role the application recognises, in ascending order of authority.
var Roles = []string{
	RoleSales,
	RoleHeadOfSales,
	RoleHeadOfBusinessControl,
	RoleCEO,
	RoleAdmin,
}

// ApproverRoles are the roles that can act on a quotation approval queue.
var ApproverRoles = []string{
	RoleHeadOfSales,
	RoleHeadOfBusinessControl,
	RoleCEO,
}

// legacyRoleAliases maps the pre-Phase-0 vocabulary onto the current one.
// Migration 015 rewrites the column, so these only matter for rows written by
// something outside the app (a manual INSERT, a restored dump) after the migration ran.
var legacyRoleAliases = map[string]string{
	"user":     RoleAdmin,
	"author":   RoleAdmin,
	"approver": RoleHeadOfSales,
	"guest":    RoleSales,
}

// NormalizeRole maps a stored role onto the canonical vocabulary.
//
// An unrecognised role normalises to the empty string rather than to a default,
// so that garbage in the column denies access instead of silently granting it.
func NormalizeRole(role string) string {
	if IsValidRole(role) {
		return role
	}
	if alias, ok := legacyRoleAliases[role]; ok {
		return alias
	}

	return ""
}

// IsValidRole reports whether role is part of the canonical vocabulary.
func IsValidRole(role string) bool {
	for _, known := range Roles {
		if known == role {
			return true
		}
	}

	return false
}

// IsApproverRole reports whether role can approve or return quotations.
func IsApproverRole(role string) bool {
	return HasRole(role, ApproverRoles...)
}

// HasRole reports whether role is one of allowed. An empty role never matches,
// so a request that never passed through RequireAuth cannot satisfy it.
func HasRole(role string, allowed ...string) bool {
	if role == "" {
		return false
	}

	for _, candidate := range allowed {
		if candidate == role {
			return true
		}
	}

	return false
}
