from __future__ import annotations

from urllib import parse

from .types import CommandExecution, RelayCommandOptions, RelayJogConfigOptions, RelayJogOptions, RelaySetOptions


class RelaysApi:
    def __init__(self, client) -> None:
        self._client = client

    def set(self, device_id: str, options: RelaySetOptions) -> CommandExecution:
        return self.control(
            device_id,
            {
                "relays": [{"index": options["index"], "action": "ON" if options["on"] else "OFF"}],
                "idempotencyKey": options.get("idempotencyKey", ""),
            },
        )

    def control(self, device_id: str, options: RelayCommandOptions) -> CommandExecution:
        response = self._client.request(
            f"/wlte/v1/devices/{parse.quote(device_id, safe='')}/relays/commands",
            method="POST",
            headers={"Idempotency-Key": options.get("idempotencyKey")},
            body={"relays": options["relays"]},
        )
        return response["data"]

    def jog(self, device_id: str, options: RelayJogOptions) -> CommandExecution:
        return self.control(
            device_id,
            {
                "relays": [{"index": options["index"], "action": "JOG"}],
                "idempotencyKey": options.get("idempotencyKey", ""),
            },
        )

    def set_jog_config(self, device_id: str, options: RelayJogConfigOptions) -> CommandExecution:
        response = self._client.request(
            f"/wlte/v1/devices/{parse.quote(device_id, safe='')}/relays/{options['index']}/jog-config",
            method="PUT",
            headers={"Idempotency-Key": options.get("idempotencyKey")},
            body={"durationSec": options["durationSec"]},
        )
        return response["data"]
