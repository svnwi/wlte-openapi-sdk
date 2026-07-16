import type { ApiEnvelope, TokenResponse, WlteClientOptions } from './types.js'

export interface TokenTransport {
  requestWithoutAuth<T>(path: string, options: { method: 'POST'; body: unknown }): Promise<T>
}

export class AuthManager {
  private accessToken?: string
  private refreshAt = 0
  private readonly refreshBufferMs: number
  private refreshPromise?: Promise<string>

  constructor(
    private readonly options: Required<Pick<WlteClientOptions, 'clientId' | 'clientSecret'>> &
      Pick<WlteClientOptions, 'tokenRefreshBufferMs'>,
    private readonly transport: TokenTransport,
  ) {
    this.refreshBufferMs = options.tokenRefreshBufferMs ?? 60_000
    if (this.refreshBufferMs < 0) {
      throw new Error('tokenRefreshBufferMs must not be negative')
    }
  }

  async getToken(rejectedToken?: string): Promise<string> {
    if (this.accessToken && this.accessToken !== rejectedToken && Date.now() < this.refreshAt) {
      return this.accessToken
    }

    const refresh = this.refreshPromise ?? this.refreshToken()
    this.refreshPromise = refresh
    try {
      return await refresh
    } finally {
      if (this.refreshPromise === refresh) {
        this.refreshPromise = undefined
      }
    }
  }

  private async refreshToken(): Promise<string> {
    const envelope = await this.transport.requestWithoutAuth<ApiEnvelope<TokenResponse>>('/wlte/v1/auth/token', {
      method: 'POST',
      body: {
        clientId: this.options.clientId,
        clientSecret: this.options.clientSecret,
      },
    })

    if (!envelope.data.accessToken) {
      throw new Error('token response did not contain accessToken')
    }
    if (envelope.data.expiresIn <= 0) {
      throw new Error('token response expiresIn must be greater than zero')
    }
    this.accessToken = envelope.data.accessToken
    const ttlMs = envelope.data.expiresIn * 1000
    const effectiveBuffer = Math.min(this.refreshBufferMs, ttlMs / 5)
    this.refreshAt = Date.now() + ttlMs - effectiveBuffer
    return this.accessToken
  }
}
