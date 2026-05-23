# WLTE OpenAPI SDK

Developer SDKs and Bruno examples for integrating with the WLTE OpenAPI HTTP platform.

This repository is the public integration entry point for WLTE OpenAPI:

- TypeScript SDK
- Python SDK
- Go SDK
- Bruno quickstart collection

Current scope:

- HTTP APIs only
- automatic client-credentials authentication
- device listing and device status queries
- device profile discovery
- relay control
- command result polling

WebSocket support is intentionally excluded until the protocol is finalized.

## Why This Repository

- Use Bruno to validate credentials and inspect real responses before writing code.
- Use the SDKs to integrate WLTE OpenAPI into your own Node.js, Python, or Go services.
- Use `openapi/openapi.yaml` as the contract source for local client generation and SDK alignment.

## Quick Start

1. Start with Bruno in `examples/bruno/WLTE-OpenAPI`.
2. Set `clientId` and `clientSecret` in `environments/quickstart.bru`.
3. Run auth, list devices, and list profiles.
4. Move to the SDK that matches your runtime.

Documentation entry points:

- SDK overview: `docs/sdk/overview.md`
- SDK quickstart: `docs/sdk/quickstart.md`
- TypeScript SDK: `docs/sdk/typescript.md`
- Python SDK: `docs/sdk/python.md`
- Go SDK: `docs/sdk/go.md`

## Repository Path

Recommended public repository path:

- GitHub organization: `svnwi`
- Repository name: `wlte-openapi-sdk`
- Repository URL: `https://github.com/svnwi/wlte-openapi-sdk`
- Go module path: `github.com/svnwi/wlte-openapi-sdk/sdk/go`

This keeps the SDK repository separate from backend service repositories while preserving a stable Go import path.

## Repository Layout

```text
openapi/
  openapi.yaml
sdk/
  typescript/
  python/
  go/
docs/
  sdk/
examples/
  bruno/
scripts/
  generate-sdk.sh
```

## SDK Coverage

Current handwritten SDK coverage:

- `devices.list()`
- `devices.get(deviceId)`
- `profiles.list()`
- `relays.set(deviceId, { index, on })`
- `relays.jog(deviceId, { index })`
- `commands.getResult(commandId)` or `commands.get_result(command_id)`

Shared SDK behavior:

- automatic client-credentials token flow
- in-memory token caching
- refresh-before-expiry where possible
- single retry on `401 AUTH_EXPIRED`
- unified API error handling
- `429 RATE_LIMITED` surfaced through SDK error types

## Development

Generate low-level clients locally from `openapi/openapi.yaml` when you need to inspect or refresh generator output:

```sh
scripts/generate-sdk.sh
```

OpenAPI Generator output is not committed to this repository by default. Handwritten SDK code is the primary public surface.

Current integration mode:

- Use the SDKs directly from this repository.
- Use Bruno for quick API validation.
- Do not rely on public `npm`, `PyPI`, or other registry distribution yet.

## Bruno

Bruno examples are available in `examples/bruno`.

Use the `quickstart` environment, set your `clientId` and `clientSecret`, then run the requests in order:

1. `00-auth/Auth.bru`
2. `01-device-queries/01-list-devices.bru`
3. `01-device-queries/03-list-profiles.bru`
4. Relay and command examples as needed.

We can provide testable API credentials for evaluation, and you do not need to purchase hardware before trying the API or testing different device models.

## License

This repository is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

## Contact

For test access or integration support, contact `support@svnwi.com`.
