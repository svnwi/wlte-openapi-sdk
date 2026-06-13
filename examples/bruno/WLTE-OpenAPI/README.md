# WLTE OpenAPI Bruno Quickstart

## Why Bruno

Bruno is a file-based API client. It is a good fit for SDK onboarding because:

- Requests can be versioned directly in Git.
- Environment variables and test scripts stay next to the request definitions.
- New users can run the real API flow without writing code first.
- Response assertions are built in, so users can quickly confirm whether the API is behaving as expected.

## What Is Included

This collection covers the core WLTE OpenAPI HTTP flow:

- `00-auth`: get access token
- `01-device-queries`: list devices, query real-time status, read device config, list device profiles
- `02-relay-control`: turn relay on, turn relay off, jog relay, set relay jog time
- `03-command-result`: query command execution result
- `04-rs485`: send RS485 transceive commands and set RS485 baud rate

## Before You Start

The shared [quickstart.bru](environments/quickstart.bru) environment is safe to commit. It contains placeholders only.

For local testing in Bruno Desktop, create a local environment file and fill in your real credentials:

```bash
cp environments/quickstart.bru environments/quickstart.local.bru
```

Then edit `environments/quickstart.local.bru` and replace:

- `clientId`
- `clientSecret`
- Optional: `deviceId`
- Optional: `rs485DeviceId`

`quickstart.local.bru` is ignored by Git and should not be committed.

You do not need to manually replace the other variables before getting started.

The collection scripts will populate the main runtime variables for you:

- `accessToken` is written after `00-auth/Auth.bru`
- `deviceId` is written after `01-device-queries/01-list-devices.bru`
- `commandId` is written after relay control, relay jog config, and RS485 requests
- `rs485DeviceId` defaults to the first listed device unless you set it manually

`relayIndex` is preset to `1`. Change it only if you want to target a different relay channel.

For RS485 examples, set `rs485DeviceId` to a device that supports RS485. The default fallback is only for quick collection navigation; RS485 commands require a compatible device.

## Quick Usage

1. Import or open the `WLTE-OpenAPI` collection in Bruno.
2. Select the `quickstart.local` environment.
3. Run `00-auth/Auth.bru`.
4. Run `01-device-queries/01-list-devices.bru`.
5. If you want to test RS485, set `rs485DeviceId` to an RS485-capable device.
6. Continue with the request you need:
   relay control, relay jog time, real-time status, device config, device profiles, RS485, or command result.

Recommended first path:

1. `00-auth/Auth.bru`
2. `01-device-queries/01-list-devices.bru`
3. `01-device-queries/02-get-device-real-time-status.bru`
4. `01-device-queries/04-get-device-config.bru`
5. `02-relay-control/04-set-relay-jog-time.bru`
6. `04-rs485/01-rs485-transceive.bru`
7. `03-command-result/01-get-command-result.bru`

## Notes

- Each request includes assertions, so a successful response is not just HTTP success but also payload-shape validation.
- The relay control, relay jog config, and RS485 command requests automatically generate an `Idempotency-Key`.
- If you want to test a specific device, you can manually override `deviceId` after running the list request once.

## Test Access

We can provide a testable API key for evaluation.

You do not need to purchase hardware first to experience the API or test against different device models. Contact `support@svnwi.com` to request test access and available demo device coverage.
