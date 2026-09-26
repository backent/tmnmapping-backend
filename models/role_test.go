package models_test

import (
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/assert"
)

// TestNormalizeRole covers the canonical vocabulary, the legacy aliases kept for
// rows written outside the app, and the deny-by-default behaviour for garbage.
func TestNormalizeRole(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"canonical admin passes through", "admin", models.RoleAdmin},
		{"canonical sales passes through", "sales", models.RoleSales},
		{"canonical head of sales passes through", "head_of_sales", models.RoleHeadOfSales},
		{"canonical business control passes through", "head_of_business_control", models.RoleHeadOfBusinessControl},
		{"canonical ceo passes through", "ceo", models.RoleCEO},

		{"legacy user maps to admin", "user", models.RoleAdmin},
		{"legacy author maps to admin", "author", models.RoleAdmin},
		{"legacy approver maps to head of sales", "approver", models.RoleHeadOfSales},
		{"legacy guest maps to sales", "guest", models.RoleSales},

		{"empty stays empty", "", ""},
		{"unknown role is not granted a default", "wizard", ""},
		{"case sensitive: Admin is not admin", "Admin", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, models.NormalizeRole(tt.input))
		})
	}
}

// TestIsValidRole guards against a typo'd constant silently becoming a real role.
func TestIsValidRole(t *testing.T) {
	for _, role := range models.Roles {
		assert.True(t, models.IsValidRole(role), "%q should be valid", role)
	}

	assert.False(t, models.IsValidRole(""))
	assert.False(t, models.IsValidRole("user"))
	assert.False(t, models.IsValidRole("approver"))
}

func TestHasRole(t *testing.T) {
	assert.True(t, models.HasRole(models.RoleAdmin, models.RoleAdmin))
	assert.True(t, models.HasRole(models.RoleCEO, models.RoleSales, models.RoleCEO))

	assert.False(t, models.HasRole(models.RoleSales, models.RoleAdmin))
	assert.False(t, models.HasRole(models.RoleAdmin), "no allowed roles means no access")

	// An empty role is what NormalizeRole returns for an unrecognised value, and
	// what a request that skipped RequireAuth would carry. It must never match.
	assert.False(t, models.HasRole("", models.RoleAdmin))
	assert.False(t, models.HasRole("", ""))
}

// TestSalesGroupConstants pins the values the users.sales_group CHECK constraint
// in migration 015 accepts.
func TestSalesGroupConstants(t *testing.T) {
	assert.Equal(t, "sales_team", models.SalesGroupSalesTeam)
	assert.Equal(t, "everyone_sales", models.SalesGroupEveryoneSales)
	assert.Equal(t, "freelancer", models.SalesGroupFreelancer)
}
