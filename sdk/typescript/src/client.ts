import { AuthManager } from './auth.js'
import { isAuthExpired, WlteApiError } from './errors.js'
import { CommandsApi } from './commands.js'
import { DevicesApi } from './devices.js'
import { ProfilesApi } from './profiles.js'
import { RelaysApi } from './relays.js'
import { RS485Api } from './rs485.js'
import type { ApiEnvelope, RequestOptions, WlteClientOptions } from './types.js'

const DEFAULT_BASE_URL = 'https://openapi.svnwi.com'
const SUCCESS_CODES = new Set(['SUCCESS', 'COMMAND_ACCEPTED', 'OK'])

export class WlteClient {
  readonly devices: DevicesApi
  readonly profiles: ProfilesApi
  readonly relays: RelaysApi
  readonly commands: CommandsApi
  readonly rs485: RS485Api

  private readonly baseUrl: string
  private readonly fetchImpl: typeof fetch
  private readonly auth: AuthManager

  constructor(options: WlteClientOptions) {
    if (!options.clientId) {
      throw new Error('clientId is required')
    }
    if (!options.clientSecret) {
      throw new Error('clientSecret is required')
    }

    this.baseUrl = (options.baseUrl ?? DEFAULT_BASE_URL).replace(/\/+$/, '')
    this.fetchImpl = options.fetch ?? fetch
    this.auth = new AuthManager(
      {
        clientId: options.clientId,
        clientSecret: options.clientSecret,
        tokenRefreshBufferMs: options.tokenRefreshBufferMs,
      },
      this,
    )
    this.devices = new DevicesApi(this)
    this.profiles = new ProfilesApi(this)
    this.relays = new RelaysApi(this)
    this.commands = new CommandsApi(this)
    this.rs485 = new RS485Api(this)
  }

  async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    return this.requestWithRetry<T>(path, options, true)
  }

  async requestWithoutAuth<T>(path: string, options: RequestOptions): Promise<T> {
    return this.send<T>(path, options)
  }

  private async requestWithRetry<T>(path: string, options: RequestOptions, retryOnAuthExpired: boolean): Promise<T> {
    const token = await this.auth.getToken()

    try {
      return await this.send<T>(path, options, token)
    } catch (error) {
      if (retryOnAuthExpired && isAuthExpired(error)) {
        const refreshedToken = await this.auth.getToken(token)
        return this.send<T>(path, options, refreshedToken)
      }
      throw error
    }
  }

  private async send<T>(path: string, options: RequestOptions, token?: string): Promise<T> {
    const response = await this.fetchImpl(this.buildUrl(path, options.query), {
      method: options.method ?? 'GET',
      headers: {
        Accept: 'application/json',
        ...(options.body === undefined ? {} : { 'Content-Type': 'application/json' }),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...definedHeaders(options.headers),
      },
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
    })

    const payload = await parseJson(response)

    if (!response.ok) {
      throw toApiError(response, payload)
    }

    if (isErrorEnvelope(payload)) {
      throw toApiError(response, payload)
    }

    return payload as T
  }

  private buildUrl(path: string, query?: RequestOptions['query']): string {
    const url = new URL(`${this.baseUrl}${path}`)

    for (const [key, value] of Object.entries(query ?? {})) {
      if (value !== undefined) {
        url.searchParams.set(key, String(value))
      }
    }

    return url.toString()
  }
}

async function parseJson(response: Response): Promise<unknown> {
  const text = await response.text()
  if (!text) {
    return undefined
  }

  try {
    return JSON.parse(text)
  } catch {
    return text
  }
}

function isErrorEnvelope(payload: unknown): payload is ApiEnvelope<unknown> {
  if (!payload || typeof payload !== 'object') {
    return false
  }

  const candidate = payload as { code?: unknown; message?: unknown }
  return typeof candidate.code === 'string' && !SUCCESS_CODES.has(candidate.code) && typeof candidate.message === 'string'
}

function toApiError(response: Response, payload: unknown): WlteApiError {
  const retryAfter = response.headers.get('Retry-After') ?? undefined

  if (payload && typeof payload === 'object') {
    const envelope = payload as { code?: unknown; message?: unknown; data?: unknown; requestId?: unknown }
    return new WlteApiError({
      status: response.status,
      code: typeof envelope.code === 'string' ? envelope.code : response.status === 429 ? 'RATE_LIMITED' : 'HTTP_ERROR',
      message: typeof envelope.message === 'string' ? envelope.message : response.statusText,
      data: envelope.data,
      retryAfter,
      requestId: typeof envelope.requestId === 'string' ? envelope.requestId : undefined,
    })
  }

  return new WlteApiError({
    status: response.status,
    code: response.status === 429 ? 'RATE_LIMITED' : 'HTTP_ERROR',
    message: typeof payload === 'string' && payload ? payload : response.statusText,
    retryAfter,
  })
}

function definedHeaders(headers?: Record<string, string | undefined>): Record<string, string> {
  const result: Record<string, string> = {}

  for (const [key, value] of Object.entries(headers ?? {})) {
    if (value !== undefined) {
      result[key] = value
    }
  }

  return result
}
