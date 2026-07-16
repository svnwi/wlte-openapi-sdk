package wlteopenapi

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type tokenRequesterFunc func(context.Context, string, string, map[string]string, map[string]string, any, any) error

func (fn tokenRequesterFunc) requestWithoutAuth(
	ctx context.Context,
	method, path string,
	query map[string]string,
	headers map[string]string,
	body any,
	out any,
) error {
	return fn(ctx, method, path, query, headers, body, out)
}

func TestAuthManagerSharesConcurrentTokenRefresh(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	requester := tokenRequesterFunc(func(_ context.Context, _, _ string, _ map[string]string, _ map[string]string, _ any, out any) error {
		mu.Lock()
		calls++
		call := calls
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		token := out.(*TokenResponse)
		token.AccessToken = fmt.Sprintf("token-%d", call)
		token.ExpiresIn = 3600
		return nil
	})
	auth := newAuthManager("client", "secret", time.Minute, requester)

	const workers = 20
	results := make(chan string, workers)
	errors := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			token, err := auth.getToken(context.Background(), "")
			results <- token
			errors <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	for token := range results {
		if token != "token-1" {
			t.Fatalf("unexpected token: %s", token)
		}
	}
	if calls != 1 {
		t.Fatalf("expected one token request, got %d", calls)
	}

	const refreshWorkers = 10
	results = make(chan string, refreshWorkers)
	errors = make(chan error, refreshWorkers)
	for i := 0; i < refreshWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			token, err := auth.getToken(context.Background(), "token-1")
			results <- token
			errors <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	for token := range results {
		if token != "token-2" {
			t.Fatalf("unexpected refreshed token: %s", token)
		}
	}
	if calls != 2 {
		t.Fatalf("expected one shared refresh, got %d total token requests", calls)
	}
}

func TestTokenRefreshAtClampsBufferForShortTokens(t *testing.T) {
	now := time.Now()
	refreshAt := tokenRefreshAt(now, 10, time.Minute)
	remaining := refreshAt.Sub(now)
	if remaining != 8*time.Second {
		t.Fatalf("expected 80%% token lifetime, got %s", remaining)
	}
}

func TestAuthManagerRejectsInvalidTokenResponse(t *testing.T) {
	requester := tokenRequesterFunc(func(_ context.Context, _, _ string, _ map[string]string, _ map[string]string, _ any, out any) error {
		token := out.(*TokenResponse)
		token.AccessToken = "token"
		token.ExpiresIn = 0
		return nil
	})
	auth := newAuthManager("client", "secret", time.Minute, requester)
	if _, err := auth.getToken(context.Background(), ""); err == nil {
		t.Fatal("expected invalid token response error")
	}
}
