# WLTE OpenAPI Go SDK

## Local Usage

Install the nested Go module directly from the repository:

```sh
go get github.com/svnwi/wlte-openapi-sdk/sdk/go@v0.3.0
```

## Integrate From Another Project

Clone this repository and use `sdk/go` as a local Go module during the current beta phase.

Typical flow:

```sh
git clone https://github.com/svnwi/wlte-openapi-sdk.git
cd wlte-openapi-sdk/sdk/go
go test ./...
```

In your own service, use a local `replace` directive that points to this `sdk/go` directory:

```go
replace github.com/svnwi/wlte-openapi-sdk/sdk/go => ../wlte-openapi-sdk/sdk/go
```

## Example Setup

Create `sdk/go/.env` from `.env.example`, then fill in:

```sh
cp .env.example .env
```

The Go examples load `.env` from `sdk/go/` when you run them from that directory.
The recommended module path is `github.com/svnwi/wlte-openapi-sdk/sdk/go`.

## Public API

```go
client, err := wlteopenapi.NewClient(wlteopenapi.ClientOptions{
    ClientID:     os.Getenv("WLTE_CLIENT_ID"),
    ClientSecret: os.Getenv("WLTE_CLIENT_SECRET"),
    BaseURL:      os.Getenv("WLTE_BASE_URL"),
})
if err != nil {
    log.Fatal(err)
}

devices, err := client.Devices.List(context.Background(), wlteopenapi.ListDevicesOptions{
    Page:     1,
    PageSize: 50,
})
```

## List Device Profiles

```go
profiles, err := client.Profiles.List(context.Background())
```

## Control Relays

```go
execution, err := client.Relays.Control(context.Background(), deviceID, wlteopenapi.RelayCommandOptions{
    Relays: []wlteopenapi.RelayCommand{
        {Index: 1, Action: wlteopenapi.RelayActionOn},
        {Index: 2, Action: wlteopenapi.RelayActionOff},
    },
    IdempotencyKey: "unique-key",
})
```

## Behavior

- Requests tokens automatically with client credentials.
- Caches tokens in memory only.
- Coalesces concurrent token requests and refreshes.
- Refreshes before expiry when possible.
- Retries once after `401 AUTH_EXPIRED`.
- Returns `APIError` for HTTP and business errors.
- Preserves the server `requestId` on `APIError`.
- Exposes `RetryAfter` for rate limit responses when provided.

## WebSocket

```go
session, err := client.WebSocket.Connect(ctx, wlteopenapi.WebSocketConnectOptions{})
if err != nil {
    log.Fatal(err)
}
defer session.Close()

device, err := session.GetDeviceState(ctx, deviceID)
if err != nil {
    log.Fatal(err)
}

for event := range session.Events() {
    log.Printf("event topic=%s data=%s", event.Topic, event.Data)
}
```

The SDK acquires the WebSocket ticket internally and does not expose it to
application code. Consume `Events()` continuously and monitor `DroppedEvents()`
if event delivery is business-critical.

## Run Examples

```sh
cd sdk/go
go run ./examples/list_devices
go run ./examples/list_profiles
go run ./examples/get_device
go run ./examples/control_relay
go run ./examples/websocket
```

## Versioning

This SDK is a nested module. Releases use tags such as `sdk/go/v0.3.0`, created
with `scripts/tag-go-sdk.sh v0.3.0` from the repository root.
