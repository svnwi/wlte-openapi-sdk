package wlteopenapi

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type tokenRequester interface {
	requestWithoutAuth(ctx context.Context, method, path string, query map[string]string, headers map[string]string, body any, out any) error
}

type authManager struct {
	mu            sync.Mutex
	clientID      string
	clientSecret  string
	refreshBuffer time.Duration
	requester     tokenRequester
	accessToken   string
	refreshAt     time.Time
	clientIDInfo  string
	scopes        []string
	refresh       *tokenRefresh
}

type tokenRefresh struct {
	done chan struct{}
	err  error
}

func newAuthManager(clientID, clientSecret string, refreshBuffer time.Duration, requester tokenRequester) *authManager {
	return &authManager{
		clientID:      clientID,
		clientSecret:  clientSecret,
		refreshBuffer: refreshBuffer,
		requester:     requester,
	}
}

// getToken returns the cached token unless rejectedToken is the token rejected
// by the server. Concurrent cache misses share one token request.
func (a *authManager) getToken(ctx context.Context, rejectedToken string) (string, error) {
	a.mu.Lock()
	if a.accessToken != "" && a.accessToken != rejectedToken && time.Now().Before(a.refreshAt) {
		token := a.accessToken
		a.mu.Unlock()
		return token, nil
	}
	if a.refresh != nil {
		refresh := a.refresh
		a.mu.Unlock()
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-refresh.done:
			if refresh.err != nil {
				return "", refresh.err
			}
			a.mu.Lock()
			token := a.accessToken
			a.mu.Unlock()
			return token, nil
		}
	}
	refresh := &tokenRefresh{done: make(chan struct{})}
	a.refresh = refresh
	a.mu.Unlock()

	var token TokenResponse
	err := a.requester.requestWithoutAuth(
		ctx,
		"POST",
		"/wlte/v1/auth/token",
		nil,
		nil,
		map[string]string{
			"clientId":     a.clientID,
			"clientSecret": a.clientSecret,
		},
		&token,
	)
	if err == nil && token.AccessToken == "" {
		err = fmt.Errorf("token response did not contain accessToken")
	}
	if err == nil && token.ExpiresIn <= 0 {
		err = fmt.Errorf("token response expiresIn must be greater than zero")
	}

	now := time.Now()
	a.mu.Lock()
	if err == nil {
		a.accessToken = token.AccessToken
		a.refreshAt = tokenRefreshAt(now, token.ExpiresIn, a.refreshBuffer)
		a.clientIDInfo = token.ClientID
		if a.clientIDInfo == "" {
			a.clientIDInfo = a.clientID
		}
		a.scopes = append(a.scopes[:0], token.Scopes...)
	}
	refresh.err = err
	a.refresh = nil
	close(refresh.done)
	a.mu.Unlock()
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (a *authManager) authorization(ctx context.Context) (AuthorizationInfo, error) {
	if _, err := a.getToken(ctx, ""); err != nil {
		return AuthorizationInfo{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return AuthorizationInfo{
		ClientID: a.clientIDInfo,
		Scopes:   append([]string(nil), a.scopes...),
	}, nil
}

func tokenRefreshAt(now time.Time, expiresIn int, configuredBuffer time.Duration) time.Time {
	ttl := time.Duration(expiresIn) * time.Second
	buffer := configuredBuffer
	maxBuffer := ttl / 5
	if buffer > maxBuffer {
		buffer = maxBuffer
	}
	return now.Add(ttl - buffer)
}
