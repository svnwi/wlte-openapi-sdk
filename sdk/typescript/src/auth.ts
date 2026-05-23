import type { ApiEnvelope, TokenResponse, WlteClientOptions } from './types.js'

export interface TokenTransport {
  requestWithoutAuth<T>(path: string, options: { method: 'POST'; body: unknown }): Promise<T>
}

export class AuthManager {
  private accessToken?: string
  private refreshAt = 0
  private readonly refreshBufferMs: number

  constructor(
    private readonly options: Required<Pick<WlteClientOptions, 'clientId' | 'clientSecret'>> &
      Pick<WlteClientOptions, 'tokenRefreshBufferMs'>,
    private readonly transport: TokenTransport,
  ) {
    this.refreshBufferMs = options.tokenRefreshBufferMs ?? 60_000
  }

  async getToken(forceRefresh = false): Promise<string> {
    if (!forceRefresh && this.accessToken && Date.now() < this.refreshAt) {
      return this.accessToken
    }

    const envelope = await this.transport.requestWithoutAuth<ApiEnvelope<TokenResponse>>('/wlte/v1/auth/token', {
      method: 'POST',
      body: {
        clientId: this.options.clientId,
        clientSecret: this.options.clientSecret,
      },
    })

    this.accessToken = envelope.data.accessToken
    this.refreshAt = Date.now() + envelope.data.expiresIn * 1000 - this.refreshBufferMs
    return this.accessToken
  }
}
