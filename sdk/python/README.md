# WLTE OpenAPI Python SDK

## Local Usage

Use this SDK from the repository for now. Public package publishing is intentionally not enabled yet.

## Integrate From Another Project

Clone this repository and add `sdk/python` to your runtime import path during the current beta phase.

Typical flow:

```sh
git clone https://github.com/svnwi/wlte-openapi-sdk.git
cd wlte-openapi-sdk
```

Then either:

- vendor `sdk/python/wlte_openapi` into your service repository
- add `sdk/python` to `PYTHONPATH` in your deployment environment

## Example Setup

The Python examples are usable now.

Create `sdk/python/.env` from `.env.example`, then fill in:

```sh
cp .env.example .env
```

The example scripts automatically load `.env` if it exists, so you do not need to export each variable manually.
They can be run directly from the repository root without setting `PYTHONPATH`.

## Authentication

```python
from wlte_openapi import WlteClient

client = WlteClient(
    client_id="xxx",
    client_secret="xxx",
    base_url="https://openapi.svnwi.com",
)
```

The SDK requests access tokens automatically, caches them in memory, refreshes before expiry when possible, and retries once after `401 AUTH_EXPIRED`.

## List Devices

```python
result = client.devices.list(page=1, page_size=50)
print(result["devices"])
```

## List Device Profiles

```python
profiles = client.profiles.list()
print(profiles["profiles"])
```

## Get Single Device

```python
device = client.devices.get(device_id)
```

## Control Relay

```python
client.relays.set(device_id, {"index": 1, "on": True, "idempotencyKey": "unique-key"})
client.relays.jog(device_id, {"index": 1, "idempotencyKey": "unique-key"})
```

## Query Command Result

```python
result = client.commands.get_result(command_id)
```

## Error Handling

```python
from wlte_openapi import WlteApiError

try:
    client.devices.list()
except WlteApiError as error:
    print(error.status, error.code, error.message, error.data)
```

## Rate Limit Handling

HTTP `429` and `RATE_LIMITED` responses are exposed as `WlteApiError`. If the server returns a `Retry-After` header, it is available on `error.retry_after`.

## Version Compatibility

This SDK tracks `openapi/openapi.yaml` in this repository and is currently intended for direct repository-based integration.

## WebSocket Status

WebSocket support is not included in the current SDK version.
It will be added after the WLTE OpenAPI WebSocket protocol is finalized.

## Run Examples

```sh
python3 sdk/python/examples/list_devices.py
python3 sdk/python/examples/list_profiles.py
python3 sdk/python/examples/get_device.py
python3 sdk/python/examples/control_relay.py
```
