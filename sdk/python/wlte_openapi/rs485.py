from __future__ import annotations

from urllib import parse

from .types import Command, RS485BaudRateOptions, RS485TransceiveOptions


class RS485Api:
    def __init__(self, client) -> None:
        self._client = client

    def transceive(self, device_id: str, options: RS485TransceiveOptions) -> Command:
        response = self._client.request(
            f"/wlte/v1/devices/{parse.quote(device_id, safe='')}/rs485/transceive",
            method="POST",
            headers={"Idempotency-Key": options.get("idempotencyKey")},
            body={"requestHex": options["requestHex"]},
        )
        return response["data"]

    def set_baud_rate(self, device_id: str, options: RS485BaudRateOptions) -> Command:
        response = self._client.request(
            f"/wlte/v1/devices/{parse.quote(device_id, safe='')}/rs485/baud-rate",
            method="PUT",
            headers={"Idempotency-Key": options.get("idempotencyKey")},
            body={"baudRate": options["baudRate"]},
        )
        return response["data"]
