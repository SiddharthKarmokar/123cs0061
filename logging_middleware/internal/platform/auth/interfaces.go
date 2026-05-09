package auth

import (
	"context"
)

// TokenManager orchestrates token retrieval, caching, and background refreshing.
type TokenManager interface {
	// GetToken returns a valid Bearer token, blocking if a refresh is currently in progress.
	GetToken(ctx context.Context) (string, error)
	// ForceRefresh forces a synchronous token refresh (useful on 401s).
	ForceRefresh(ctx context.Context) (string, error)
	// Start initializes background refresh loops.
	Start(ctx context.Context)
	// Stop gracefully halts background refreshes.
	Stop()
}

// AuthClient handles the low-level HTTP communication with the auth server.
type AuthClient interface {
	Authenticate(ctx context.Context, req AuthRequest) (*AuthResponse, error)
}
