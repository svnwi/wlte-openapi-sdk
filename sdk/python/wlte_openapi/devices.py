from __future__ import annotations

from urllib import parse

from .types import (
    AddDeviceOptions,
    AddDeviceResult,
    Device,
    DeviceConfig,
    DeviceList,
    ModifyDevicePasswordOptions,
    ModifyDevicePasswordResult,
    RemoveDeviceResult,
)


class DevicesApi:
    def __init__(self, client) -> None:
        self._client = client

    def list(self, *, page: int | None = None, page_size: int | None = None) -> DeviceList:
        response = self._client.request(
            "/wlte/v1/devices",
            query={
                "page": page,
                "pageSize": page_size,
            },
        )
        return response["data"]

    def get(self, device_id: str) -> Device:
        response = self._client.request(f"/wlte/v1/devices/{parse.quote(device_id, safe='')}")
        return response["data"]

    def get_config(self, device_id: str) -> DeviceConfig:
        response = self._client.request(f"/wlte/v1/devices/{parse.quote(device_id, safe='')}/config")
        return response["data"]

    def add(self, options: AddDeviceOptions) -> AddDeviceResult:
        response = self._client.request("/wlte/v1/devices", method="POST", body=options)
        return response["data"]

    def remove(self, device_id: str) -> RemoveDeviceResult:
        response = self._client.request(
            f"/wlte/v1/devices/{parse.quote(device_id, safe='')}",
            method="DELETE",
        )
        return response["data"]

    def modify_password(
        self,
        device_id: str,
        options: ModifyDevicePasswordOptions,
    ) -> ModifyDevicePasswordResult:
        response = self._client.request(
            f"/wlte/v1/devices/{parse.quote(device_id, safe='')}/password",
            method="PUT",
            body=options,
        )
        return response["data"]
