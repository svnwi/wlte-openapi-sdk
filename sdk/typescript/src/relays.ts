import type { WlteClient } from './client.js'
import type { ApiEnvelope, Command, RelayJogConfigOptions, RelayJogOptions, RelaySetOptions } from './types.js'

export class RelaysApi {
  constructor(private readonly client: WlteClient) {}

  async set(deviceId: string, options: RelaySetOptions): Promise<Command> {
    const response = await this.client.request<ApiEnvelope<Command>>(
      `/wlte/v1/devices/${encodeURIComponent(deviceId)}/relays/${options.index}/commands`,
      {
        method: 'POST',
        headers: {
          'Idempotency-Key': options.idempotencyKey,
        },
        body: {
          action: options.on ? 'ON' : 'OFF',
        },
      },
    )
    return response.data
  }

  async jog(deviceId: string, options: RelayJogOptions): Promise<Command> {
    const response = await this.client.request<ApiEnvelope<Command>>(
      `/wlte/v1/devices/${encodeURIComponent(deviceId)}/relays/${options.index}/commands`,
      {
        method: 'POST',
        headers: {
          'Idempotency-Key': options.idempotencyKey,
        },
        body: {
          action: 'JOG',
        },
      },
    )
    return response.data
  }

  async setJogConfig(deviceId: string, options: RelayJogConfigOptions): Promise<Command> {
    const response = await this.client.request<ApiEnvelope<Command>>(
      `/wlte/v1/devices/${encodeURIComponent(deviceId)}/relays/${options.index}/jog-config`,
      {
        method: 'PUT',
        headers: {
          'Idempotency-Key': options.idempotencyKey,
        },
        body: {
          durationSec: options.durationSec,
        },
      },
    )
    return response.data
  }
}
