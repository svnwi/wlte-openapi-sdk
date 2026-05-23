export interface WlteClientOptions {
  clientId: string
  clientSecret: string
  baseUrl?: string
  fetch?: typeof fetch
  tokenRefreshBufferMs?: number
}

export interface ApiEnvelope<T> {
  code: string
  message: string
  requestId?: string
  data: T
}

export interface TokenResponse {
  accessToken: string
  tokenType?: string
  expiresIn: number
}

export interface Pagination {
  page: number
  pageSize: number
  total: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

export interface DeviceStats {
  total: number
  online: number
  offline: number
}

export interface DeviceList {
  devices: Device[]
  stats: DeviceStats
  pagination: Pagination
}

export interface DeviceProfileList {
  profiles: DeviceProfile[]
}

export interface Device {
  deviceId: string
  name: string
  status: 'ONLINE' | 'OFFLINE'
  deviceType?: string
  stateUpdatedAt?: string
  peripherals?: Peripherals
  [key: string]: unknown
}

export interface Peripherals {
  relays?: RelayState[]
  digitalInputs?: DigitalInputState[]
  analogInputs?: AnalogInputState[]
  sensors?: SensorState[]
}

export interface DeviceProfile {
  deviceType: string
  capabilities: DeviceProfileCapabilities
}

export interface DeviceProfileCapabilities {
  relayCount?: number
  digitalInputCount?: number
  analogInputCount?: number
  sensorInterfaces?: SensorInterface[]
  operationSpecs?: OperationSpecs
}

export interface SensorInterface {
  index: number
  supportedTypes: string[]
}

export interface OperationSpecs {
  relay?: {
    actions: RelayAction[]
  }
}

export interface RelayState {
  index: number
  on: boolean | null
}

export interface DigitalInputState {
  index: number
  active: boolean | null
}

export interface AnalogInputState {
  index: number
  type: string
  value?: number
  unit?: string
  status: string
  measurement?: {
    value: number
    unit: string
  }
}

export interface SensorState {
  index: number
  type: string
  value: number
  unit: string
  status: string
}

export interface ListDevicesOptions {
  page?: number
  pageSize?: number
}

export interface RelaySetOptions {
  index: number
  on: boolean
  idempotencyKey?: string
}

export interface RelayJogOptions {
  index: number
  durationMs?: number
  idempotencyKey?: string
}

export type RelayAction = 'ON' | 'OFF' | 'JOG'
export type CommandStatus = 'SENT' | 'SUCCESS' | 'FAILED' | 'TIMEOUT'

export interface Command {
  id: string
  deviceId: string
  relayIndex: number
  action: RelayAction
  status?: CommandStatus
  createdAt?: string
  [key: string]: unknown
}

export interface CommandResult extends Command {
  status: CommandStatus
  createdAt: string
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  query?: Record<string, string | number | boolean | undefined>
  headers?: Record<string, string | undefined>
  body?: unknown
}
