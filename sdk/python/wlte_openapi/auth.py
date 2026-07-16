from __future__ import annotations

import threading
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
        self._condition = threading.Condition()
        self._refreshing = False
        self._refresh_error: BaseException | None = None
        if token_refresh_buffer_seconds < 0:
            raise ValueError("token_refresh_buffer_seconds must not be negative")

    def get_token(self, *, rejected_token: str | None = None) -> str:
        with self._condition:
            if self._access_token and self._access_token != rejected_token and time.time() < self._refresh_at:
                return self._access_token
            if self._refreshing:
                while self._refreshing:
                    self._condition.wait()
                if self._refresh_error is not None:
                    raise self._refresh_error
                if self._access_token is None:
                    raise RuntimeError("token refresh completed without an access token")
                return self._access_token
            self._refreshing = True
            self._refresh_error = None

        try:
            response = self._transport.request_without_auth(
                "/wlte/v1/auth/token",
                method="POST",
                body={
                    "clientId": self._client_id,
                    "clientSecret": self._client_secret,
                },
            )
            token = response["data"]
            access_token = str(token.get("accessToken") or "")
            expires_in = int(token.get("expiresIn") or 0)
            if not access_token:
                raise ValueError("token response did not contain accessToken")
            if expires_in <= 0:
                raise ValueError("token response expiresIn must be greater than zero")
            effective_buffer = min(self._token_refresh_buffer_seconds, expires_in / 5)
            with self._condition:
                self._access_token = access_token
                self._refresh_at = time.time() + expires_in - effective_buffer
                return access_token
        except BaseException as error:
            with self._condition:
                self._refresh_error = error
            raise
        finally:
            with self._condition:
                self._refreshing = False
                self._condition.notify_all()
