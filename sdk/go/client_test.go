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
				return jsonResponse(429, `{"code":"RATE_LIMITED","message":"too many requests","requestId":"req-rate","data":null}`, map[string]string{"Retry-After": "5"}), nil
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
	if apiErr.Status != 429 || apiErr.Code != "RATE_LIMITED" || apiErr.RetryAfter != "5" || apiErr.RequestID != "req-rate" {
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
				return jsonResponse(202, `{"code":"COMMAND_ACCEPTED","message":"accepted","data":{"command":{"id":"cmd-1","deviceId":"device-1","operation":"device.relay.set","status":"SUCCESS","params":{"relays":[{"index":1,"action":"ON"}]},"createdAt":"2026-07-15T00:00:00Z"},"state":{"deviceId":"device-1","status":"ONLINE","peripherals":{"relays":[{"index":1,"on":true}]}}}}`, nil), nil
			}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	execution, err := client.Relays.Set(context.Background(), "device-1", RelaySetOptions{
		Index:          1,
		On:             true,
		IdempotencyKey: "idem-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	if execution.Command.Operation != CommandOperationRelaySet {
		t.Fatalf("unexpected operation: %s", execution.Command.Operation)
	}
	if execution.State == nil || execution.State.Peripherals == nil || len(execution.State.Peripherals.Relays) != 1 {
		t.Fatalf("unexpected state: %+v", execution.State)
	}
	if commandURL != "https://api.test/wlte/v1/devices/device-1/relays/commands" {
		t.Fatalf("unexpected url: %s", commandURL)
	}
	if commandBody != `{"relays":[{"index":1,"action":"ON"}]}` {
		t.Fatalf("unexpected body: %s", commandBody)
	}
	if idemKey != "idem-1" {
		t.Fatalf("unexpected idempotency key: %s", idemKey)
	}
}

func TestRelayControlSupportsMultipleRelays(t *testing.T) {
	var commandBody string
	client, err := NewClient(ClientOptions{
		ClientID: "client", ClientSecret: "secret", BaseURL: "https://api.test",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/wlte/v1/auth/token" {
				return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"accessToken":"token","expiresIn":3600,"tokenType":"Bearer"}}`, nil), nil
			}
			payload, _ := io.ReadAll(req.Body)
			commandBody = string(payload)
			return jsonResponse(202, `{"code":"COMMAND_ACCEPTED","message":"accepted","data":{"command":{"id":"cmd-2","deviceId":"device-1","operation":"device.relay.set","status":"SUCCESS","params":{"relays":[{"index":1,"action":"ON"},{"index":2,"action":"OFF"}]},"createdAt":"2026-07-15T00:00:00Z"}}}`, nil), nil
		})},
	})
	if err != nil {
		t.Fatal(err)
	}

	execution, err := client.Relays.Control(context.Background(), "device-1", RelayCommandOptions{
		Relays:         []RelayCommand{{Index: 1, Action: RelayActionOn}, {Index: 2, Action: RelayActionOff}},
		IdempotencyKey: "idem-multi",
	})
	if err != nil {
		t.Fatal(err)
	}
	if execution.Command.Operation != CommandOperationRelaySet {
		t.Fatalf("unexpected operation: %s", execution.Command.Operation)
	}
	if commandBody != `{"relays":[{"index":1,"action":"ON"},{"index":2,"action":"OFF"}]}` {
		t.Fatalf("unexpected body: %s", commandBody)
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
				return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"profiles":[{"deviceType":"RL1","capabilities":{"relayCount":1,"supportedOperations":["device.relay.set"],"operationSpecs":{"relay":{"actions":["ON","OFF","JOG"]}}}}]}}`, nil), nil
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
	if len(result.Profiles[0].Capabilities.SupportedOperations) != 1 || result.Profiles[0].Capabilities.SupportedOperations[0] != "device.relay.set" {
		t.Fatalf("unexpected supported operations: %+v", result.Profiles[0].Capabilities.SupportedOperations)
	}
}

func TestGetDeviceConfig(t *testing.T) {
	var configURL string
	client, err := NewClient(ClientOptions{
		ClientID:     "client",
		ClientSecret: "secret",
		BaseURL:      "https://api.test",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/wlte/v1/auth/token" {
					return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"accessToken":"token","expiresIn":3600,"tokenType":"Bearer"}}`, nil), nil
				}
				configURL = req.URL.String()
				return jsonResponse(200, `{"code":"SUCCESS","message":"ok","data":{"relay":{"channels":[{"index":1,"jogTimeSeconds":2}]},"rs485":{"baudRate":9600}}}`, nil), nil
			}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	config, err := client.Devices.GetConfig(context.Background(), "device-1")
	if err != nil {
		t.Fatal(err)
	}
	if config.Relay == nil || config.Relay.Channels[0].JogTimeSeconds != 2 {
		t.Fatalf("unexpected config: %+v", config)
	}
	if configURL != "https://api.test/wlte/v1/devices/device-1/config" {
		t.Fatalf("unexpected url: %s", configURL)
	}
}

func TestRelaySetJogConfig(t *testing.T) {
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
				return jsonResponse(202, `{"code":"COMMAND_ACCEPTED","message":"accepted","data":{"command":{"id":"cmd-1","deviceId":"device-1","operation":"device.relay.jogConfig.set","status":"SUCCESS","params":{"relayIndex":1,"durationSec":2},"result":{"relayIndex":1,"durationSec":2},"createdAt":"2026-07-15T00:00:00Z"}}}`, nil), nil
			}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	execution, err := client.Relays.SetJogConfig(context.Background(), "device-1", RelayJogConfigOptions{Index: 1, DurationSec: 2, IdempotencyKey: "idem-jog"})
	if err != nil {
		t.Fatal(err)
	}
	if execution.Command.Operation != CommandOperationRelayJogConfigSet {
		t.Fatalf("unexpected operation: %s", execution.Command.Operation)
	}
	if commandURL != "https://api.test/wlte/v1/devices/device-1/relays/1/jog-config" {
		t.Fatalf("unexpected url: %s", commandURL)
	}
	if commandBody != `{"durationSec":2}` {
		t.Fatalf("unexpected body: %s", commandBody)
	}
	if idemKey != "idem-jog" {
		t.Fatalf("unexpected idempotency key: %s", idemKey)
	}
}

func TestRS485Transceive(t *testing.T) {
	var commandURL string
	var commandBody string
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
				payload, _ := io.ReadAll(req.Body)
				commandBody = string(payload)
				return jsonResponse(202, `{"code":"COMMAND_ACCEPTED","message":"accepted","data":{"command":{"id":"cmd-485","deviceId":"device-1","operation":"device.rs485.transceive","status":"SUCCESS","params":{"requestHex":"020600340000C837"},"result":{"responseHex":"020600340000C837"},"createdAt":"2026-07-15T00:00:00Z"}}}`, nil), nil
			}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	execution, err := client.RS485.Transceive(context.Background(), "device-1", RS485TransceiveOptions{RequestHex: "020600340000C837", IdempotencyKey: "idem-485"})
	if err != nil {
		t.Fatal(err)
	}
	if execution.Command.Operation != CommandOperationRS485Transceive {
		t.Fatalf("unexpected operation: %s", execution.Command.Operation)
	}
	if commandURL != "https://api.test/wlte/v1/devices/device-1/rs485/transceive" {
		t.Fatalf("unexpected url: %s", commandURL)
	}
	if commandBody != `{"requestHex":"020600340000C837"}` {
		t.Fatalf("unexpected body: %s", commandBody)
	}
}

func TestRS485SetBaudRate(t *testing.T) {
	var commandURL string
	var commandBody string
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
				payload, _ := io.ReadAll(req.Body)
				commandBody = string(payload)
				return jsonResponse(202, `{"code":"COMMAND_ACCEPTED","message":"accepted","data":{"command":{"id":"cmd-baud","deviceId":"device-1","operation":"device.rs485.baudRate.set","status":"SUCCESS","params":{"baudRate":9600},"result":{"baudRate":9600},"createdAt":"2026-07-15T00:00:00Z"}}}`, nil), nil
			}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	execution, err := client.RS485.SetBaudRate(context.Background(), "device-1", RS485BaudRateOptions{BaudRate: 9600, IdempotencyKey: "idem-baud"})
	if err != nil {
		t.Fatal(err)
	}
	if execution.Command.Operation != CommandOperationRS485BaudRateSet {
		t.Fatalf("unexpected operation: %s", execution.Command.Operation)
	}
	if commandURL != "https://api.test/wlte/v1/devices/device-1/rs485/baud-rate" {
		t.Fatalf("unexpected url: %s", commandURL)
	}
	if commandBody != `{"baudRate":9600}` {
		t.Fatalf("unexpected body: %s", commandBody)
	}
}
