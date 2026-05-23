# SDK Quickstart

## Step 1: Validate the API with Bruno

Use the Bruno collection in `examples/bruno/WLTE-OpenAPI`.

Before you start, open `examples/bruno/WLTE-OpenAPI/environments/quickstart.bru` and replace only:

- `clientId`
- `clientSecret`

The collection scripts will populate the other runtime variables automatically:

- `accessToken`
- `deviceId`
- `commandId`

Recommended request order:

1. `00-auth/Auth.bru`
2. `01-device-queries/01-list-devices.bru`
3. `01-device-queries/03-list-profiles.bru`
4. continue with real-time status, relay control, or command result requests

## Step 2: Run an SDK Example

### TypeScript

```sh
cd sdk/typescript
cp .env.example .env
npm install
npm run example:list-devices
npm run example:list-profiles
```

### Python

```sh
cp sdk/python/.env.example sdk/python/.env
python3 sdk/python/examples/list_devices.py
python3 sdk/python/examples/list_profiles.py
```

### Go

```sh
cd sdk/go
cp .env.example .env
go run ./examples/list_devices
go run ./examples/list_profiles
```

## Step 3: Integrate into Your Own Service

- TypeScript: build and reference `sdk/typescript` as a local package
- Python: add `sdk/python` to `PYTHONPATH` or vendor `wlte_openapi`
- Go: use `github.com/svnwi/wlte-openapi-sdk/sdk/go` with a temporary local `replace` directive if needed

## Support

For test access or integration support, contact `support@svnwi.com`.
