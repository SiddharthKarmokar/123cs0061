package auth

import (
	"fmt"
	"net/http"
)

// Transport wraps an http.RoundTripper, automatically injecting bearer tokens.
type Transport struct {
	Base    http.RoundTripper
	Manager TokenManager
}

// RoundTrip executes a single HTTP transaction, handling auth injection and 401 retries.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.Manager.GetToken(req.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Clone request to avoid mutating the original
	reqWithAuth := req.Clone(req.Context())
	reqWithAuth.Header.Set("Authorization", "Bearer "+token)

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	resp, err := base.RoundTrip(reqWithAuth)
	if err != nil {
		return nil, err
	}

	// Retry exactly once on 401 Unauthorized
	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()

		newToken, refreshErr := t.Manager.ForceRefresh(req.Context())
		if refreshErr != nil {
			return nil, fmt.Errorf("failed to refresh token on 401: %w", refreshErr)
		}

		reqRetry := req.Clone(req.Context())
		reqRetry.Header.Set("Authorization", "Bearer "+newToken)

		return base.RoundTrip(reqRetry)
	}

	return resp, nil
}
