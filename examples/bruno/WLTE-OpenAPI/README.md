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
- `01-device-queries`: list devices, query real-time status, list device profiles
- `02-relay-control`: turn relay on, turn relay off, jog relay
- `03-command-result`: query command execution result

## Before You Start

Open [quickstart.bru](environments/quickstart.bru) and replace only:

- `clientId`
- `clientSecret`

You do not need to manually replace the other variables before getting started.

The collection scripts will populate the main runtime variables for you:

- `accessToken` is written after `00-auth/Auth.bru`
- `deviceId` is written after `01-device-queries/01-list-devices.bru`
- `commandId` is written after relay control requests

`relayIndex` is preset to `1`. Change it only if you want to target a different relay channel.

## Quick Usage

1. Import or open the `WLTE-OpenAPI` collection in Bruno.
2. Select the `quickstart` environment.
3. Update `clientId` and `clientSecret`.
4. Run `00-auth/Auth.bru`.
5. Run `01-device-queries/01-list-devices.bru`.
6. Continue with the request you need:
   relay control, real-time status, device profiles, or command result.

Recommended first path:

1. `00-auth/Auth.bru`
2. `01-device-queries/01-list-devices.bru`
3. `01-device-queries/02-get-device-real-time-status.bru`
4. `02-relay-control/01-turn-relay-on.bru`
5. `03-command-result/01-get-command-result.bru`

## Notes

- Each request includes assertions, so a successful response is not just HTTP success but also payload-shape validation.
- The relay control requests automatically generate an `Idempotency-Key`.
- If you want to test a specific device, you can manually override `deviceId` after running the list request once.

## Test Access

We can provide a testable API key for evaluation.

You do not need to purchase hardware first to experience the API or test against different device models. Contact `support@svnwi.com` to request test access and available demo device coverage.
