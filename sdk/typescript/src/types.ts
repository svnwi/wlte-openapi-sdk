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
  clientId: string
  scopes: string[]
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

export interface DeviceConfig {
  relay?: RelayConfig
  rs485?: RS485Config
  updatedAt?: string
  [key: string]: unknown
}

export interface RelayConfig {
  channels: RelayChannelConfig[]
}

export interface RelayChannelConfig {
  index: number
  jogTimeSeconds?: number
}

export interface RS485Config {
  baudRate?: number
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
  idempotencyKey: string
}

export interface RelayCommand {
  index: number
  action: RelayAction
}

export interface RelayCommandOptions {
  relays: RelayCommand[]
  idempotencyKey: string
}

export interface RelayJogOptions {
  index: number
  idempotencyKey: string
}

export interface RelayJogConfigOptions {
  index: number
  durationSec: number
  idempotencyKey: string
}

export interface RS485TransceiveOptions {
  requestHex: string
  idempotencyKey: string
}

export interface RS485BaudRateOptions {
  baudRate: number
  idempotencyKey: string
}

export type RelayAction = 'ON' | 'OFF' | 'JOG'
export type CommandStatus = 'SENT' | 'SUCCESS' | 'FAILED' | 'TIMEOUT'
export type CommandOperation =
  | 'device.relay.set'
  | 'device.rs485.transceive'
  | 'device.rs485.baudRate.set'
  | 'device.relay.jogConfig.set'

export interface RS485TransceiveResult {
  responseHex?: string
}

export interface RS485BaudRateResult {
  baudRate: number
}

export interface RelayJogConfigResult {
  relayIndex: number
  durationSec: number
}

export type CommandResultData = RS485TransceiveResult | RS485BaudRateResult | RelayJogConfigResult | Record<string, unknown>

export interface Command {
  id: string
  deviceId: string
  operation: CommandOperation
  status: CommandStatus
  params?: Record<string, unknown>
  result?: CommandResultData
  createdAt: string
  [key: string]: unknown
}

export type CommandResult = Command

export interface CommandDeviceState {
  deviceId: string
  status: 'ONLINE' | 'OFFLINE'
  peripherals?: Peripherals
  stateUpdatedAt?: string
}

export interface CommandExecution {
  command: Command
  state?: CommandDeviceState
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  query?: Record<string, string | number | boolean | undefined>
  headers?: Record<string, string | undefined>
  body?: unknown
}
