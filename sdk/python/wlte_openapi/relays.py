from __future__ import annotations

from urllib import parse

from .types import Command, RelayJogOptions, RelaySetOptions


class RelaysApi:
    def __init__(self, client) -> None:
        self._client = client

    def set(self, device_id: str, options: RelaySetOptions) -> Command:
        response = self._client.request(
            f"/wlte/v1/devices/{parse.quote(device_id, safe='')}/relays/{options['index']}/commands",
            method="POST",
            headers={"Idempotency-Key": options.get("idempotencyKey")},
            body={"action": "ON" if options["on"] else "OFF"},
        )
        return response["data"]

    def jog(self, device_id: str, options: RelayJogOptions) -> Command:
        response = self._client.request(
            f"/wlte/v1/devices/{parse.quote(device_id, safe='')}/relays/{options['index']}/commands",
            method="POST",
            headers={"Idempotency-Key": options.get("idempotencyKey")},
            body={"action": "JOG"},
        )
        return response["data"]
