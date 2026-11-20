package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTokenServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
}

func TestRequestAccessToken(t *testing.T) {
	cfg := &OAuthConfig{Provider: ProviderTypeGithub}

	tests := []struct {
		name   string
		status int
		body   string
		token  string
		errMsg string
	}{
		{"ok", 200, `{"access_token":"abc","token_type":"bearer"}`, "abc", ""},
		{"error in 200 response", 200, `{"error":"bad_verification_code","error_description":"expired"}`, "", "token error: bad_verification_code (expired)"},
		{"empty token", 200, `{}`, "", "empty access token in token response"},
		{"bad status", 401, `{"error":"invalid_client"}`, "", "unexpected status-code in token response: 401"},
		{"invalid json", 200, `<html>`, "", "token response decode error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTokenServer(tc.status, tc.body)
			defer srv.Close()

			req, err := http.NewRequest("POST", srv.URL, nil)
			assert.NoError(t, err)

			token, err := requestAccessToken(cfg, req)
			assert.Equal(t, tc.token, token)
			if tc.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errMsg)
			}
		})
	}
}
