//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin_HappyPath(t *testing.T) {
	resp, err := http.Post(
		testSuite.server.URL+"/login",
		"application/json",
		mustJSONReader(t, map[string]interface{}{
			"username": "testadmin",
			"password": "testpass123",
		}),
	)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// auth_token cookie must be set
	var authCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" {
			authCookie = c
		}
	}
	require.NotNil(t, authCookie, "auth_token cookie should be set on login")
	assert.NotEmpty(t, authCookie.Value)
	assert.True(t, authCookie.HttpOnly)

	// Response body must contain user info
	wr := DecodeWebResponse(t, resp)
	assert.Equal(t, "OK", wr.Status)

	var data map[string]interface{}
	err = json.Unmarshal(wr.Data, &data)
	require.NoError(t, err)

	user, ok := data["user"].(map[string]interface{})
	require.True(t, ok, "data.user should be an object")
	assert.Equal(t, "testadmin", user["username"])
	assert.Equal(t, "admin", user["role"])
}

func TestLogin_WrongPassword(t *testing.T) {
	resp, err := http.Post(
		testSuite.server.URL+"/login",
		"application/json",
		mustJSONReader(t, map[string]interface{}{
			"username": "testadmin",
			"password": "wrongpassword",
		}),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLogin_UnknownUser(t *testing.T) {
	resp, err := http.Post(
		testSuite.server.URL+"/login",
		"application/json",
		mustJSONReader(t, map[string]interface{}{
			"username": "doesnotexist",
			"password": "anypassword",
		}),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLogin_MissingUsername(t *testing.T) {
	resp, err := http.Post(
		testSuite.server.URL+"/login",
		"application/json",
		mustJSONReader(t, map[string]interface{}{
			"password": "testpass123",
		}),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLogin_MissingPassword(t *testing.T) {
	resp, err := http.Post(
		testSuite.server.URL+"/login",
		"application/json",
		mustJSONReader(t, map[string]interface{}{
			"username": "testadmin",
		}),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLogout_ClearsCookie(t *testing.T) {
	token := MustLogin(t, "testadmin", "testpass123")

	req, err := http.NewRequest(http.MethodPost, testSuite.server.URL+"/logout", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var authCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" {
			authCookie = c
		}
	}
	require.NotNil(t, authCookie, "auth_token cookie should be in logout response")
	assert.Equal(t, -1, authCookie.MaxAge, "cookie MaxAge must be -1 to expire it immediately")
}

func TestCurrentUser_Authenticated(t *testing.T) {
	adminID := seedAdminID(t)
	token := MintToken(t, adminID)

	req := NewAuthRequest(t, http.MethodGet, "/current-user", nil, token)
	resp := Do(t, req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	wr := DecodeWebResponse(t, resp)
	var user map[string]interface{}
	DecodeData(t, wr, &user)

	assert.Equal(t, "testadmin", user["username"])
	assert.Equal(t, "admin", user["role"])
	assert.NotNil(t, user["id"])
}

func TestCurrentUser_NoToken(t *testing.T) {
	req := NewRequest(t, http.MethodGet, "/current-user", nil)
	resp := Do(t, req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCurrentUser_InvalidToken(t *testing.T) {
	req := NewAuthRequest(t, http.MethodGet, "/current-user", nil, "invalid.token.here")
	resp := Do(t, req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// mustJSONReader marshals v to JSON and returns a *bytes.Reader.
func mustJSONReader(t *testing.T, v interface{}) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}
