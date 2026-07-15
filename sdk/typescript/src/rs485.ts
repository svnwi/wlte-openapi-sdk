import type { WlteClient } from './client.js'
import type { ApiEnvelope, CommandExecution, RS485BaudRateOptions, RS485TransceiveOptions } from './types.js'

export class RS485Api {
  constructor(private readonly client: WlteClient) {}

  async transceive(deviceId: string, options: RS485TransceiveOptions): Promise<CommandExecution> {
    const response = await this.client.request<ApiEnvelope<CommandExecution>>(
      `/wlte/v1/devices/${encodeURIComponent(deviceId)}/rs485/transceive`,
      {
        method: 'POST',
        headers: {
          'Idempotency-Key': options.idempotencyKey,
        },
        body: {
          requestHex: options.requestHex,
        },
      },
    )
    return response.data
  }

  async setBaudRate(deviceId: string, options: RS485BaudRateOptions): Promise<CommandExecution> {
    const response = await this.client.request<ApiEnvelope<CommandExecution>>(
      `/wlte/v1/devices/${encodeURIComponent(deviceId)}/rs485/baud-rate`,
      {
        method: 'PUT',
        headers: {
          'Idempotency-Key': options.idempotencyKey,
        },
        body: {
          baudRate: options.baudRate,
        },
      },
    )
    return response.data
  }
}
