# SDK Overview

WLTE OpenAPI currently provides a Bruno collection plus three repository-based SDKs:

- TypeScript
- Python
- Go

## Choose the Right Entry Point

- Use Bruno when you want to validate credentials, inspect payloads, and confirm the API flow without writing code.
- Use an SDK when you are integrating WLTE OpenAPI into an application or service.

Recommended first path for new users:

1. Run the Bruno quickstart.
2. Confirm device listing and device profile listing succeed.
3. Move to the SDK that matches your runtime.

## Current Coverage

The current handwritten SDK layer covers:

- automatic client-credentials token flow
- `devices.list()`
- `devices.get(deviceId)`
- `profiles.list()`
- `relays.set(...)`
- `relays.jog(...)`
- `commands.getResult(...)` or `commands.get_result(...)`

## Authentication Behavior

All three SDKs currently behave the same way:

- request access tokens automatically
- cache tokens in memory only
- refresh before expiry when possible
- retry once after `401 AUTH_EXPIRED`
- expose `429 RATE_LIMITED` details through the SDK error type

## Recommended Repository Path

For GitHub delivery, use:

- `github.com/svnwi/wlte-openapi-sdk`

That naming keeps the SDK repository separate from backend service repositories and gives the Go SDK a stable import path:

- `github.com/svnwi/wlte-openapi-sdk/sdk/go`

## Support

For test access or integration support, contact `support@svnwi.com`.
