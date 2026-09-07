package models_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/assert"
)

// TestPermissions_EveryEntryGrantsSomething catches an entry left empty by a bad
// merge, which would silently lock everyone out of that area.
func TestPermissions_EveryEntryGrantsSomething(t *testing.T) {
	for permission, roles := range models.Permissions {
		assert.NotEmpty(t, roles, "%q grants no roles", permission)

		for _, role := range roles {
			assert.True(t, models.IsValidRole(role),
				"%q lists %q, which is not a valid role", permission, role)
		}
	}
}

// TestPermissions_NamingConvention keeps the keys parseable and consistent, since
// the frontend receives them as opaque strings.
func TestPermissions_NamingConvention(t *testing.T) {
	// .publish exists because publishing a rate card is a different act from editing
	// one: it changes what every future quotation is priced against.
	// .approve exists because acting on an approval is not editing: it is restricted
	// to the approver roles and excludes admin.
	allowedSuffixes := []string{".view", ".manage", ".screen", ".publish", ".approve"}

	for permission := range models.Permissions {
		matched := false
		for _, suffix := range allowedSuffixes {
			if strings.HasSuffix(permission, suffix) {
				matched = true
				break
			}
		}
		assert.True(t, matched, "%q does not end in .view, .manage or .screen", permission)
	}
}

// Writes are admin-only across the board today. If that ever stops being true this
// test should be updated deliberately, not deleted.
func TestPermissions_WritesAreAdminOnly(t *testing.T) {
	exceptions := map[string]bool{
		// Saved polygons are a map working tool, not master data.
		models.PermissionSavedPolygonsManage: true,

		// Quotations are open to every role because the SERVICE scopes them per
		// user: a salesperson only ever sees and edits their own, and an approver
		// only their queue. Gating the endpoint by role instead would stop
		// approvers reading the quotations they have to decide on.
		models.PermissionQuotationsManage: true,
	}

	for permission, roles := range models.Permissions {
		isWrite := strings.HasSuffix(permission, ".manage") || strings.HasSuffix(permission, ".publish")
		if !isWrite || exceptions[permission] {
			continue
		}

		assert.Equal(t, []string{models.RoleAdmin}, roles,
			"%q should be admin-only", permission)
	}
}

func TestRoleCan(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		permission string
		expected   bool
	}{
		{"admin can manage pois", models.RoleAdmin, models.PermissionPOIsManage, true},
		{"sales cannot manage pois", models.RoleSales, models.PermissionPOIsManage, false},
		{"sales can view pois", models.RoleSales, models.PermissionPOIsView, true},
		{"ceo can view the mapping page", models.RoleCEO, models.PermissionMappingView, true},

		// Deliberate: the mapping page's polygon tool is open to everyone.
		{"sales can manage saved polygons", models.RoleSales, models.PermissionSavedPolygonsManage, true},

		// Deliberate: master data reads stay open so the mapping page works,
		// while the management screens do not.
		{"sales can view master data", models.RoleSales, models.PermissionMasterDataView, true},
		{"sales cannot open the master data screens", models.RoleSales, models.PermissionMasterDataScreen, false},

		{"sales cannot view users", models.RoleSales, models.PermissionUsersView, false},
		{"admin can view users", models.RoleAdmin, models.PermissionUsersView, true},

		{"empty role holds nothing", "", models.PermissionPOIsView, false},
		{"unknown role holds nothing", "wizard", models.PermissionPOIsView, false},
		{"unknown permission is never held", models.RoleAdmin, "nonsense.manage", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, models.RoleCan(tt.role, tt.permission))
		})
	}
}

func TestIsValidPermission(t *testing.T) {
	assert.True(t, models.IsValidPermission(models.PermissionBuildingsManage))
	assert.False(t, models.IsValidPermission("buildings.destroy"))
	assert.False(t, models.IsValidPermission(""))
}

func TestPermissionsForRole(t *testing.T) {
	// Admin holds everything except approval. Approval authority follows the sales
	// hierarchy, not system administration -- the same reason admin is excluded from
	// ApproverRoles.
	t.Run("admin holds everything except approval", func(t *testing.T) {
		held := models.PermissionsForRole(models.RoleAdmin)

		assert.Len(t, held, len(models.Permissions)-1)
		assert.NotContains(t, held, models.PermissionQuotationsApprove)
	})

	t.Run("only the approver roles may approve", func(t *testing.T) {
		for _, role := range []string{models.RoleHeadOfSales, models.RoleHeadOfBusinessControl, models.RoleCEO} {
			assert.True(t, models.RoleCan(role, models.PermissionQuotationsApprove), role)
		}

		assert.False(t, models.RoleCan(models.RoleAdmin, models.PermissionQuotationsApprove))
		assert.False(t, models.RoleCan(models.RoleSales, models.PermissionQuotationsApprove))
	})

	t.Run("is sorted, so the API response is stable", func(t *testing.T) {
		held := models.PermissionsForRole(models.RoleSales)

		assert.True(t, sort.StringsAreSorted(held))
	})

	t.Run("sales holds reads but no writes", func(t *testing.T) {
		held := models.PermissionsForRole(models.RoleSales)

		assert.Contains(t, held, models.PermissionBuildingsView)
		assert.Contains(t, held, models.PermissionMappingView)
		assert.Contains(t, held, models.PermissionSavedPolygonsManage)

		assert.NotContains(t, held, models.PermissionBuildingsManage)
		assert.NotContains(t, held, models.PermissionUsersView)
		assert.NotContains(t, held, models.PermissionMasterDataScreen)
	})

	t.Run("an unrecognised role holds nothing", func(t *testing.T) {
		assert.Empty(t, models.PermissionsForRole("wizard"))
		assert.Empty(t, models.PermissionsForRole(""))
	})

	t.Run("every role can read the mapping page", func(t *testing.T) {
		for _, role := range models.Roles {
			assert.Contains(t, models.PermissionsForRole(role), models.PermissionMappingView)
		}
	})
}
