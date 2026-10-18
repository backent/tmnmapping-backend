package brand

import (
	"reflect"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/assert"
)

// A quotation is addressed to a person, and the printed document carries all four
// fields, so a brand without them cannot be quoted usefully.
func TestBrandContactErrorNamesTheFirstMissingField(t *testing.T) {
	tests := []struct {
		name                                              string
		attentionTo, jobTitle, contactPhone, contactEmail string
		expected                                          string
	}{
		{"complete", "Budi", "Director", "+62 812", "budi@example.com", ""},
		{"no name", "", "Director", "+62 812", "budi@example.com", "Attention To is required"},
		{"no title", "Budi", "", "+62 812", "budi@example.com", "Job Title is required"},
		{"no phone", "Budi", "Director", "", "budi@example.com", "Contact Phone is required"},
		{"no email", "Budi", "Director", "+62 812", "", "Contact Email is required"},
		{"nothing at all", "", "", "", "", "Attention To is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected,
				brandContactError(tt.attentionTo, tt.jobTitle, tt.contactPhone, tt.contactEmail))
		})
	}
}

// The import template and the spreadsheet parser read the same column keys, so a
// rename in one place without the other silently drops a column.
func TestTemplateCarriesTheContactColumns(t *testing.T) {
	keys := map[string]bool{}
	required := map[string]bool{}
	for _, column := range TemplateColumns {
		keys[column.Key] = true
		required[column.Key] = column.Required
	}

	for _, key := range []string{"attention_to", "job_title", "contact_phone", "contact_email"} {
		assert.True(t, keys[key], "template is missing the %s column", key)
		assert.True(t, required[key], "%s must be marked required in the template", key)
	}
}

// The bug this guards: the contact columns were added to the model, the SQL and the
// response struct, but not to this mapper — so a brand saved its contact correctly
// and then served it back empty, and every screen looked like the edit had failed.
func TestBrandToResponseCarriesTheContact(t *testing.T) {
	response := brandToResponse(models.Brand{
		Id: 7, Code: "BRND-001", Name: "Kopi Kenangan", Status: "active",
		AttentionTo: "Budi Santoso", JobTitle: "Marketing Director",
		ContactPhone: "+62 812 3456 7890", ContactEmail: "budi@example.com",
	})

	assert.Equal(t, "Budi Santoso", response.AttentionTo)
	assert.Equal(t, "Marketing Director", response.JobTitle)
	assert.Equal(t, "+62 812 3456 7890", response.ContactPhone)
	assert.Equal(t, "budi@example.com", response.ContactEmail)
}

// Every field the response declares should be populated by the mapper; a field added
// to one and not the other is exactly how the contact went missing.
func TestBrandToResponseLeavesNoDeclaredFieldBehind(t *testing.T) {
	full := models.Brand{
		Id: 7, Code: "BRND-001", CustomerId: 3, CustomerCode: "CUST-001",
		CustomerName: "PT Example", Name: "Kopi Kenangan", Category: "Coffee",
		Status: "active", AttentionTo: "Budi", JobTitle: "Director",
		ContactPhone: "+62 812", ContactEmail: "budi@example.com",
		CreatedAt: "2026-01-01", UpdatedAt: "2026-01-02",
	}

	response := brandToResponse(full)

	value := reflect.ValueOf(response)
	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		assert.False(t, value.Field(i).IsZero(),
			"%s is declared on BrandResponse but never populated by brandToResponse", field.Name)
	}
}
