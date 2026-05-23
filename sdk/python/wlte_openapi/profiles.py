from __future__ import annotations

from .types import DeviceProfileList


class ProfilesApi:
    def __init__(self, client) -> None:
        self._client = client

    def list(self) -> DeviceProfileList:
        response = self._client.request("/wlte/v1/device-profiles")
        return response["data"]

