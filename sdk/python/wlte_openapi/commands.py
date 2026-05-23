from __future__ import annotations

from urllib import parse

from .types import CommandResult


class CommandsApi:
    def __init__(self, client) -> None:
        self._client = client

    def get_result(self, command_id: str) -> CommandResult:
        response = self._client.request(f"/wlte/v1/commands/{parse.quote(command_id, safe='')}")
        return response["data"]
