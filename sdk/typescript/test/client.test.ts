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
          data: {
            command: {
              id: 'cmd-1',
              deviceId: 'device-1',
              operation: 'device.relay.set',
              status: 'SUCCESS',
              params: { relays: [{ index: 1, action: 'ON' }] },
              createdAt: '2026-07-15T00:00:00Z',
            },
            state: { deviceId: 'device-1', status: 'ONLINE', peripherals: { relays: [{ index: 1, on: true }] } },
          },
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

    const execution = await client.relays.set('device-1', { index: 1, on: true, idempotencyKey: 'idem-1' })

    expect(execution.command.operation).toBe('device.relay.set')
    expect(execution.state?.peripherals?.relays?.[0]?.on).toBe(true)
    expect(calls[1]?.url).toBe('https://api.test/wlte/v1/devices/device-1/relays/commands')
    expect(calls[1]?.init?.body).toBe(JSON.stringify({ relays: [{ index: 1, action: 'ON' }] }))
  })

  it('controls multiple relays in one request', async () => {
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
          data: {
            command: {
              id: 'cmd-2',
              deviceId: 'device-1',
              operation: 'device.relay.set',
              status: 'SUCCESS',
              params: { relays: [{ index: 1, action: 'ON' }, { index: 2, action: 'OFF' }] },
              createdAt: '2026-07-15T00:00:00Z',
            },
          },
        },
        { status: 202 },
      )
    }
    const client = new WlteClient({ clientId: 'client', clientSecret: 'secret', baseUrl: 'https://api.test', fetch: fetchMock })

    const execution = await client.relays.control('device-1', {
      relays: [{ index: 1, action: 'ON' }, { index: 2, action: 'OFF' }],
      idempotencyKey: 'idem-multi',
    })

    expect(execution.command.operation).toBe('device.relay.set')
    expect(calls[1]?.init?.body).toBe(JSON.stringify({ relays: [{ index: 1, action: 'ON' }, { index: 2, action: 'OFF' }] }))
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


  it('gets device config', async () => {
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
        data: { relay: { channels: [{ index: 1, jogTimeSeconds: 2 }] }, rs485: { baudRate: 9600 } },
      })
    }

    const client = new WlteClient({ clientId: 'client', clientSecret: 'secret', baseUrl: 'https://api.test', fetch: fetchMock })

    const config = await client.devices.getConfig('device-1')

    expect(config.relay?.channels[0]?.jogTimeSeconds).toBe(2)
    expect(calls[1]).toBe('https://api.test/wlte/v1/devices/device-1/config')
  })

  it('sets relay jog config', async () => {
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
          data: { command: { id: 'cmd-1', deviceId: 'device-1', operation: 'device.relay.jogConfig.set', status: 'SUCCESS', params: { relayIndex: 1, durationSec: 2 }, result: { relayIndex: 1, durationSec: 2 }, createdAt: '2026-07-15T00:00:00Z' } },
        },
        { status: 202 },
      )
    }

    const client = new WlteClient({ clientId: 'client', clientSecret: 'secret', baseUrl: 'https://api.test', fetch: fetchMock })

    const execution = await client.relays.setJogConfig('device-1', { index: 1, durationSec: 2, idempotencyKey: 'idem-jog' })

    expect(execution.command.operation).toBe('device.relay.jogConfig.set')
    expect(calls[1]?.url).toBe('https://api.test/wlte/v1/devices/device-1/relays/1/jog-config')
    expect(calls[1]?.init?.method).toBe('PUT')
    expect(calls[1]?.init?.body).toBe(JSON.stringify({ durationSec: 2 }))
  })

  it('sends rs485 transceive requests', async () => {
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
          data: { command: { id: 'cmd-485', deviceId: 'device-1', operation: 'device.rs485.transceive', status: 'SUCCESS', params: { requestHex: '020600340000C837' }, result: { responseHex: '020600340000C837' }, createdAt: '2026-07-15T00:00:00Z' } },
        },
        { status: 202 },
      )
    }

    const client = new WlteClient({ clientId: 'client', clientSecret: 'secret', baseUrl: 'https://api.test', fetch: fetchMock })

    const execution = await client.rs485.transceive('device-1', { requestHex: '020600340000C837', idempotencyKey: 'idem-485' })

    expect(execution.command.operation).toBe('device.rs485.transceive')
    expect(calls[1]?.url).toBe('https://api.test/wlte/v1/devices/device-1/rs485/transceive')
    expect(calls[1]?.init?.method).toBe('POST')
    expect(calls[1]?.init?.body).toBe(JSON.stringify({ requestHex: '020600340000C837' }))
  })

  it('sets rs485 baud rate', async () => {
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
          data: { command: { id: 'cmd-baud', deviceId: 'device-1', operation: 'device.rs485.baudRate.set', status: 'SUCCESS', params: { baudRate: 9600 }, result: { baudRate: 9600 }, createdAt: '2026-07-15T00:00:00Z' } },
        },
        { status: 202 },
      )
    }

    const client = new WlteClient({ clientId: 'client', clientSecret: 'secret', baseUrl: 'https://api.test', fetch: fetchMock })

    const execution = await client.rs485.setBaudRate('device-1', { baudRate: 9600, idempotencyKey: 'idem-baud' })

    expect(execution.command.operation).toBe('device.rs485.baudRate.set')
    expect(calls[1]?.url).toBe('https://api.test/wlte/v1/devices/device-1/rs485/baud-rate')
    expect(calls[1]?.init?.method).toBe('PUT')
    expect(calls[1]?.init?.body).toBe(JSON.stringify({ baudRate: 9600 }))
  })

})
