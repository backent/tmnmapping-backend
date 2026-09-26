package libs

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func NewValidator() *validator.Validate {
	validate := validator.New()

	// Report the JSON name of a field rather than its Go name, so a rejection can
	// be phrased in terms of what the client sent (contact_email) and pointed at
	// the input it came from. Without this the only name available is the struct
	// field, and a message ends up naming internals: "ContactEmail".
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}

		return name
	})

	return validate
}
