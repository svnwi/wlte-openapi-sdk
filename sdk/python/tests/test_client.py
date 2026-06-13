import json
import unittest

from wlte_openapi import WlteApiError, WlteClient
from wlte_openapi.client import TransportResponse


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
                    {"code": "RATE_LIMITED", "message": "too many requests", "data": None},
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

    def test_maps_relay_set_requests_to_relay_command_action(self):
        transport = FakeTransport(
            [
                response({"code": "SUCCESS", "message": "ok", "data": {"accessToken": "token", "expiresIn": 3600}}),
                response(
                    {
                        "code": "COMMAND_ACCEPTED",
                        "message": "accepted",
                        "data": {"id": "cmd-1", "deviceId": "device-1", "relayIndex": 1, "action": "ON"},
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

        command = client.relays.set("device-1", {"index": 1, "on": True, "idempotencyKey": "idem-1"})

        self.assertEqual(command["action"], "ON")
        self.assertEqual(transport.calls[1]["url"], "https://api.test/wlte/v1/devices/device-1/relays/1/commands")
        self.assertEqual(json.loads(transport.calls[1]["body"]), {"action": "ON"})

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
                        "data": {"id": "cmd-1", "deviceId": "device-1", "type": "RELAY_JOG_CONFIG_SET", "result": {"relayIndex": 1, "durationSec": 2}},
                    },
                    status=202,
                ),
            ]
        )
        client = WlteClient(client_id="client", client_secret="secret", base_url="https://api.test", transport=transport)

        command = client.relays.set_jog_config("device-1", {"index": 1, "durationSec": 2, "idempotencyKey": "idem-jog"})

        self.assertEqual(command["type"], "RELAY_JOG_CONFIG_SET")
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
                        "data": {"id": "cmd-485", "deviceId": "device-1", "type": "RS485_TRANSCEIVE", "result": {"requestHex": "020600340000C837"}},
                    },
                    status=202,
                ),
            ]
        )
        client = WlteClient(client_id="client", client_secret="secret", base_url="https://api.test", transport=transport)

        command = client.rs485.transceive("device-1", {"requestHex": "020600340000C837", "idempotencyKey": "idem-485"})

        self.assertEqual(command["type"], "RS485_TRANSCEIVE")
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
                        "data": {"id": "cmd-baud", "deviceId": "device-1", "type": "RS485_BAUD_RATE_SET", "result": {"baudRate": 9600}},
                    },
                    status=202,
                ),
            ]
        )
        client = WlteClient(client_id="client", client_secret="secret", base_url="https://api.test", transport=transport)

        command = client.rs485.set_baud_rate("device-1", {"baudRate": 9600, "idempotencyKey": "idem-baud"})

        self.assertEqual(command["type"], "RS485_BAUD_RATE_SET")
        self.assertEqual(transport.calls[1]["method"], "PUT")
        self.assertEqual(transport.calls[1]["url"], "https://api.test/wlte/v1/devices/device-1/rs485/baud-rate")
        self.assertEqual(json.loads(transport.calls[1]["body"]), {"baudRate": 9600})


if __name__ == "__main__":
    unittest.main()
