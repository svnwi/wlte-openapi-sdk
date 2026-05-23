import type { WlteClient } from './client.js'
import type { ApiEnvelope, DeviceProfileList } from './types.js'

export class ProfilesApi {
  constructor(private readonly client: WlteClient) {}

  async list(): Promise<DeviceProfileList> {
    const response = await this.client.request<ApiEnvelope<DeviceProfileList>>('/wlte/v1/device-profiles')
    return response.data
  }
}

