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


CommandStatus = Literal["SENT", "SUCCESS", "FAILED", "TIMEOUT"]
RelayAction = Literal["ON", "OFF", "JOG"]


class Command(TypedDict, total=False):
    id: str
    deviceId: str
    relayIndex: int
    action: RelayAction
    status: CommandStatus
    createdAt: str


class CommandResult(Command, total=False):
    pass
