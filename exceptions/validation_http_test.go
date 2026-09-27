package exceptions_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/libs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mirrors web/brand.CreateBrandRequest closely enough to exercise the real tags.
type brandRequest struct {
	Code         string `json:"code" validate:"required,max=50"`
	Name         string `json:"name" validate:"required,max=255"`
	Status       string `json:"status" validate:"required,oneof=active inactive"`
	ContactEmail string `json:"contact_email" validate:"required,email,max=255"`
	ContactPhone string `json:"contact_phone" validate:"required"`
}

func validateStruct(t *testing.T, value interface{}) error {
	t.Helper()

	return libs.NewValidator().Struct(value)
}

// The message that prompted this: a non-email in the brand form used to surface
// go-playground's own wording, naming a Go struct and a tag.
func TestBadEmailReadsLikeASentence(t *testing.T) {
	err := validateStruct(t, brandRequest{
		Code: "BR-1", Name: "Brand", Status: "active",
		ContactEmail: "budi.example.com", ContactPhone: "0812",
	})
	require.Error(t, err)

	status, body := handle(t, err)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "Contact Email must be a valid email address.", body["data"])

	// None of the library's vocabulary may survive.
	data := body["data"].(string)
	for _, leak := range []string{"CreateBrandRequest", "brandRequest", "ContactEmail", "tag", "Key:", "Error:"} {
		assert.NotContains(t, data, leak)
	}
}

func TestEachTagInUseHasWording(t *testing.T) {
	cases := []struct {
		name    string
		request brandRequest
		want    string
	}{
		{
			"required",
			brandRequest{Name: "Brand", Status: "active", ContactEmail: "a@b.com", ContactPhone: "1"},
			"Code is required.",
		},
		{
			"oneof",
			brandRequest{Code: "B", Name: "Brand", Status: "retired", ContactEmail: "a@b.com", ContactPhone: "1"},
			"Status must be one of: active, inactive.",
		},
		{
			"max",
			brandRequest{Code: str(51), Name: "Brand", Status: "active", ContactEmail: "a@b.com", ContactPhone: "1"},
			"Code must be at most 50 characters.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateStruct(t, testCase.request)
			require.Error(t, err)

			_, body := handle(t, err)
			assert.Equal(t, testCase.want, body["data"])
		})
	}
}

// A form with several blanks should read as several sentences, not one run-on.
func TestSeveralProblemsAreAllReported(t *testing.T) {
	err := validateStruct(t, brandRequest{ContactEmail: "nope"})
	require.Error(t, err)

	_, body := handle(t, err)
	data := body["data"].(string)

	assert.Contains(t, data, "Code is required.")
	assert.Contains(t, data, "Name is required.")
	assert.Contains(t, data, "Contact Email must be a valid email address.")
	assert.Contains(t, data, "Contact Phone is required.")
}

// Extras carries the wire field name so a client can attach each message to the
// input it belongs to rather than dropping them all in one snackbar.
func TestFieldsAreReportedStructurally(t *testing.T) {
	err := validateStruct(t, brandRequest{
		Code: "B", Name: "Brand", Status: "active",
		ContactEmail: "nope", ContactPhone: "1",
	})
	require.Error(t, err)

	_, body := handle(t, err)

	extras, ok := body["extras"].(map[string]interface{})
	require.True(t, ok)

	raw, err2 := json.Marshal(extras["fields"])
	require.NoError(t, err2)

	var fields []exceptions.FieldError
	require.NoError(t, json.Unmarshal(raw, &fields))
	require.Len(t, fields, 1)

	assert.Equal(t, "contact_email", fields[0].Field)
	assert.Equal(t, "Contact Email must be a valid email address.", fields[0].Message)
}

// humanizeField's output only reads well because the validator reports JSON names.
// If that registration is dropped, messages silently regress to "Contactemail".
func TestFieldNamesComeFromJSONTags(t *testing.T) {
	err := validateStruct(t, brandRequest{
		Code: "B", Name: "Brand", Status: "active",
		ContactEmail: "nope", ContactPhone: "1",
	})
	require.Error(t, err)

	_, body := handle(t, err)
	assert.Contains(t, body["data"].(string), "Contact Email")
	assert.NotContains(t, body["data"].(string), "Contactemail")
}

func str(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = 'a'
	}

	return string(out)
}

// handle runs one panic value through the router's panic handler and returns the
// decoded body. Duplicated from the internal test because this file lives in the
// external package: libs imports exceptions, so only exceptions_test can reach both.
func handle(t *testing.T, panicValue interface{}) (int, map[string]interface{}) {
	t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/brands", nil)

	exceptions.RouterPanicHandler(recorder, request, panicValue)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &decoded))

	return recorder.Code, decoded
}
