package exceptions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumanizeField(t *testing.T) {
	cases := map[string]string{
		"contact_email":   "Contact Email",
		"code":            "Code",
		"project_id_iris": "Project ID IRIS",
		"vat_percent":     "VAT Percent",
		"Items[0].price":  "Price",
		"":                "",
	}

	for input, want := range cases {
		assert.Equal(t, want, humanizeField(input), "input %q", input)
	}
}

func TestDateLayoutIsDescribedInHumanTerms(t *testing.T) {
	assert.Equal(t, "YYYY-MM-DD", describeDateLayout("2006-01-02"))
	assert.Equal(t, "Start Date must be a date in the format YYYY-MM-DD.",
		messageFor("Start Date", "datetime", "2006-01-02"))
}

// An unmapped tag must still produce something readable rather than leaking it.
func TestUnknownTagFallsBackToPlainWording(t *testing.T) {
	assert.Equal(t, "Code is not valid.", messageFor("Code", "somethingnew", ""))
}
