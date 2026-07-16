package wlteopenapi

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestDeviceManagementMethods(t *testing.T) {
	var requests []struct {
		method string
		path   string
		body   string
	}
	client, err := NewClient(ClientOptions{
		ClientID: "client", ClientSecret: "secret", BaseURL: "https://api.test",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/wlte/v1/auth/token" {
				return jsonResponse(200, `{"code":"SUCCESS","data":{"accessToken":"token","expiresIn":3600}}`, nil), nil
			}
			var body []byte
			if req.Body != nil {
				body, _ = io.ReadAll(req.Body)
			}
			requests = append(requests, struct {
				method string
				path   string
				body   string
			}{req.Method, req.URL.Path, string(body)})
			switch req.Method {
			case http.MethodPost:
				return jsonResponse(201, `{"code":"SUCCESS","data":{"deviceId":"dev-1","name":"Demo"}}`, nil), nil
			case http.MethodDelete:
				return jsonResponse(200, `{"code":"SUCCESS","data":{"deviceId":"dev-1"}}`, nil), nil
			default:
				return jsonResponse(200, `{"code":"SUCCESS","data":{"deviceId":"dev-1","updated":true}}`, nil), nil
			}
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Devices.Add(context.Background(), AddDeviceOptions{DeviceID: "dev-1", Password: "1234", Name: "Demo"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Devices.Remove(context.Background(), "dev-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Devices.ModifyPassword(context.Background(), "dev-1", ModifyDevicePasswordOptions{OldPassword: "1234", NewPassword: "5678"}); err != nil {
		t.Fatal(err)
	}
	if len(requests) != 3 {
		t.Fatalf("unexpected requests: %+v", requests)
	}
	if requests[0].method != http.MethodPost || requests[0].path != "/wlte/v1/devices" || requests[0].body != `{"deviceId":"dev-1","name":"Demo","password":"1234"}` {
		t.Fatalf("unexpected add request: %+v", requests[0])
	}
	if requests[1].method != http.MethodDelete || requests[1].path != "/wlte/v1/devices/dev-1" {
		t.Fatalf("unexpected remove request: %+v", requests[1])
	}
	if requests[2].method != http.MethodPut || requests[2].path != "/wlte/v1/devices/dev-1/password" || requests[2].body != `{"newPassword":"5678","oldPassword":"1234"}` {
		t.Fatalf("unexpected password request: %+v", requests[2])
	}
}
