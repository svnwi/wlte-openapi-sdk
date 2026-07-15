# WLTE OpenAPI Go SDK

## Local Usage

Use this SDK from the repository for now. Public package publishing is intentionally not enabled yet.

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
- Refreshes before expiry when possible.
- Retries once after `401 AUTH_EXPIRED`.
- Returns `APIError` for HTTP and business errors.
- Exposes `RetryAfter` for rate limit responses when provided.

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
