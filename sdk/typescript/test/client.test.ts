import { describe, expect, it } from 'vitest'
import { WlteClient, WlteApiError } from '../src/index.js'

function jsonResponse(body: unknown, init: ResponseInit = {}): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'Content-Type': 'application/json', ...(init.headers ?? {}) },
    ...init,
  })
}

describe('WlteClient', () => {
  it('lists devices with automatic token request', async () => {
    const calls: string[] = []
    const fetchMock: typeof fetch = async (input) => {
      const url = String(input)
      calls.push(url)

      if (url.endsWith('/wlte/v1/auth/token')) {
        return jsonResponse({ code: 'SUCCESS', message: 'ok', data: { accessToken: 'token', expiresIn: 3600 } })
      }

      return jsonResponse({
        code: 'SUCCESS',
        message: 'ok',
        data: {
          devices: [{ deviceId: 'device-1', name: 'Device 1', status: 'ONLINE' }],
          stats: { total: 1, online: 1, offline: 0 },
          pagination: { page: 1, pageSize: 50, total: 1, totalPages: 1, hasNext: false, hasPrev: false },
        },
      })
    }

    const client = new WlteClient({
      clientId: 'client',
      clientSecret: 'secret',
      baseUrl: 'https://api.test',
      fetch: fetchMock,
    })

    const result = await client.devices.list({ page: 1, pageSize: 50 })
    expect(result.devices[0]?.deviceId).toBe('device-1')
    expect(calls).toHaveLength(2)
  })

  it('retries once on AUTH_EXPIRED', async () => {
    let deviceCalls = 0
    const fetchMock: typeof fetch = async (input) => {
      const url = String(input)

      if (url.endsWith('/wlte/v1/auth/token')) {
        return jsonResponse({ code: 'SUCCESS', message: 'ok', data: { accessToken: `token-${deviceCalls}`, expiresIn: 3600 } })
      }

      deviceCalls += 1
      if (deviceCalls === 1) {
        return jsonResponse({ code: 'AUTH_EXPIRED', message: 'expired', data: null }, { status: 401 })
      }

      return jsonResponse({ code: 'SUCCESS', message: 'ok', data: { deviceId: 'device-1', name: 'Device 1', status: 'ONLINE' } })
    }

    const client = new WlteClient({
      clientId: 'client',
      clientSecret: 'secret',
      baseUrl: 'https://api.test',
      fetch: fetchMock,
    })

    const device = await client.devices.get('device-1')
    expect(device.deviceId).toBe('device-1')
    expect(deviceCalls).toBe(2)
  })

  it('parses rate limit errors', async () => {
    const fetchMock: typeof fetch = async (input) => {
      const url = String(input)

      if (url.endsWith('/wlte/v1/auth/token')) {
        return jsonResponse({ code: 'SUCCESS', message: 'ok', data: { accessToken: 'token', expiresIn: 3600 } })
      }

      return jsonResponse(
        { code: 'RATE_LIMITED', message: 'too many requests', data: null },
        { status: 429, headers: { 'Retry-After': '5' } },
      )
    }

    const client = new WlteClient({
      clientId: 'client',
      clientSecret: 'secret',
      baseUrl: 'https://api.test',
      fetch: fetchMock,
    })

    await expect(client.devices.list()).rejects.toMatchObject<WlteApiError>({
      status: 429,
      code: 'RATE_LIMITED',
      retryAfter: '5',
    })
  })

  it('maps relay set requests to relay command action', async () => {
    const calls: Array<{ url: string; init?: RequestInit }> = []
    const fetchMock: typeof fetch = async (input, init) => {
      const url = String(input)
      calls.push({ url, init })

      if (url.endsWith('/wlte/v1/auth/token')) {
        return jsonResponse({ code: 'SUCCESS', message: 'ok', data: { accessToken: 'token', expiresIn: 3600 } })
      }

      return jsonResponse(
        {
          code: 'COMMAND_ACCEPTED',
          message: 'accepted',
          data: { id: 'cmd-1', deviceId: 'device-1', relayIndex: 1, action: 'ON' },
        },
        { status: 202 },
      )
    }

    const client = new WlteClient({
      clientId: 'client',
      clientSecret: 'secret',
      baseUrl: 'https://api.test',
      fetch: fetchMock,
    })

    const command = await client.relays.set('device-1', { index: 1, on: true, idempotencyKey: 'idem-1' })

    expect(command.action).toBe('ON')
    expect(calls[1]?.url).toBe('https://api.test/wlte/v1/devices/device-1/relays/1/commands')
    expect(calls[1]?.init?.body).toBe(JSON.stringify({ action: 'ON' }))
  })

  it('lists device profiles', async () => {
    const fetchMock: typeof fetch = async (input) => {
      const url = String(input)

      if (url.endsWith('/wlte/v1/auth/token')) {
        return jsonResponse({ code: 'SUCCESS', message: 'ok', data: { accessToken: 'token', expiresIn: 3600 } })
      }

      return jsonResponse({
        code: 'SUCCESS',
        message: 'ok',
        data: {
          profiles: [
            {
              deviceType: 'RL1',
              capabilities: {
                relayCount: 1,
                operationSpecs: { relay: { actions: ['ON', 'OFF', 'JOG'] } },
              },
            },
          ],
        },
      })
    }

    const client = new WlteClient({
      clientId: 'client',
      clientSecret: 'secret',
      baseUrl: 'https://api.test',
      fetch: fetchMock,
    })

    const result = await client.profiles.list()
    expect(result.profiles[0]?.deviceType).toBe('RL1')
  })
})
