import type { WlteClient } from './client.js'
import type { ApiEnvelope, Device, DeviceList, ListDevicesOptions } from './types.js'

export class DevicesApi {
  constructor(private readonly client: WlteClient) {}

  async list(options: ListDevicesOptions = {}): Promise<DeviceList> {
    const response = await this.client.request<ApiEnvelope<DeviceList>>('/wlte/v1/devices', {
      query: {
        page: options.page,
        pageSize: options.pageSize,
      },
    })
    return response.data
  }

  async get(deviceId: string): Promise<Device> {
    const response = await this.client.request<ApiEnvelope<Device>>(`/wlte/v1/devices/${encodeURIComponent(deviceId)}`)
    return response.data
  }
}
