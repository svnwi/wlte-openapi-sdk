import type { WlteClient } from './client.js'
import type {
  AddDeviceOptions,
  AddDeviceResult,
  ApiEnvelope,
  Device,
  DeviceConfig,
  DeviceList,
  ListDevicesOptions,
  ModifyDevicePasswordOptions,
  ModifyDevicePasswordResult,
  RemoveDeviceResult,
} from './types.js'

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

  async getConfig(deviceId: string): Promise<DeviceConfig> {
    const response = await this.client.request<ApiEnvelope<DeviceConfig>>(
      `/wlte/v1/devices/${encodeURIComponent(deviceId)}/config`,
    )
    return response.data
  }

  async add(options: AddDeviceOptions): Promise<AddDeviceResult> {
    const response = await this.client.request<ApiEnvelope<AddDeviceResult>>('/wlte/v1/devices', {
      method: 'POST',
      body: options,
    })
    return response.data
  }

  async remove(deviceId: string): Promise<RemoveDeviceResult> {
    const response = await this.client.request<ApiEnvelope<RemoveDeviceResult>>(
      `/wlte/v1/devices/${encodeURIComponent(deviceId)}`,
      { method: 'DELETE' },
    )
    return response.data
  }

  async modifyPassword(deviceId: string, options: ModifyDevicePasswordOptions): Promise<ModifyDevicePasswordResult> {
    const response = await this.client.request<ApiEnvelope<ModifyDevicePasswordResult>>(
      `/wlte/v1/devices/${encodeURIComponent(deviceId)}/password`,
      { method: 'PUT', body: options },
    )
    return response.data
  }
}
