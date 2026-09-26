package buildingproject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/malikabdulaziz/tmn-backend/services/buildingproject"
)

// Matching loosely on the way in and storing canonically is the whole point: a sheet
// written by a person holds "signed" and " Signed ", and rejecting those teaches
// nothing, while storing them verbatim would break every exact-match filter.
func TestCanonicalVocabularyValue_NormalisesCaseAndSpace(t *testing.T) {
	tests := []struct {
		field, input, want string
	}{
		{"contract_status", "signed", "Signed"},
		{"contract_status", "  SIGNED  ", "Signed"},
		{"contract_status", "Under review", "Under Review"},
		{"tmn_project_status", "cancelled", "Cancelled"},
		{"payment_term", "SEMESTERLY", "Semesterly"},
		{"exclusivity", "non-exclusive", "Non-Exclusive"},
		{"doc_type", "pks", "PKS"},
		{"contract_type", "addendum", "Addendum"},
	}

	for _, test := range tests {
		got, ok := buildingproject.CanonicalVocabularyValue(test.field, test.input)
		assert.True(t, ok, "%s=%q should be accepted", test.field, test.input)
		assert.Equal(t, test.want, got)
	}
}

func TestCanonicalVocabularyValue_RejectsUnknown(t *testing.T) {
	_, ok := buildingproject.CanonicalVocabularyValue("contract_status", "Almost Signed")
	assert.False(t, ok)

	_, ok = buildingproject.CanonicalVocabularyValue("doc_type", "Handshake")
	assert.False(t, ok)
}

// Blank is not an unknown value. Every one of these columns is nullable, and an empty
// cell means "not set" -- rejecting it would make a partly-filled project unsaveable.
func TestCanonicalVocabularyValue_BlankIsAccepted(t *testing.T) {
	got, ok := buildingproject.CanonicalVocabularyValue("contract_status", "   ")
	assert.True(t, ok)
	assert.Equal(t, "", got)
}

// Grade and building type are open on purpose: rejecting them would break the first
// time the business adds a category, and neither drives a workflow.
func TestCanonicalVocabularyValue_OpenFieldsPassThrough(t *testing.T) {
	got, ok := buildingproject.CanonicalVocabularyValue("grade", "  Grade AAA  ")
	assert.True(t, ok)
	assert.Equal(t, "Grade AAA", got)

	assert.Nil(t, buildingproject.AllowedValues("grade"),
		"an open field must report no fixed vocabulary, so no dropdown is generated for it")
}

// The error names the field and lists what it accepts. An error that only says
// "invalid" sends the reader back to the spec.
func TestVocabularyError_ListsWhatIsAllowed(t *testing.T) {
	message := buildingproject.VocabularyError("doc_type", "Handshake")

	assert.Contains(t, message, "doc_type")
	assert.Contains(t, message, "Handshake")
	assert.Contains(t, message, "PKS")
	assert.Contains(t, message, "MOU")
	assert.Contains(t, message, "PO")
}

// The form, the importer and the template all read their vocabulary from here. A
// caller that mutates the returned slice must not be able to corrupt the source.
func TestAllowedValues_ReturnsACopy(t *testing.T) {
	first := buildingproject.AllowedValues("doc_type")
	first[0] = "MUTATED"

	second := buildingproject.AllowedValues("doc_type")
	assert.Equal(t, "PKS", second[0])
}

func TestClosedVocabularyFields_IsStable(t *testing.T) {
	assert.Equal(t, []string{
		"contract_status", "contract_type", "doc_type",
		"exclusivity", "payment_term", "tmn_project_status",
	}, buildingproject.ClosedVocabularyFields())
}
