export class WlteApiError extends Error {
  readonly status: number
  readonly code: string
  readonly data?: unknown
  readonly retryAfter?: string

  constructor(params: {
    status: number
    code: string
    message: string
    data?: unknown
    retryAfter?: string
  }) {
    super(params.message)
    this.name = 'WlteApiError'
    this.status = params.status
    this.code = params.code
    this.data = params.data
    this.retryAfter = params.retryAfter
  }
}

export function isAuthExpired(error: unknown): boolean {
  return error instanceof WlteApiError && error.status === 401 && error.code === 'AUTH_EXPIRED'
}

