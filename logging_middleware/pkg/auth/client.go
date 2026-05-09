package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type defaultAuthClient struct {
	authURL    string
	httpClient *http.Client
}

// NewAuthClient creates a client tailored for evaluating server auth.
func NewAuthClient(baseURL string) AuthClient {
	return &defaultAuthClient{
		authURL: fmt.Sprintf("%s/auth", baseURL),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *defaultAuthClient) Authenticate(ctx context.Context, req AuthRequest) (*AuthResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.authURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("auth request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("auth failed with status %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	// Safety check if evaluation server doesn't return expires_in. Default to 1 hour.
	if authResp.ExpiresIn == 0 {
		authResp.ExpiresIn = 3600
	}

	return &authResp, nil
}
