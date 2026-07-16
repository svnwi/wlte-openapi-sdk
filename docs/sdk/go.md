# Go SDK

The Go SDK is available in this repository for direct integration.

## Local Integration

The Go SDK is a nested module under `sdk/go`.

Recommended repository path:

- `github.com/svnwi/wlte-openapi-sdk`
- Go module: `github.com/svnwi/wlte-openapi-sdk/sdk/go`

## Setup

```sh
go get github.com/svnwi/wlte-openapi-sdk/sdk/go@v0.3.0
```

For local repository development, use a `replace` directive that points to
`sdk/go` and run `go test ./...` from that directory.

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
profiles, err := client.Profiles.List(context.Background())
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

The SDK obtains and consumes the short-lived WebSocket ticket internally. A
session supports concurrent requests, protocol/application pings, typed device
state and operation helpers, and asynchronous events. Events use a bounded
buffer; monitor `DroppedEvents()` and consume `Events()` continuously.

## Generation

Run from repository root:

```sh
scripts/generate-sdk.sh
```

Generated Go output can be produced locally with `scripts/generate-sdk.sh`, but it is not committed by default. The public SDK code lives under `sdk/go/`.

## Run Examples

```sh
cd sdk/go
go run ./examples/list_devices
go run ./examples/list_profiles
go run ./examples/get_device
go run ./examples/control_relay
go run ./examples/websocket
```

## Release Tags

Because the Go SDK is a nested module, its Git tag must include the module
directory. Run `scripts/tag-go-sdk.sh v0.3.0`; it creates
`sdk/go/v0.3.0`. A root tag such as `v0.3.0` does not version this module.
