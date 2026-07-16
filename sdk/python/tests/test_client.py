import json
import threading
import time
import unittest
from concurrent.futures import ThreadPoolExecutor

from wlte_openapi import WlteApiError, WlteClient
from wlte_openapi.client import TransportResponse
from wlte_openapi.auth import AuthManager


class FakeTransport:
    def __init__(self, responses):
        self.responses = list(responses)
        self.calls = []

    def send(self, *, method, url, headers, body):
        self.calls.append({"method": method, "url": url, "headers": headers, "body": body})
        return self.responses.pop(0)


def response(body, status=200, headers=None):
    return TransportResponse(
        status=status,
        headers=headers or {},
        body=json.dumps(body),
    )


class WlteClientTest(unittest.TestCase):
    def test_lists_devices_with_automatic_token_request(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response(
                    {
                        "code": "SUCCESS",
                        "message": "ok",
                        "data": {
                            "devices": [{"deviceId": "device-1", "name": "Device 1", "status": "ONLINE"}],
                            "stats": {"total": 1, "online": 1, "offline": 0},
                            "pagination": {"page": 1, "pageSize": 50, "total": 1, "hasNext": False, "hasPrev": False},
                        },
                    }
                ),
            ]
        )
        client = WlteClient(
            client_id="client",
            client_secret="secret",
            base_url="https://api.test",
            transport=transport,
        )

        result = client.devices.list(page=1, page_size=50)

        self.assertEqual(result["devices"][0]["deviceId"], "device-1")
        self.assertEqual(len(transport.calls), 2)

    def test_retries_once_on_auth_expired(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token-1", "expiresIn": 3600}}),
                response({"code": "AUTH_EXPIRED", "message": "expired", "data": None}, status=401),
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token-2", "expiresIn": 3600}}),
                response({"code": "SUCCESS", "message": "ok", "data": {"deviceId": "device-1", "name": "Device 1", "status": "ONLINE"}}),
            ]
        )
        client = WlteClient(
            client_id="client",
            client_secret="secret",
            base_url="https://api.test",
            transport=transport,
        )

        device = client.devices.get("device-1")

        self.assertEqual(device["deviceId"], "device-1")
        self.assertEqual(len(transport.calls), 4)

    def test_parses_rate_limit_errors(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response(
                    {"code": "RATE_LIMITED", "message": "too many requests", "requestId": "req-rate", "data": None},
                    status=429,
                    headers={"Retry-After": "5"},
                ),
            ]
        )
        client = WlteClient(
            client_id="client",
            client_secret="secret",
            base_url="https://api.test",
            transport=transport,
        )

        with self.assertRaises(WlteApiError) as context:
            client.devices.list()

        self.assertEqual(context.exception.status, 429)
        self.assertEqual(context.exception.code, "RATE_LIMITED")
        self.assertEqual(context.exception.retry_after, "5")
        self.assertEqual(context.exception.request_id, "req-rate")

    def test_maps_relay_set_requests_to_relay_command_action(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response(
                    {
                        "code": "COMMAND_ACCEPTED",
                        "message": "accepted",
                        "data": {
                            "command": {
                                "id": "cmd-1",
                                "deviceId": "device-1",
                                "operation": "device.relay.set",
                                "status": "SUCCESS",
                                "params": {"relays": [{"index": 1, "action": "ON"}]},
                                "createdAt": "2026-07-15T00:00:00Z",
                            },
                            "state": {
                                "deviceId": "device-1",
                                "status": "ONLINE",
                                "peripherals": {"relays": [{"index": 1, "on": True}]},
                            },
                        },
                    },
                    status=202,
                ),
            ]
        )
        client = WlteClient(
            client_id="client",
            client_secret="secret",
            base_url="https://api.test",
            transport=transport,
        )

        execution = client.relays.set("device-1", {"index": 1, "on": True, "idempotencyKey": "idem-1"})

        self.assertEqual(execution["command"]["operation"], "device.relay.set")
        self.assertTrue(execution["state"]["peripherals"]["relays"][0]["on"])
        self.assertEqual(transport.calls[1]["url"], "https://api.test/wlte/v1/devices/device-1/relays/commands")
        self.assertEqual(json.loads(transport.calls[1]["body"]), {"relays": [{"index": 1, "action": "ON"}]})

    def test_controls_multiple_relays_in_one_request(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response(
                    {
                        "code": "COMMAND_ACCEPTED",
                        "message": "accepted",
                        "data": {
                            "command": {
                                "id": "cmd-2",
                                "deviceId": "device-1",
                                "operation": "device.relay.set",
                                "status": "SUCCESS",
                                "params": {"relays": [{"index": 1, "action": "ON"}, {"index": 2, "action": "OFF"}]},
                                "createdAt": "2026-07-15T00:00:00Z",
                            }
                        },
                    },
                    status=202,
                ),
            ]
        )
        client = WlteClient(client_id="client", client_secret="secret", base_url="https://api.test", transport=transport)

        execution = client.relays.control(
            "device-1",
            {
                "relays": [{"index": 1, "action": "ON"}, {"index": 2, "action": "OFF"}],
                "idempotencyKey": "idem-multi",
            },
        )

        self.assertEqual(execution["command"]["operation"], "device.relay.set")
        self.assertEqual(
            json.loads(transport.calls[1]["body"]),
            {"relays": [{"index": 1, "action": "ON"}, {"index": 2, "action": "OFF"}]},
        )

    def test_lists_device_profiles(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response(
                    {
                        "code": "SUCCESS",
                        "message": "ok",
                        "data": {
                            "profiles": [
                                {
                                    "deviceType": "RL1",
                                    "capabilities": {
                                    "relayCount": 1,
                                    "supportedOperations": ["device.relay.set"],
                                        "operationSpecs": {"relay": {"actions": ["ON", "OFF", "JOG"]}},
                                    },
                                }
                            ]
                        },
                    }
                ),
            ]
        )
        client = WlteClient(
            client_id="client",
            client_secret="secret",
            base_url="https://api.test",
            transport=transport,
        )

        result = client.profiles.list()

        self.assertEqual(result["profiles"][0]["deviceType"], "RL1")
        self.assertEqual(result["profiles"][0]["capabilities"]["supportedOperations"], ["device.relay.set"])

    def test_supports_device_management_methods(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response({"code": "SUCCESS", "data": {"deviceId": "dev-1", "name": "Demo"}}, status=201),
                response({"code": "SUCCESS", "data": {"deviceId": "dev-1"}}),
                response({"code": "SUCCESS", "data": {"deviceId": "dev-1", "updated": True}}),
            ]
        )
        client = WlteClient(client_id="client", client_secret="secret", base_url="https://api.test", transport=transport)

        client.devices.add({"deviceId": "dev-1", "password": "1234", "name": "Demo"})
        client.devices.remove("dev-1")
        client.devices.modify_password("dev-1", {"oldPassword": "1234", "newPassword": "5678"})

        self.assertEqual(transport.calls[1]["method"], "POST")
        self.assertEqual(transport.calls[2]["method"], "DELETE")
        self.assertEqual(transport.calls[3]["method"], "PUT")
        self.assertEqual(json.loads(transport.calls[3]["body"]), {"oldPassword": "1234", "newPassword": "5678"})

    def test_auth_manager_shares_concurrent_token_request(self):
        class ConcurrentTransport:
            def __init__(self):
                self.calls = 0
                self.lock = threading.Lock()

            def request_without_auth(self, path, *, method, body=None):
                with self.lock:
                    self.calls += 1
                time.sleep(0.02)
                return {"code": "SUCCESS", "data": {"accessToken": "shared-token", "expiresIn": 3600}}

        transport = ConcurrentTransport()
        auth = AuthManager(client_id="client", client_secret="secret", transport=transport)
        with ThreadPoolExecutor(max_workers=10) as executor:
            tokens = list(executor.map(lambda _: auth.get_token(), range(10)))

        self.assertEqual(tokens, ["shared-token"] * 10)
        self.assertEqual(transport.calls, 1)


    def test_gets_device_config(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response(
                    {
                        "code": "SUCCESS",
                        "message": "ok",
                        "data": {"relay": {"channels": [{"index": 1, "jogTimeSeconds": 2}]}, "rs485": {"baudRate": 9600}},
                    }
                ),
            ]
        )
        client = WlteClient(client_id="client", client_secret="secret", base_url="https://api.test", transport=transport)

        config = client.devices.get_config("device-1")

        self.assertEqual(config["relay"]["channels"][0]["jogTimeSeconds"], 2)
        self.assertEqual(transport.calls[1]["url"], "https://api.test/wlte/v1/devices/device-1/config")

    def test_sets_relay_jog_config(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response(
                    {
                        "code": "COMMAND_ACCEPTED",
                        "message": "accepted",
                        "data": {"command": {"id": "cmd-1", "deviceId": "device-1", "operation": "device.relay.jogConfig.set", "status": "SUCCESS", "params": {"relayIndex": 1, "durationSec": 2}, "result": {"relayIndex": 1, "durationSec": 2}, "createdAt": "2026-07-15T00:00:00Z"}},
                    },
                    status=202,
                ),
            ]
        )
        client = WlteClient(client_id="client", client_secret="secret", base_url="https://api.test", transport=transport)

        execution = client.relays.set_jog_config("device-1", {"index": 1, "durationSec": 2, "idempotencyKey": "idem-jog"})

        self.assertEqual(execution["command"]["operation"], "device.relay.jogConfig.set")
        self.assertEqual(transport.calls[1]["method"], "PUT")
        self.assertEqual(transport.calls[1]["url"], "https://api.test/wlte/v1/devices/device-1/relays/1/jog-config")
        self.assertEqual(json.loads(transport.calls[1]["body"]), {"durationSec": 2})

    def test_sends_rs485_transceive_requests(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response(
                    {
                        "code": "COMMAND_ACCEPTED",
                        "message": "accepted",
                        "data": {"command": {"id": "cmd-485", "deviceId": "device-1", "operation": "device.rs485.transceive", "status": "SUCCESS", "params": {"requestHex": "020600340000C837"}, "result": {"responseHex": "020600340000C837"}, "createdAt": "2026-07-15T00:00:00Z"}},
                    },
                    status=202,
                ),
            ]
        )
        client = WlteClient(client_id="client", client_secret="secret", base_url="https://api.test", transport=transport)

        execution = client.rs485.transceive("device-1", {"requestHex": "020600340000C837", "idempotencyKey": "idem-485"})

        self.assertEqual(execution["command"]["operation"], "device.rs485.transceive")
        self.assertEqual(transport.calls[1]["method"], "POST")
        self.assertEqual(transport.calls[1]["url"], "https://api.test/wlte/v1/devices/device-1/rs485/transceive")
        self.assertEqual(json.loads(transport.calls[1]["body"]), {"requestHex": "020600340000C837"})

    def test_sets_rs485_baud_rate(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response(
                    {
                        "code": "COMMAND_ACCEPTED",
                        "message": "accepted",
                        "data": {"command": {"id": "cmd-baud", "deviceId": "device-1", "operation": "device.rs485.baudRate.set", "status": "SUCCESS", "params": {"baudRate": 9600}, "result": {"baudRate": 9600}, "createdAt": "2026-07-15T00:00:00Z"}},
                    },
                    status=202,
                ),
            ]
        )
        client = WlteClient(client_id="client", client_secret="secret", base_url="https://api.test", transport=transport)

        execution = client.rs485.set_baud_rate("device-1", {"baudRate": 9600, "idempotencyKey": "idem-baud"})

        self.assertEqual(execution["command"]["operation"], "device.rs485.baudRate.set")
        self.assertEqual(transport.calls[1]["method"], "PUT")
        self.assertEqual(transport.calls[1]["url"], "https://api.test/wlte/v1/devices/device-1/rs485/baud-rate")
        self.assertEqual(json.loads(transport.calls[1]["body"]), {"baudRate": 9600})


if __name__ == "__main__":
    unittest.main()
