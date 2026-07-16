# TypeScript SDK

## Local Integration

Use the TypeScript SDK from this repository for now. Registry publishing is intentionally not enabled yet.

## Setup

```sh
cd sdk/typescript
cp .env.example .env
npm install
npm run build
```

## Public API

```ts
const client = new WlteClient({
  clientId: process.env.WLTE_CLIENT_ID!,
  clientSecret: process.env.WLTE_CLIENT_SECRET!,
  baseUrl: process.env.WLTE_BASE_URL,
})

const devices = await client.devices.list({ page: 1, pageSize: 50 })
const profiles = await client.profiles.list()
const device = await client.devices.get(deviceId)
await client.devices.add({ deviceId, password: '1234', name: 'Demo' })
await client.devices.modifyPassword(deviceId, { oldPassword: '1234', newPassword: '5678' })
const command = await client.relays.set(deviceId, { index: 1, on: true, idempotencyKey: 'unique-key' })
const result = await client.commands.getResult(command.id)
```

The import path depends on how you wire `sdk/typescript` into your own workspace during the current beta phase.

## Behavior

- Requests tokens automatically with client credentials.
- Caches tokens in memory only.
- Coalesces concurrent token requests and refreshes.
- Refreshes before expiry when possible.
- Retries once after `401 AUTH_EXPIRED`.
- Raises `WlteApiError` for HTTP and business errors.
- Preserves the server `requestId` on `WlteApiError`.
- Exposes `retryAfter` for rate limit responses when provided.

## Generation

Run from repository root:

```sh
scripts/generate-sdk.sh
```

Generated TypeScript output can be produced locally with `scripts/generate-sdk.sh`, but it is not committed by default. The public SDK code lives under `sdk/typescript/src/`.

## Run Examples

```sh
npm run example:list-devices
npm run example:list-profiles
npm run example:get-device
npm run example:control-relay
```

## WebSocket Coverage

This package currently provides REST APIs only. Go services can use the Go SDK
when WebSocket requests and events are required.
