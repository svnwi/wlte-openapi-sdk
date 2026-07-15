import type { WlteClient } from './client.js'
import type {
  ApiEnvelope,
  CommandExecution,
  RelayCommandOptions,
  RelayJogConfigOptions,
  RelayJogOptions,
  RelaySetOptions,
} from './types.js'

export class RelaysApi {
  constructor(private readonly client: WlteClient) {}

  async set(deviceId: string, options: RelaySetOptions): Promise<CommandExecution> {
    return this.control(deviceId, {
      relays: [{ index: options.index, action: options.on ? 'ON' : 'OFF' }],
      idempotencyKey: options.idempotencyKey,
    })
  }

  async control(deviceId: string, options: RelayCommandOptions): Promise<CommandExecution> {
    const response = await this.client.request<ApiEnvelope<CommandExecution>>(
      `/wlte/v1/devices/${encodeURIComponent(deviceId)}/relays/commands`,
      {
        method: 'POST',
        headers: {
          'Idempotency-Key': options.idempotencyKey,
        },
        body: {
          relays: options.relays,
        },
      },
    )
    return response.data
  }

  async jog(deviceId: string, options: RelayJogOptions): Promise<CommandExecution> {
    return this.control(deviceId, {
      relays: [{ index: options.index, action: 'JOG' }],
      idempotencyKey: options.idempotencyKey,
    })
  }

  async setJogConfig(deviceId: string, options: RelayJogConfigOptions): Promise<CommandExecution> {
    const response = await this.client.request<ApiEnvelope<CommandExecution>>(
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
