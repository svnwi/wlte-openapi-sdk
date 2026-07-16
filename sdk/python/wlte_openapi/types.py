from __future__ import annotations

from typing import Literal, TypedDict


class TokenResponse(TypedDict):
    accessToken: str
    tokenType: Literal["Bearer"]
    expiresIn: int
    clientId: str
    scopes: list[str]


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
    deviceId: str
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
    supportedOperations: list[str]
    operationSpecs: OperationSpecs


class _AddDeviceRequired(TypedDict):
    deviceId: str
    password: str


class AddDeviceOptions(_AddDeviceRequired, total=False):
    name: str


class AddDeviceResult(TypedDict, total=False):
    deviceId: str
    name: str


class RemoveDeviceResult(TypedDict):
    deviceId: str


class ModifyDevicePasswordOptions(TypedDict):
    oldPassword: str
    newPassword: str


class ModifyDevicePasswordResult(TypedDict):
    deviceId: str
    updated: bool


class DeviceProfile(TypedDict):
    deviceType: str
    capabilities: DeviceProfileCapabilities


class DeviceProfileList(TypedDict):
    profiles: list[DeviceProfile]


class RelaySetOptions(TypedDict):
    index: int
    on: bool
    idempotencyKey: str


class RelayCommand(TypedDict):
    index: int
    action: "RelayAction"


class RelayCommandOptions(TypedDict):
    relays: list[RelayCommand]
    idempotencyKey: str


class RelayJogOptions(TypedDict):
    index: int
    idempotencyKey: str


class RelayJogConfigOptions(TypedDict):
    index: int
    durationSec: int
    idempotencyKey: str


class RS485TransceiveOptions(TypedDict):
    requestHex: str
    idempotencyKey: str


class RS485BaudRateOptions(TypedDict):
    baudRate: int
    idempotencyKey: str


CommandStatus = Literal["SENT", "SUCCESS", "FAILED", "TIMEOUT"]
RelayAction = Literal["ON", "OFF", "JOG"]
CommandOperation = Literal[
    "device.relay.set",
    "device.rs485.transceive",
    "device.rs485.baudRate.set",
    "device.relay.jogConfig.set",
]


class RS485TransceiveResult(TypedDict, total=False):
    responseHex: str


class RS485BaudRateResult(TypedDict):
    baudRate: int


class RelayJogConfigResult(TypedDict):
    relayIndex: int
    durationSec: int


class Command(TypedDict, total=False):
    id: str
    deviceId: str
    operation: CommandOperation
    status: CommandStatus
    params: dict[str, object]
    result: RS485TransceiveResult | RS485BaudRateResult | RelayJogConfigResult | dict[str, object]
    createdAt: str


class CommandResult(Command):
    status: CommandStatus
    createdAt: str


class CommandDeviceState(TypedDict, total=False):
    deviceId: str
    status: Literal["ONLINE", "OFFLINE"]
    peripherals: Peripherals
    stateUpdatedAt: str


class CommandExecution(TypedDict, total=False):
    command: Command
    state: CommandDeviceState
