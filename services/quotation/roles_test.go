package quotation_test

import (
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/services/quotation"
	"github.com/stretchr/testify/assert"
)

// TestIsApproverRole documents that admin is deliberately not an approver:
// approval authority follows the sales hierarchy, not system administration.
func TestIsApproverRole(t *testing.T) {
	assert.True(t, quotation.IsApproverRole(models.RoleHeadOfSales))
	assert.True(t, quotation.IsApproverRole(models.RoleHeadOfBusinessControl))
	assert.True(t, quotation.IsApproverRole(models.RoleCEO))

	assert.False(t, quotation.IsApproverRole(models.RoleAdmin))
	assert.False(t, quotation.IsApproverRole(models.RoleSales))
	assert.False(t, quotation.IsApproverRole(""))
}

// The three bands must stay in ascending order — Phase 3's routing resolves an
// approver by walking this list.
func TestApproverRoles_Order(t *testing.T) {
	assert.Equal(t, []string{
		models.RoleHeadOfSales,
		models.RoleHeadOfBusinessControl,
		models.RoleCEO,
	}, quotation.ApproverRoles)
}
