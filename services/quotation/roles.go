// Package quotation holds the sales quotation feature. Pricing and approval
// routing land here in Phase 3; see docs/PHASE_0_ROLES_PROGRESS.md.
package quotation

import "github.com/malikabdulaziz/tmn-backend/models"

// ApproverRoles are the roles that can act on a quotation approval queue, in
// ascending order of the discount band they cover.
//
// This lives here rather than in models/role.go on purpose: "approval queue" is a
// quotation concept, and the shared role vocabulary should not have to know about
// it. models/role.go describes the org chart; this describes what one feature does
// with it.
//
// admin is deliberately absent — approval authority follows the sales hierarchy,
// not system administration.
var ApproverRoles = []string{
	models.RoleHeadOfSales,
	models.RoleHeadOfBusinessControl,
	models.RoleCEO,
}

// IsApproverRole reports whether role can approve or return quotations.
func IsApproverRole(role string) bool {
	return models.HasRole(role, ApproverRoles...)
}
