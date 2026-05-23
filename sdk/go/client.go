package wlteopenapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultBaseURL = "https://openapi.svnwi.com"

var successCodes = map[string]struct{}{
	"SUCCESS":          {},
	"COMMAND_ACCEPTED": {},
	"OK":               {},
}

type ClientOptions struct {
	ClientID           string
	ClientSecret       string
	BaseURL            string
	HTTPClient         *http.Client
	TokenRefreshBuffer time.Duration
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	auth       *authManager

	Devices  *DevicesService
	Profiles *ProfilesService
	Relays   *RelaysService
	Commands *CommandsService
}

func NewClient(options ClientOptions) (*Client, error) {
	if strings.TrimSpace(options.ClientID) == "" {
		return nil, fmt.Errorf("client id is required")
	}
	if strings.TrimSpace(options.ClientSecret) == "" {
		return nil, fmt.Errorf("client secret is required")
	}

	baseURL := strings.TrimRight(options.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	refreshBuffer := options.TokenRefreshBuffer
	if refreshBuffer == 0 {
		refreshBuffer = time.Minute
	}

	client := &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
	client.auth = newAuthManager(options.ClientID, options.ClientSecret, refreshBuffer, client)
	client.Devices = &DevicesService{client: client}
	client.Profiles = &ProfilesService{client: client}
	client.Relays = &RelaysService{client: client}
	client.Commands = &CommandsService{client: client}
	return client, nil
}

func (c *Client) request(ctx context.Context, method, path string, query map[string]string, headers map[string]string, body any, out any) error {
	token, err := c.auth.getToken(ctx, false)
	if err != nil {
		return err
	}

	err = c.send(ctx, method, path, query, headers, body, token, out)
	if err == nil {
		return nil
	}
	if !isAuthExpired(err) {
		return err
	}

	token, err = c.auth.getToken(ctx, true)
	if err != nil {
		return err
	}
	return c.send(ctx, method, path, query, headers, body, token, out)
}

func (c *Client) requestWithoutAuth(ctx context.Context, method, path string, query map[string]string, headers map[string]string, body any, out any) error {
	return c.send(ctx, method, path, query, headers, body, "", out)
}

func (c *Client) send(ctx context.Context, method, path string, query map[string]string, headers map[string]string, body any, token string, out any) error {
	requestURL, err := c.buildURL(path, query)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for key, value := range headers {
		if value != "" {
			req.Header.Set(key, value)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return decodeResponse(resp, responseBody, out)
}

func (c *Client) buildURL(path string, query map[string]string) (string, error) {
	parsed, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", err
	}

	values := parsed.Query()
	for key, value := range query {
		if value != "" {
			values.Set(key, value)
		}
	}
	parsed.RawQuery = values.Encode()
	return parsed.String(), nil
}

type rawEnvelope struct {
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	RequestID string          `json:"requestId"`
	Data      json.RawMessage `json:"data"`
}

func decodeResponse(resp *http.Response, responseBody []byte, out any) error {
	if len(bytes.TrimSpace(responseBody)) == 0 {
		return nil
	}

	var envelope rawEnvelope
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return &APIError{
				Status:     resp.StatusCode,
				Code:       "HTTP_ERROR",
				Message:    strings.TrimSpace(string(responseBody)),
				RetryAfter: resp.Header.Get("Retry-After"),
			}
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(responseBody, out)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			Status:     resp.StatusCode,
			Code:       envelope.Code,
			Message:    envelope.Message,
			Data:       envelope.Data,
			RetryAfter: resp.Header.Get("Retry-After"),
		}
	}

	if _, ok := successCodes[envelope.Code]; !ok {
		return &APIError{
			Status:     resp.StatusCode,
			Code:       envelope.Code,
			Message:    envelope.Message,
			Data:       envelope.Data,
			RetryAfter: resp.Header.Get("Retry-After"),
		}
	}

	if out == nil || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil
	}

	return json.Unmarshal(envelope.Data, out)
}
