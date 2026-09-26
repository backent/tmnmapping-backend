package exceptions

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// FieldError names one field the request got wrong, in words the person filling
// the form can act on. It rides along under Extras so a client can attach the
// message to the input it belongs to instead of dropping it all in a snackbar.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationMessage turns go-playground/validator's output into sentences.
//
// Left alone, that library reports itself, not the request:
//
//	Key: 'CreateBrandRequest.ContactEmail' Error:Field validation for 'ContactEmail' failed on the 'email' tag
//
// which names a Go struct nobody outside this repo has heard of and buries the one
// useful word ("email") at the end. The same string was being shown to whoever was
// filling in the form.
func ValidationMessage(errs validator.ValidationErrors) (string, []FieldError) {
	fields := make([]FieldError, 0, len(errs))
	sentences := make([]string, 0, len(errs))

	for _, err := range errs {
		label := humanizeField(err.Field())
		message := messageFor(label, err.Tag(), err.Param())

		fields = append(fields, FieldError{Field: err.Field(), Message: message})
		sentences = append(sentences, message)
	}

	return strings.Join(sentences, " "), fields
}

// messageFor covers the tags this codebase actually uses. Anything else falls back
// to a sentence that is vague but still readable, which beats leaking the tag name.
func messageFor(label, tag, param string) string {
	switch tag {
	case "required":
		return label + " is required."
	case "email":
		return label + " must be a valid email address."
	case "max":
		return label + " must be at most " + param + " characters."
	case "min":
		return label + " must be at least " + param + " characters."
	case "len":
		return label + " must be exactly " + param + " characters."
	case "gte":
		return label + " must be " + param + " or more."
	case "gt":
		return label + " must be more than " + param + "."
	case "lte":
		return label + " must be " + param + " or less."
	case "lt":
		return label + " must be less than " + param + "."
	case "oneof":
		return label + " must be one of: " + strings.Join(strings.Fields(param), ", ") + "."
	case "datetime":
		return label + " must be a date in the format " + describeDateLayout(param) + "."
	case "url":
		return label + " must be a valid URL."
	case "numeric":
		return label + " must be a number."
	default:
		return label + " is not valid."
	}
}

// humanizeField turns a wire field name into something that reads like the label
// printed above the input: contact_email -> Contact Email.
//
// It relies on the validator reporting the JSON name rather than the Go field name;
// libs.NewValidator registers that. If that registration is ever removed this still
// produces something readable (ContactEmail -> Contactemail), just less pretty --
// TestFieldNamesComeFromJSONTags guards against the regression.
func humanizeField(field string) string {
	// Slice elements arrive as "Items[0].Price"; the last segment is the one the
	// person can see on screen.
	if idx := strings.LastIndex(field, "."); idx >= 0 {
		field = field[idx+1:]
	}
	if idx := strings.Index(field, "["); idx > 0 {
		field = field[:idx]
	}

	words := strings.FieldsFunc(field, func(r rune) bool { return r == '_' || r == '-' })
	for i, word := range words {
		if word == "" {
			continue
		}
		// Keep initialisms the business uses as they are written.
		if upper := strings.ToUpper(word); upper == "ID" || upper == "IRIS" || upper == "PIC" || upper == "VAT" || upper == "LCD" || upper == "URL" {
			words[i] = upper

			continue
		}
		words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
	}

	if len(words) == 0 {
		return field
	}

	return strings.Join(words, " ")
}

// describeDateLayout rewrites a Go reference layout as the pattern people write
// dates in. "2006-01-02" means nothing to anyone who does not know Go.
func describeDateLayout(layout string) string {
	replacer := strings.NewReplacer(
		"2006", "YYYY",
		"01", "MM",
		"02", "DD",
		"15", "HH",
		"04", "mm",
		"05", "ss",
	)

	described := replacer.Replace(layout)
	if described == "" {
		return layout
	}

	return described
}
