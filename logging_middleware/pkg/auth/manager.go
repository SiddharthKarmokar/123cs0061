package auth

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type defaultTokenManager struct {
	client         AuthClient
	cache          *tokenCache
	credentials    AuthRequest
	refreshMu      sync.Mutex // Prevent refresh stampede
	quit           chan struct{}
	wg             sync.WaitGroup
	useStaticToken bool
	staticToken    string
}

// NewTokenManager creates a new manager for handling auth tokens.
func NewTokenManager(client AuthClient, creds AuthRequest, useStatic bool, staticToken string) TokenManager {
	return &defaultTokenManager{
		client:         client,
		cache:          newTokenCache(),
		credentials:    creds,
		quit:           make(chan struct{}),
		useStaticToken: useStatic,
		staticToken:    staticToken,
	}
}

func (m *defaultTokenManager) Start(ctx context.Context) {
	if m.useStaticToken {
		log.Println("DEBUG MODE: Using static token, skipping auth endpoints.")
		return
	}

	// Perform initial synchronous auth, but don't fail the entire app if the auth server is temporarily down
	if _, err := m.ForceRefresh(ctx); err != nil {
		log.Printf("WARNING: failed initial authentication, will retry in background: %v", err)
	}

	m.wg.Add(1)
	go m.refreshLoop(ctx)
}

func (m *defaultTokenManager) Stop() {
	close(m.quit)
	m.wg.Wait()
}

func (m *defaultTokenManager) GetToken(ctx context.Context) (string, error) {
	if m.useStaticToken {
		return m.staticToken, nil
	}

	if token, valid := m.cache.get(); valid {
		return token, nil
	}
	return m.ForceRefresh(ctx)
}

func (m *defaultTokenManager) ForceRefresh(ctx context.Context) (string, error) {
	if m.useStaticToken {
		return m.staticToken, nil
	}

	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	// Double check cache after acquiring lock
	if token, valid := m.cache.get(); valid {
		return token, nil
	}

	// Exponential backoff strategy for sync refresh
	backoff := 500 * time.Millisecond
	maxRetries := 3

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := m.client.Authenticate(ctx, m.credentials)
		if err == nil {
			m.cache.set(resp.Token, resp.ExpiresIn)
			return resp.Token, nil
		}

		lastErr = err
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		}
	}

	return "", fmt.Errorf("auth refresh failed after retries, last error: %w", lastErr)
}

func (m *defaultTokenManager) refreshLoop(ctx context.Context) {
	defer m.wg.Done()

	for {
		sleepDuration := m.cache.timeUntilExpiry()

		// If expired or missing, retry immediately
		if sleepDuration <= 0 {
			sleepDuration = 5 * time.Second
		}

		select {
		case <-time.After(sleepDuration):
			// Proactively refresh
			_, err := m.ForceRefresh(ctx)
			if err != nil {
				// We don't fatal here, GetToken will retry or bubble up errors
				log.Printf("background token refresh failed: %v", err)
			}
		case <-m.quit:
			return
		case <-ctx.Done():
			return
		}
	}
}
