package wlteopenapi

import (
	"context"
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
}

func newAuthManager(clientID, clientSecret string, refreshBuffer time.Duration, requester tokenRequester) *authManager {
	return &authManager{
		clientID:      clientID,
		clientSecret:  clientSecret,
		refreshBuffer: refreshBuffer,
		requester:     requester,
	}
}

func (a *authManager) getToken(ctx context.Context, forceRefresh bool) (string, error) {
	a.mu.Lock()
	if !forceRefresh && a.accessToken != "" && time.Now().Before(a.refreshAt) {
		token := a.accessToken
		a.mu.Unlock()
		return token, nil
	}
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
	if err != nil {
		return "", err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.accessToken = token.AccessToken
	a.refreshAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Add(-a.refreshBuffer)
	return a.accessToken, nil
}
