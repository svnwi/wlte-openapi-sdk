# Python SDK

## Local Integration

Use the Python SDK from this repository for now. Registry publishing is intentionally not enabled yet.

## Setup

```sh
cp sdk/python/.env.example sdk/python/.env
```

## Public API

```python
from wlte_openapi import WlteClient

client = WlteClient(
    client_id="xxx",
    client_secret="xxx",
    base_url="https://openapi.svnwi.com",
)

devices = client.devices.list(page=1, page_size=50)
profiles = client.profiles.list()
device = client.devices.get(device_id)
client.devices.add({"deviceId": device_id, "password": "1234", "name": "Demo"})
client.devices.modify_password(device_id, {"oldPassword": "1234", "newPassword": "5678"})
command = client.relays.set(device_id, {"index": 1, "on": True, "idempotencyKey": "unique-key"})
result = client.commands.get_result(command["id"])
```

## Behavior

- Requests tokens automatically with client credentials.
- Caches tokens in memory only.
- Coalesces concurrent token requests and refreshes.
- Refreshes before expiry when possible.
- Retries once after `401 AUTH_EXPIRED`.
- Raises `WlteApiError` for HTTP and business errors.
- Preserves the server `requestId` on `WlteApiError`.
- Exposes `retry_after` for rate limit responses when provided.

## Generation

Run from repository root:

```sh
scripts/generate-sdk.sh
```

Generated Python output can be produced locally with `scripts/generate-sdk.sh`, but it is not committed by default. The public SDK code lives under `sdk/python/wlte_openapi/`.

## Run Examples

```sh
python3 sdk/python/examples/list_devices.py
python3 sdk/python/examples/list_profiles.py
python3 sdk/python/examples/get_device.py
python3 sdk/python/examples/control_relay.py
```

## WebSocket Coverage

This package currently provides REST APIs only. Go services can use the Go SDK
when WebSocket requests and events are required.
