from __future__ import annotations

import time
from typing import Protocol


class TokenTransport(Protocol):
    def request_without_auth(self, path: str, *, method: str, body: object | None = None) -> dict:
        ...


class AuthManager:
    def __init__(
        self,
        *,
        client_id: str,
        client_secret: str,
        transport: TokenTransport,
        token_refresh_buffer_seconds: int = 60,
    ) -> None:
        self._client_id = client_id
        self._client_secret = client_secret
        self._transport = transport
        self._token_refresh_buffer_seconds = token_refresh_buffer_seconds
        self._access_token: str | None = None
        self._refresh_at = 0.0

    def get_token(self, *, force_refresh: bool = False) -> str:
        if not force_refresh and self._access_token and time.time() < self._refresh_at:
            return self._access_token

        response = self._transport.request_without_auth(
            "/wlte/v1/auth/token",
            method="POST",
            body={
                "clientId": self._client_id,
                "clientSecret": self._client_secret,
            },
        )
        token = response["data"]
        self._access_token = token["accessToken"]
        self._refresh_at = time.time() + int(token["expiresIn"]) - self._token_refresh_buffer_seconds
        return self._access_token
