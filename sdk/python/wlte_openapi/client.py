from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Protocol
from urllib import parse, request
from urllib.error import HTTPError

from .auth import AuthManager
from .commands import CommandsApi
from .devices import DevicesApi
from .errors import WlteApiError, is_auth_expired
from .profiles import ProfilesApi
from .relays import RelaysApi
from .rs485 import RS485Api

DEFAULT_BASE_URL = "https://openapi.svnwi.com"
SUCCESS_CODES = {"SUCCESS", "COMMAND_ACCEPTED", "OK"}


@dataclass
class TransportResponse:
    status: int
    headers: dict[str, str]
    body: str
    reason: str = ""


class Transport(Protocol):
    def send(self, *, method: str, url: str, headers: dict[str, str], body: str | None) -> TransportResponse:
        ...


class UrlLibTransport:
    def send(self, *, method: str, url: str, headers: dict[str, str], body: str | None) -> TransportResponse:
        req = request.Request(
            url,
            data=body.encode("utf-8") if body is not None else None,
            headers=headers,
            method=method,
        )

        try:
            with request.urlopen(req) as response:
                return TransportResponse(
                    status=response.status,
                    headers=dict(response.headers.items()),
                    body=response.read().decode("utf-8"),
                    reason=response.reason,
                )
        except HTTPError as error:
            return TransportResponse(
                status=error.code,
                headers=dict(error.headers.items()),
                body=error.read().decode("utf-8"),
                reason=error.reason,
            )


class WlteClient:
    def __init__(
        self,
        *,
        client_id: str,
        client_secret: str,
        base_url: str | None = None,
        transport: Transport | None = None,
        token_refresh_buffer_seconds: int = 60,
    ) -> None:
        if not client_id:
            raise ValueError("client_id is required")
        if not client_secret:
            raise ValueError("client_secret is required")

        self._base_url = (base_url or DEFAULT_BASE_URL).rstrip("/")
        self._transport = transport or UrlLibTransport()
        self._auth = AuthManager(
            client_id=client_id,
            client_secret=client_secret,
            transport=self,
            token_refresh_buffer_seconds=token_refresh_buffer_seconds,
        )
        self.devices = DevicesApi(self)
        self.profiles = ProfilesApi(self)
        self.relays = RelaysApi(self)
        self.commands = CommandsApi(self)
        self.rs485 = RS485Api(self)

    def request(
        self,
        path: str,
        *,
        method: str = "GET",
        query: dict[str, str | int | bool | None] | None = None,
        headers: dict[str, str | None] | None = None,
        body: object | None = None,
    ) -> dict:
        token = self._auth.get_token()
        try:
            return self._send(path, method=method, query=query, headers=headers, body=body, token=token)
        except WlteApiError as error:
            if is_auth_expired(error):
                refreshed_token = self._auth.get_token(force_refresh=True)
                return self._send(path, method=method, query=query, headers=headers, body=body, token=refreshed_token)
            raise

    def request_without_auth(self, path: str, *, method: str, body: object | None = None) -> dict:
        return self._send(path, method=method, body=body)

    def _send(
        self,
        path: str,
        *,
        method: str,
        query: dict[str, str | int | bool | None] | None = None,
        headers: dict[str, str | None] | None = None,
        body: object | None = None,
        token: str | None = None,
    ) -> dict:
        request_headers = {"Accept": "application/json"}
        encoded_body = None

        if body is not None:
            request_headers["Content-Type"] = "application/json"
            encoded_body = json.dumps(body)

        if token:
            request_headers["Authorization"] = f"Bearer {token}"

        for key, value in (headers or {}).items():
            if value is not None:
                request_headers[key] = value

        response = self._transport.send(
            method=method,
            url=self._build_url(path, query=query),
            headers=request_headers,
            body=encoded_body,
        )
        payload = self._parse_body(response.body)

        if response.status < 200 or response.status >= 300:
            raise self._to_api_error(response, payload)

        if self._is_error_envelope(payload):
            raise self._to_api_error(response, payload)

        return payload

    def _build_url(self, path: str, *, query: dict[str, str | int | bool | None] | None = None) -> str:
        url = f"{self._base_url}{path}"
        query_items = {key: value for key, value in (query or {}).items() if value is not None}
        if not query_items:
            return url
        return f"{url}?{parse.urlencode(query_items)}"

    @staticmethod
    def _parse_body(body: str) -> object:
        if not body:
            return None
        try:
            return json.loads(body)
        except json.JSONDecodeError:
            return body

    @staticmethod
    def _is_error_envelope(payload: object) -> bool:
        return (
            isinstance(payload, dict)
            and isinstance(payload.get("code"), str)
            and payload.get("code") not in SUCCESS_CODES
            and isinstance(payload.get("message"), str)
        )

    @staticmethod
    def _to_api_error(response: TransportResponse, payload: object) -> WlteApiError:
        retry_after = response.headers.get("Retry-After") or response.headers.get("retry-after")

        if isinstance(payload, dict):
            return WlteApiError(
                status=response.status,
                code=str(payload.get("code") or ("RATE_LIMITED" if response.status == 429 else "HTTP_ERROR")),
                message=str(payload.get("message") or response.reason),
                data=payload.get("data"),
                retry_after=retry_after,
            )

        return WlteApiError(
            status=response.status,
            code="RATE_LIMITED" if response.status == 429 else "HTTP_ERROR",
            message=str(payload or response.reason),
            retry_after=retry_after,
        )
