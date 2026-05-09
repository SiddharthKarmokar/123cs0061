package auth

import (
	"testing"
	"time"
)

func TestTokenCache_SetAndGet(t *testing.T) {
	cache := newTokenCache()

	// Initially empty
	if _, ok := cache.get(); ok {
		t.Errorf("expected cache to be empty initially")
	}

	// Set valid token
	cache.set("valid_token", 3600)

	token, ok := cache.get()
	if !ok || token != "valid_token" {
		t.Errorf("expected to retrieve valid_token, got %s (ok: %v)", token, ok)
	}
}

func TestTokenCache_Expiry(t *testing.T) {
	cache := newTokenCache()

	// Set token with very short expiry (1 second, meaning 0.9 second safe duration)
	cache.set("short_lived_token", 1)

	// Wait for expiry
	time.Sleep(1 * time.Second)

	if _, ok := cache.get(); ok {
		t.Errorf("expected token to be expired")
	}
}

func TestTokenCache_TimeUntilExpiry(t *testing.T) {
	cache := newTokenCache()

	if d := cache.timeUntilExpiry(); d != 0 {
		t.Errorf("expected 0 for empty cache, got %v", d)
	}

	cache.set("token", 100) // 90 seconds safe

	d := cache.timeUntilExpiry()
	if d <= 0 || d > 90*time.Second {
		t.Errorf("expected time until expiry to be > 0 and <= 90s, got %v", d)
	}
}
