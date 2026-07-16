# WLTE OpenAPI TypeScript SDK

## Local Usage

Use this SDK from the repository for now. Public package publishing is intentionally not enabled yet.

## Integrate From Another Project

Clone this repository and reference `sdk/typescript` as a local package during the current beta phase.

Typical flow:

```sh
git clone https://github.com/svnwi/wlte-openapi-sdk.git
cd wlte-openapi-sdk/sdk/typescript
npm install
npm run build
```

Then consume the built files from your own Node.js service by referencing this package directory directly in your workspace or with a local file dependency.

## Example Setup

The TypeScript examples need a local `.env` file and a TypeScript runtime in the dev environment.

Create `sdk/typescript/.env` from `.env.example`, then fill in:

```sh
cp .env.example .env
npm install
```

The package includes ready-to-run example scripts that use Node's `--env-file` support.

## Authentication

```ts
import { WlteClient } from './dist/index.js'

const client = new WlteClient({
  clientId: process.env.WLTE_CLIENT_ID!,
  clientSecret: process.env.WLTE_CLIENT_SECRET!,
  baseUrl: process.env.WLTE_BASE_URL,
})
```

The SDK requests access tokens automatically, caches them in memory, coalesces
concurrent refreshes, refreshes before expiry when possible, and retries once
after `401 AUTH_EXPIRED`.

## List Devices

```ts
const result = await client.devices.list({ page: 1, pageSize: 50 })
console.log(result.devices)
```

## List Device Profiles

```ts
const profiles = await client.profiles.list()
console.log(profiles.profiles)
```

## Get Single Device

```ts
const device = await client.devices.get(deviceId)
```

## Manage Devices

```ts
await client.devices.add({ deviceId, password: '1234', name: 'Demo' })
await client.devices.modifyPassword(deviceId, { oldPassword: '1234', newPassword: '5678' })
await client.devices.remove(deviceId)
```

## Control Relay

```ts
const execution = await client.relays.set(deviceId, { index: 1, on: true, idempotencyKey: 'unique-key' })
console.log(execution.command, execution.state)

await client.relays.control(deviceId, {
  relays: [
    { index: 1, action: 'ON' },
    { index: 2, action: 'OFF' },
  ],
  idempotencyKey: 'unique-multi-key',
})
```

## Query Command Result

```ts
const result = await client.commands.getResult(commandId)
```

## Error Handling

```ts
import { WlteApiError } from './dist/index.js'

try {
  await client.devices.list()
} catch (error) {
  if (error instanceof WlteApiError) {
    console.error(error.status, error.code, error.message, error.requestId, error.data)
  }
}
```

## Rate Limit Handling

HTTP `429` and `RATE_LIMITED` responses are exposed as `WlteApiError`. If the server returns a `Retry-After` header, it is available on `error.retryAfter`.

## Version Compatibility

This SDK tracks `openapi/openapi.yaml` in this repository and is currently intended for direct repository-based integration.

## WebSocket Coverage

This package currently provides REST APIs only. Use the Go SDK when WebSocket
requests and events are required.

## Run Examples

```sh
npm run example:list-devices
npm run example:list-profiles
npm run example:get-device
npm run example:control-relay
```
