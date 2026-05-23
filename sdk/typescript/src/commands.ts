import type { WlteClient } from './client.js'
import type { ApiEnvelope, CommandResult } from './types.js'

export class CommandsApi {
  constructor(private readonly client: WlteClient) {}

  async getResult(commandId: string): Promise<CommandResult> {
    const response = await this.client.request<ApiEnvelope<CommandResult>>(
      `/wlte/v1/commands/${encodeURIComponent(commandId)}`,
    )
    return response.data
  }
}
