package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// maximum response body size read from a provider
const maxBodySize = 1024 * 1024

// maximum body size included in log-messages and errors
const maxLoggedBodySize = 2048

var httpClient = &http.Client{Timeout: 15 * time.Second}

func (cfg *OAuthConfig) logger() *slog.Logger {
	l := slog.Default()
	if cfg != nil && cfg.Logger != nil {
		l = cfg.Logger
	}
	if cfg != nil {
		l = l.With("provider", cfg.Provider)
	}
	return l
}

func truncateBody(body []byte) string {
	if len(body) > maxLoggedBodySize {
		return string(body[:maxLoggedBodySize]) + fmt.Sprintf("...(%d bytes truncated)", len(body)-maxLoggedBodySize)
	}
	return string(body)
}

// doJSON executes the request, checks the status-code and decodes the json response into target.
// The response body is logged on failures (non-2xx status or invalid json).
// "step" is a short description used in log-messages and errors (e.g. "token", "userinfo").
func doJSON(cfg *OAuthConfig, step string, req *http.Request, target any) ([]byte, error) {
	log := cfg.logger().With("step", step, "method", req.Method, "url", req.URL.String())

	start := time.Now()
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error("oauth request failed", "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("%s request error: %w", step, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		log.Error("reading oauth response body failed", "status", resp.StatusCode, "error", err)
		return nil, fmt.Errorf("%s response read error: %w", step, err)
	}

	log = log.With("status", resp.StatusCode, "content_type", resp.Header.Get("Content-Type"), "duration", time.Since(start))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error("unexpected oauth response status", "body", truncateBody(body))
		return body, fmt.Errorf("unexpected status-code in %s response: %d", step, resp.StatusCode)
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		log.Error("invalid oauth json response", "error", err, "body", truncateBody(body))
		return body, fmt.Errorf("%s response decode error: %w", step, err)
	}

	// body is intentionally not logged here, it may contain access tokens
	log.Debug("oauth request successful", "body_size", len(body))
	return body, nil
}

// requestAccessToken executes the token request and validates the response
func requestAccessToken(cfg *OAuthConfig, req *http.Request) (string, error) {
	tokenData := AccessTokenResponse{}
	body, err := doJSON(cfg, "token", req, &tokenData)
	if err != nil {
		return "", err
	}

	// some providers (github) respond with 200 and an error object
	if tokenData.Error != "" || tokenData.AccessToken == "" {
		attrs := []any{"error", tokenData.Error, "error_description", tokenData.ErrorDescription}
		if tokenData.AccessToken == "" {
			// only dump the body if it doesn't contain a token
			attrs = append(attrs, "body", truncateBody(body))
		}
		cfg.logger().Error("token response contains no valid access token", attrs...)
		if tokenData.Error != "" {
			return "", fmt.Errorf("token error: %s (%s)", tokenData.Error, tokenData.ErrorDescription)
		}
		return "", fmt.Errorf("empty access token in token response")
	}

	cfg.logger().Debug("access token received", "token_type", tokenData.TokenType, "scope", tokenData.Scope)
	return tokenData.AccessToken, nil
}
