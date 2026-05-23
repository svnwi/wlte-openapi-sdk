from __future__ import annotations

from urllib import parse

from .types import Device, DeviceList


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
