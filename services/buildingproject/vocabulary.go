package buildingproject

import (
	"fmt"
	"sort"
	"strings"
)

// The vocabularies a building project's columns accept.
//
// This file is the single place that decides what those columns may hold. The form
// populates its dropdowns from here, the importer validates rows against it, and the
// spreadsheet template writes its data-validation lists from it. If the form kept its
// own copy of these lists they would disagree with the importer within a release or
// two, and the disagreement would only surface as a rejected upload.
//
// CLOSED vocabularies reject an unknown value: they describe a defined workflow, and a
// typo in one silently breaks a filter. OPEN ones (grade, building type) accept
// anything, because rejecting would break the first time the business adds a category.
var (
	TmnProjectStatuses = []string{"Active", "Confirmed", "Installation", "Expired", "Cancelled", "Terminated"}
	ContractTypes      = []string{"Initial", "Renewal", "Addendum"}
	PaymentTerms       = []string{"Monthly", "Quarterly", "Semesterly", "Annually"}
	Exclusivities      = []string{"Exclusive", "Non-Exclusive"}
	DocTypes           = []string{"PKS", "MOU", "PO"}
	ContractStatuses   = []string{"Draft", "Under Review", "Signed", "Closed"}

	// Open: suggested in the form and the template, never enforced.
	SuggestedGrades = []string{"Premium", "Grade A", "Grade B", "Standard"}
)

// closedVocabularies maps a column to the values it accepts. Keyed by the model's
// JSON name so an error message names the field the caller sent.
var closedVocabularies = map[string][]string{
	"tmn_project_status": TmnProjectStatuses,
	"contract_type":      ContractTypes,
	"payment_term":       PaymentTerms,
	"exclusivity":        Exclusivities,
	"doc_type":           DocTypes,
	"contract_status":    ContractStatuses,
}

// CanonicalVocabularyValue matches value against a closed vocabulary, ignoring case
// and surrounding space, and returns the canonical spelling.
//
// Matching loosely on the way in is deliberate: a spreadsheet written by a person
// contains "signed" and "  Signed ", and rejecting those teaches nothing. What is
// STORED is always the canonical spelling, so a filter comparing exact strings works.
func CanonicalVocabularyValue(field string, value string) (string, bool) {
	allowed, known := closedVocabularies[field]
	if !known {
		return strings.TrimSpace(value), true
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", true
	}

	for _, candidate := range allowed {
		if strings.EqualFold(candidate, trimmed) {
			return candidate, true
		}
	}

	return "", false
}

// VocabularyError is the message shown for a rejected value. It lists what is allowed,
// because an error that only says "invalid" sends the reader back to the spec.
func VocabularyError(field string, value string) string {
	allowed, known := closedVocabularies[field]
	if !known {
		return fmt.Sprintf("%s: unexpected value %q", field, value)
	}

	return fmt.Sprintf("%s: %q is not one of %s", field, strings.TrimSpace(value), strings.Join(allowed, ", "))
}

// ClosedVocabularyFields lists the columns with a fixed vocabulary, sorted so the
// template and any generated documentation are stable.
func ClosedVocabularyFields() []string {
	fields := make([]string, 0, len(closedVocabularies))
	for field := range closedVocabularies {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	return fields
}

// AllowedValues returns the vocabulary for field, or nil when the field is open.
// The form and the spreadsheet template both build their dropdowns from this.
func AllowedValues(field string) []string {
	allowed, known := closedVocabularies[field]
	if !known {
		return nil
	}

	out := make([]string, len(allowed))
	copy(out, allowed)

	return out
}
