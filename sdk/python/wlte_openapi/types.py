from __future__ import annotations

from typing import Literal, TypedDict


class Pagination(TypedDict):
    page: int
    pageSize: int
    total: int
    totalPages: int
    hasNext: bool
    hasPrev: bool


class Device(TypedDict, total=False):
    deviceId: str
    name: str
    status: Literal["ONLINE", "OFFLINE"]
    deviceType: str
    stateUpdatedAt: str
    peripherals: "Peripherals"


class DeviceStats(TypedDict):
    total: int
    online: int
    offline: int


class RelayState(TypedDict, total=False):
    index: int
    on: bool | None


class DigitalInputState(TypedDict, total=False):
    index: int
    active: bool | None


class AnalogMeasurement(TypedDict):
    value: float
    unit: str


class AnalogInputState(TypedDict, total=False):
    index: int
    type: str
    unit: str
    value: float
    status: str
    measurement: AnalogMeasurement


class SensorState(TypedDict):
    index: int
    type: str
    value: float
    unit: str
    status: str


class Peripherals(TypedDict, total=False):
    relays: list[RelayState]
    digitalInputs: list[DigitalInputState]
    analogInputs: list[AnalogInputState]
    sensors: list[SensorState]


class DeviceList(TypedDict):
    devices: list[Device]
    stats: DeviceStats
    pagination: Pagination


class RelayChannelConfig(TypedDict, total=False):
    index: int
    jogTimeSeconds: int


class RelayConfig(TypedDict):
    channels: list[RelayChannelConfig]


class RS485Config(TypedDict, total=False):
    baudRate: int


class DeviceConfig(TypedDict, total=False):
    relay: RelayConfig
    rs485: RS485Config
    updatedAt: str


class SensorInterface(TypedDict):
    index: int
    supportedTypes: list[str]


class RelayOperationSpec(TypedDict):
    actions: list[Literal["ON", "OFF", "JOG"]]


class OperationSpecs(TypedDict, total=False):
    relay: RelayOperationSpec


class DeviceProfileCapabilities(TypedDict, total=False):
    relayCount: int
    digitalInputCount: int
    analogInputCount: int
    sensorInterfaces: list[SensorInterface]
    operationSpecs: OperationSpecs


class DeviceProfile(TypedDict):
    deviceType: str
    capabilities: DeviceProfileCapabilities


class DeviceProfileList(TypedDict):
    profiles: list[DeviceProfile]


class RelaySetOptions(TypedDict, total=False):
    index: int
    on: bool
    idempotencyKey: str


class RelayJogOptions(TypedDict, total=False):
    index: int
    durationMs: int
    idempotencyKey: str


class RelayJogConfigOptions(TypedDict, total=False):
    index: int
    durationSec: int
    idempotencyKey: str


class RS485TransceiveOptions(TypedDict, total=False):
    requestHex: str
    idempotencyKey: str


class RS485BaudRateOptions(TypedDict, total=False):
    baudRate: int
    idempotencyKey: str


CommandStatus = Literal["SENT", "SUCCESS", "FAILED", "TIMEOUT"]
RelayAction = Literal["ON", "OFF", "JOG"]
CommandType = Literal["RELAY_SET", "RS485_TRANSCEIVE", "RS485_BAUD_RATE_SET", "RELAY_JOG_CONFIG_SET"]


class RS485TransceiveResult(TypedDict, total=False):
    requestHex: str
    responseHex: str


class RS485BaudRateResult(TypedDict):
    baudRate: int


class RelayJogConfigResult(TypedDict):
    relayIndex: int
    durationSec: int


class Command(TypedDict, total=False):
    id: str
    deviceId: str
    type: CommandType
    relayIndex: int
    action: RelayAction
    status: CommandStatus
    result: RS485TransceiveResult | RS485BaudRateResult | RelayJogConfigResult | dict[str, object]
    createdAt: str


class CommandResult(Command):
    status: CommandStatus
    createdAt: str
