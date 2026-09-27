package models_test

import (
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/assert"
)

// The flag decides, except for admin. Both the service that refuses a create and the
// login response that hides the button read this, so they cannot drift apart.
func TestCanRaiseQuotations(t *testing.T) {
	cases := []struct {
		name string
		role string
		flag bool
		want bool
	}{
		{"sales with the flag", models.RoleSales, true, true},
		{"sales without it", models.RoleSales, false, false},
		{"head of sales without it", models.RoleHeadOfSales, false, false},
		{"business control without it", models.RoleHeadOfBusinessControl, false, false},
		{"ceo without it", models.RoleCEO, false, false},
		{"admin without it is still exempt", models.RoleAdmin, false, true},
		{"admin with it", models.RoleAdmin, true, true},
		{"legacy author alias normalises to admin", "author", false, true},
		{"unknown role denies", "nonsense", false, false},
		{"unknown role with the flag still allows", "nonsense", true, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			user := models.User{Role: c.role, CanCreateQuotations: c.flag}

			assert.Equal(t, c.want, user.CanRaiseQuotations())
		})
	}
}
