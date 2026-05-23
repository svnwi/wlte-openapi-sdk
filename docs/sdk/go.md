# Go SDK

The Go SDK is available in this repository for direct integration.

## Local Integration

Use the Go SDK from this repository for now. Registry publishing is intentionally not enabled yet.

Recommended repository path:

- `github.com/svnwi/wlte-openapi-sdk`
- Go module: `github.com/svnwi/wlte-openapi-sdk/sdk/go`

## Setup

```sh
cd sdk/go
cp .env.example .env
go test ./...
```

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
- Refreshes before expiry when possible.
- Retries once after `401 AUTH_EXPIRED`.
- Returns `APIError` for HTTP and business errors.
- Exposes `RetryAfter` for rate limit responses when provided.

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
```

## WebSocket Status

WebSocket support is not included in the current SDK version.
It will be added after the WLTE OpenAPI WebSocket protocol is finalized.
