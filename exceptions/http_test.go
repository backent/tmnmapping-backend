package exceptions

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// handle runs one panic value through the router's panic handler and gives back
// the decoded response body.
func handle(t *testing.T, panicValue interface{}) (int, map[string]interface{}) {
	t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/customers", nil)

	RouterPanicHandler(recorder, request, panicValue)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))

	return recorder.Code, body
}

// These are real shapes of error text that used to reach the browser verbatim. The
// pgx one is the reason this test exists: its connection errors name the database
// host and the user we connect as.
var leakyErrors = map[string]string{
	"pgx connection":    `failed to connect to ` + "`host=tmn-db.abc123.ap-southeast-1.rds.amazonaws.com user=tmnapp database=tmn`" + `: dial error`,
	"unique violation":  `ERROR: duplicate key value violates unique constraint "unique_customer_code" (SQLSTATE 23505)`,
	"nil dereference":   "runtime error: invalid memory address or nil pointer dereference",
	"sql in the string": "pq: syntax error at or near \"SELECT\" in query SELECT id, code FROM customers",
}

func TestServerError_NeverEchoesTheUnderlyingError(t *testing.T) {
	for name, raw := range leakyErrors {
		t.Run(name, func(t *testing.T) {
			status, body := handle(t, errors.New(raw))

			assert.Equal(t, http.StatusInternalServerError, status)

			data, _ := body["data"].(string)
			assert.NotContains(t, data, raw, "the raw error reached the response body")

			// Spot-check the specific things that must never travel: host names,
			// the database user, constraint names, SQL keywords.
			for _, secret := range []string{"rds.amazonaws.com", "user=tmnapp", "unique_customer_code", "SELECT"} {
				assert.NotContains(t, data, secret)
			}
		})
	}
}

func TestServerError_TellsTheOperatorWhatToDo(t *testing.T) {
	_, body := handle(t, errors.New("boom"))

	data, ok := body["data"].(string)
	require.True(t, ok, "data should be a plain sentence, not an object")

	assert.Contains(t, data, "Something went wrong on our side")
	assert.Contains(t, data, "quote reference")
	assert.Regexp(t, `ERR-[0-9A-F]{6}`, data)
}

// The operator reads the reference off the screen; support greps the log for it.
// If the one in Extras were a second, freshly generated id, that trail would break.
func TestServerError_ReferenceInExtrasMatchesTheOneInTheSentence(t *testing.T) {
	_, body := handle(t, errors.New("boom"))

	extras, ok := body["extras"].(map[string]interface{})
	require.True(t, ok, "extras should carry the reference")

	reference, ok := extras["reference"].(string)
	require.True(t, ok)

	assert.Contains(t, body["data"].(string), reference)
}

func TestServerError_HidesDebugDetailUnlessExplicitlyEnabled(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		_, body := handle(t, errors.New("the raw detail"))

		extras := body["extras"].(map[string]interface{})
		assert.NotContains(t, extras, "debug_detail")
	})

	t.Run("enabled", func(t *testing.T) {
		t.Setenv("DEBUG_ERRORS", "true")

		_, body := handle(t, errors.New("the raw detail"))

		extras := body["extras"].(map[string]interface{})
		assert.Equal(t, "the raw detail", extras["debug_detail"])
	})

	// Anything that is not an explicit "true" must fail closed -- a misspelled or
	// half-configured variable in a new environment must not start disclosing.
	for _, value := range []string{"", "1", "yes", "on", "development", "false"} {
		t.Run("fails closed on "+strconv.Quote(value), func(t *testing.T) {
			t.Setenv("DEBUG_ERRORS", value)

			_, body := handle(t, errors.New("the raw detail"))

			extras := body["extras"].(map[string]interface{})
			assert.NotContains(t, extras, "debug_detail")
		})
	}
}

// A panic that is not an error at all (panic("...") with a bare string) took the
// other branch and used to serialise whatever it was handed.
func TestServerError_NonErrorPanicIsAlsoHidden(t *testing.T) {
	status, body := handle(t, "bare string panic naming host=tmn-db")

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NotContains(t, body["data"].(string), "tmn-db")
	assert.Regexp(t, `ERR-[0-9A-F]{6}`, body["data"].(string))
}

// The whole point is to hide the 500s. The errors we wrote for operators on
// purpose must still arrive word for word.
func TestDeliberateMessagesStillReachTheOperator(t *testing.T) {
	cases := []struct {
		name   string
		panic  interface{}
		status int
		want   string
	}{
		{"bad request", NewBadRequestError("Customer Code is required"), http.StatusBadRequest, "Customer Code is required"},
		{"not found", NewNotFoundError("customer not found"), http.StatusNotFound, "customer not found"},
		{"unauthorized", Unauthorized{Error: "authorization required"}, http.StatusUnauthorized, "authorization required"},
		{"forbidden", ForbiddenError{Error: "you cannot edit this quotation"}, http.StatusForbidden, "you cannot edit this quotation"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			status, body := handle(t, testCase.panic)

			assert.Equal(t, testCase.status, status)
			assert.Equal(t, testCase.want, body["data"])
		})
	}
}

// Sales packages surface duplicate rows out of Extras; a 400 must keep carrying them.
func TestBadRequestKeepsItsExtras(t *testing.T) {
	_, body := handle(t, NewBadRequestWithExtras("Duplicate rows found", map[string]interface{}{
		"duplicates": []interface{}{"row 3"},
	}))

	extras, ok := body["extras"].(map[string]interface{})
	require.True(t, ok)
	assert.NotNil(t, extras["duplicates"])
}

func TestReferencesAreDistinct(t *testing.T) {
	pattern := regexp.MustCompile(`ERR-[0-9A-F]{6}`)
	seen := map[string]bool{}

	for i := 0; i < 200; i++ {
		reference := NewErrorReference()
		require.Regexp(t, pattern, reference)
		seen[reference] = true
	}

	// Collisions are possible in 16M values but 200 draws should not produce a pile
	// of them; this catches a generator that is not actually random.
	assert.Greater(t, len(seen), 190)
}
