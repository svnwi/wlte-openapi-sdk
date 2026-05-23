package wlteopenapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func jsonResponse(status int, body string, headers map[string]string) *http.Response {
	responseHeaders := make(http.Header)
	for key, value := range headers {
		responseHeaders.Set(key, value)
	}

	return &http.Response{
		StatusCode: status,
		Header:     responseHeaders,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestListDevicesRequestsTokenAutomatically(t *testing.T) {
	var calls []string
	client, err := NewClient(ClientOptions{
		ClientID:     "client",
		ClientSecret: "secret",
		BaseURL:      "https://api.test",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls = append(calls, req.URL.String())
				if req.URL.Path == "/wlte/v1/auth/token" {
					return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"accessToken":"token","expiresIn":3600,"tokenType":"Bearer"}}`, nil), nil
				}
				return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"devices":[{"deviceId":"device-1","name":"Device 1","status":"ONLINE"}],"stats":{"total":1,"online":1,"offline":0},"pagination":{"page":1,"pageSize":50,"total":1,"totalPages":1,"hasNext":false,"hasPrev":false}}}`, nil), nil
			}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Devices.List(context.Background(), ListDevicesOptions{Page: 1, PageSize: 50})
	if err != nil {
		t.Fatal(err)
	}

	if result.Devices[0].DeviceID != "device-1" {
		t.Fatalf("unexpected device id: %s", result.Devices[0].DeviceID)
	}
	if len(calls) != 2 {
		t.Fatalf("unexpected call count: %d", len(calls))
	}
}

func TestRetriesOnceOnAuthExpired(t *testing.T) {
	deviceCalls := 0
	client, err := NewClient(ClientOptions{
		ClientID:     "client",
		ClientSecret: "secret",
		BaseURL:      "https://api.test",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/wlte/v1/auth/token" {
					return jsonResponse(200, fmt.Sprintf(`{"code":"SUCCESS","message":"ok","data":{"accessToken":"token-%d","expiresIn":3600,"tokenType":"Bearer"}}`, deviceCalls), nil), nil
				}

				deviceCalls++
				if deviceCalls == 1 {
					return jsonResponse(401, `{"code":"AUTH_EXPIRED","message":"expired","data":null}`, nil), nil
				}
				return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"deviceId":"device-1","name":"Device 1","status":"ONLINE"}}`, nil), nil
			}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	device, err := client.Devices.Get(context.Background(), "device-1")
	if err != nil {
		t.Fatal(err)
	}
	if device.DeviceID != "device-1" {
		t.Fatalf("unexpected device id: %s", device.DeviceID)
	}
	if deviceCalls != 2 {
		t.Fatalf("unexpected retry count: %d", deviceCalls)
	}
}

func TestParsesRateLimitErrors(t *testing.T) {
	client, err := NewClient(ClientOptions{
		ClientID:     "client",
		ClientSecret: "secret",
		BaseURL:      "https://api.test",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/wlte/v1/auth/token" {
					return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"accessToken":"token","expiresIn":3600,"tokenType":"Bearer"}}`, nil), nil
				}
				return jsonResponse(429, `{"code":"RATE_LIMITED","message":"too many requests","data":null}`, map[string]string{"Retry-After": "5"}), nil
			}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Devices.List(context.Background(), ListDevicesOptions{})
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("unexpected error type: %T", err)
	}
	if apiErr.Status != 429 || apiErr.Code != "RATE_LIMITED" || apiErr.RetryAfter != "5" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestRelaySetMapsToCommandAction(t *testing.T) {
	var commandURL string
	var commandBody string
	var idemKey string
	client, err := NewClient(ClientOptions{
		ClientID:     "client",
		ClientSecret: "secret",
		BaseURL:      "https://api.test",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/wlte/v1/auth/token" {
					return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"accessToken":"token","expiresIn":3600,"tokenType":"Bearer"}}`, nil), nil
				}
				commandURL = req.URL.String()
				idemKey = req.Header.Get("Idempotency-Key")
				payload, _ := io.ReadAll(req.Body)
				commandBody = string(payload)
				return jsonResponse(202, `{"code":"COMMAND_ACCEPTED","message":"accepted","data":{"id":"cmd-1","deviceId":"device-1","relayIndex":1,"action":"ON"}}`, nil), nil
			}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	command, err := client.Relays.Set(context.Background(), "device-1", RelaySetOptions{
		Index:          1,
		On:             true,
		IdempotencyKey: "idem-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	if command.Action != RelayActionOn {
		t.Fatalf("unexpected action: %s", command.Action)
	}
	if commandURL != "https://api.test/wlte/v1/devices/device-1/relays/1/commands" {
		t.Fatalf("unexpected url: %s", commandURL)
	}
	if commandBody != `{"action":"ON"}` {
		t.Fatalf("unexpected body: %s", commandBody)
	}
	if idemKey != "idem-1" {
		t.Fatalf("unexpected idempotency key: %s", idemKey)
	}
}

func TestListProfiles(t *testing.T) {
	client, err := NewClient(ClientOptions{
		ClientID:     "client",
		ClientSecret: "secret",
		BaseURL:      "https://api.test",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/wlte/v1/auth/token" {
					return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"accessToken":"token","expiresIn":3600,"tokenType":"Bearer"}}`, nil), nil
				}
				return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"profiles":[{"deviceType":"RL1","capabilities":{"relayCount":1,"operationSpecs":{"relay":{"actions":["ON","OFF","JOG"]}}}}]}}`, nil), nil
			}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Profiles.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Profiles[0].DeviceType != "RL1" {
		t.Fatalf("unexpected profile type: %s", result.Profiles[0].DeviceType)
	}
}
