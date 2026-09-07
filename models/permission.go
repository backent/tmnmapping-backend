package models

import "sort"

// Permission keys.
//
// This file is the single place that decides who may do what. Adding a feature
// means adding keys here and listing the roles that hold them — the role
// vocabulary and the route definitions stay untouched.
//
// Naming: "<area>.view" reads, "<area>.manage" writes.
const (
	PermissionDashboardView = "dashboard.view"

	PermissionBuildingsView   = "buildings.view"
	PermissionBuildingsManage = "buildings.manage"

	PermissionMappingView = "mapping.view"

	PermissionPOIsView   = "pois.view"
	PermissionPOIsManage = "pois.manage"

	PermissionSalesPackagesView   = "sales-packages.view"
	PermissionSalesPackagesManage = "sales-packages.manage"

	PermissionBuildingRestrictionsView   = "building-restrictions.view"
	PermissionBuildingRestrictionsManage = "building-restrictions.manage"

	PermissionSavedPolygonsView   = "saved-polygons.view"
	PermissionSavedPolygonsManage = "saved-polygons.manage"

	PermissionMasterDataView   = "master-data.view"
	PermissionMasterDataManage = "master-data.manage"

	PermissionUsersView   = "users.view"
	PermissionUsersManage = "users.manage"

	// Phase 1 advertiser master data. Sales need to read customers and brands to
	// raise a quotation; only admin maintains them.
	PermissionCustomersView   = "customers.view"
	PermissionCustomersManage = "customers.manage"

	PermissionBrandsView   = "brands.view"
	PermissionBrandsManage = "brands.manage"

	PermissionSalesAssignmentsView   = "sales-assignments.view"
	PermissionSalesAssignmentsManage = "sales-assignments.manage"

	// The rate card. Everyone reads it -- a quotation cannot be priced otherwise --
	// but only admin edits a draft, and publishing is its own permission because it
	// changes what every future quotation is priced against.
	PermissionRateCardsView    = "rate-cards.view"
	PermissionRateCardsManage  = "rate-cards.manage"
	PermissionRateCardsPublish = "rate-cards.publish"

	// Navigation-only permissions. No route enforces these; they decide which
	// sections the frontend shows. They live here so that the whole policy is
	// readable in one file, and so the frontend does not need a second copy of it.
	//
	// They exist because some screens are administration tooling even though the
	// data behind them is readable by everyone: the mapping page loads categories,
	// mother brands and restrictions for every role, so the reads must stay open
	// while the management screens stay hidden.
	PermissionMasterDataScreen           = "master-data.screen"
	PermissionBuildingRestrictionsScreen = "building-restrictions.screen"
	PermissionAdvertiserScreen           = "advertiser.screen"
)

// Permissions maps each permission to the roles that hold it.
//
// Every authenticated role can read; only admin can write. The two deliberate
// exceptions are documented inline.
var Permissions = map[string][]string{
	PermissionDashboardView: Roles,

	PermissionBuildingsView:   Roles,
	PermissionBuildingsManage: {RoleAdmin},

	PermissionMappingView: Roles,

	PermissionPOIsView:   Roles,
	PermissionPOIsManage: {RoleAdmin},

	PermissionSalesPackagesView:   Roles,
	PermissionSalesPackagesManage: {RoleAdmin},

	PermissionBuildingRestrictionsView:   Roles,
	PermissionBuildingRestrictionsManage: {RoleAdmin},

	// Saved polygons are a working tool on the map, not master data, and the table
	// is not user-scoped. Restricting writes would break the mapping page for
	// every non-admin.
	PermissionSavedPolygonsView:   Roles,
	PermissionSavedPolygonsManage: Roles,

	PermissionMasterDataView:   Roles,
	PermissionMasterDataManage: {RoleAdmin},

	// Unlike the other list screens, /users has no mapping-page dependency forcing
	// its reads open, so it is admin-only end to end.
	PermissionUsersView:   {RoleAdmin},
	PermissionUsersManage: {RoleAdmin},

	PermissionCustomersView:   Roles,
	PermissionCustomersManage: {RoleAdmin},

	PermissionBrandsView:   Roles,
	PermissionBrandsManage: {RoleAdmin},

	PermissionSalesAssignmentsView:   Roles,
	PermissionSalesAssignmentsManage: {RoleAdmin},

	PermissionRateCardsView:    Roles,
	PermissionRateCardsManage:  {RoleAdmin},
	PermissionRateCardsPublish: {RoleAdmin},

	PermissionMasterDataScreen:           {RoleAdmin},
	PermissionBuildingRestrictionsScreen: {RoleAdmin},
	PermissionAdvertiserScreen:           {RoleAdmin},
}

// RoleCan reports whether role holds permission.
// An empty role, or an unknown permission, holds nothing.
func RoleCan(role string, permission string) bool {
	allowed, ok := Permissions[permission]
	if !ok {
		return false
	}

	return HasRole(role, allowed...)
}

// IsValidPermission reports whether permission is one this application defines.
// RequirePermission uses it to fail at startup on a typo rather than silently
// rejecting every request to that route.
func IsValidPermission(permission string) bool {
	_, ok := Permissions[permission]

	return ok
}

// PermissionsForRole lists everything role holds, sorted for a stable API response.
// The frontend uses this instead of keeping its own copy of the policy.
func PermissionsForRole(role string) []string {
	held := make([]string, 0, len(Permissions))

	for permission, allowed := range Permissions {
		if HasRole(role, allowed...) {
			held = append(held, permission)
		}
	}

	sort.Strings(held)

	return held
}
