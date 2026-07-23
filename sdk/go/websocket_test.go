package wlteopenapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestWebSocketSessionReusesSDKAuthAndDispatchesEvents(t *testing.T) {
	var tokenCalls, ticketCalls int
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wlte/v1/auth/token", func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		writeJSONResponse(w, `{"code":"SUCCESS","message":"OK.","requestId":"req-token","data":{"accessToken":"token","expiresIn":3600,"tokenType":"Bearer","scopes":["device:read","device:control"]}}`)
	})
	mux.HandleFunc("GET /wlte/v1/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing authorization on devices request")
		}
		writeJSONResponse(w, `{"code":"SUCCESS","message":"OK.","data":{"devices":[],"stats":{"total":0,"online":0,"offline":0},"pagination":{"page":1,"pageSize":20,"total":0,"totalPages":0,"hasNext":false,"hasPrev":false}}}`)
	})
	mux.HandleFunc("POST /wlte/v1/ws/ticket", func(w http.ResponseWriter, r *http.Request) {
		ticketCalls++
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("WebSocket ticket did not reuse SDK access token")
		}
		writeJSONResponse(w, `{"code":"SUCCESS","message":"OK.","requestId":"req-ticket","data":{"ticket":"ticket-1","expiresIn":60}}`)
	})
	mux.HandleFunc("GET /wlte/v1/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("ticket") != "ticket-1" {
			t.Fatalf("unexpected WebSocket ticket")
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept websocket: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "done")
		ctx := context.Background()
		_ = wsjson.Write(ctx, conn, map[string]any{
			"type": "event", "topic": "device.connection.online", "data": map[string]any{"deviceId": "device-1"},
		})
		for {
			var request struct {
				RequestID string `json:"requestId"`
				Topic     string `json:"topic"`
			}
			if err := wsjson.Read(ctx, conn, &request); err != nil {
				return
			}
			data := any(map[string]any{})
			code := "SUCCESS"
			message := "OK."
			switch request.Topic {
			case "device.state.get":
				data = map[string]any{"deviceId": "device-1", "name": "Demo", "status": "ONLINE"}
			case "device.operation.execute":
				code = "COMMAND_ACCEPTED"
				message = "Command accepted."
				data = map[string]any{"command": map[string]any{
					"id": "cmd-1", "deviceId": "device-1", "operation": "device.relay.set", "status": "SUCCESS", "createdAt": "2026-07-16T00:00:00Z",
				}}
			case "test.denied":
				code = "AUTH_SCOPE_DENIED"
				message = "permission denied"
				data = map[string]any{"requiredScope": "device:control"}
			}
			_ = wsjson.Write(ctx, conn, map[string]any{
				"type": "reply", "requestId": request.RequestID, "code": code, "message": message, "data": data,
			})
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client, err := NewClient(ClientOptions{
		ClientID: "client", ClientSecret: "secret", BaseURL: server.URL, HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Devices.List(context.Background(), ListDevicesOptions{}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	session, err := client.WebSocket.Connect(ctx, WebSocketConnectOptions{EventBuffer: 4})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if tokenCalls != 1 || ticketCalls != 1 {
		t.Fatalf("unexpected auth calls: token=%d ticket=%d", tokenCalls, ticketCalls)
	}
	if err := session.ProtocolPing(ctx); err != nil {
		t.Fatal(err)
	}
	if err := session.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	reply, err := session.RequestWithID(ctx, "caller-request-1", WebSocketTopicSessionPing, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if reply.RequestID != "caller-request-1" {
		t.Fatalf("unexpected caller request ID: %q", reply.RequestID)
	}
	device, err := session.GetDeviceState(ctx, "device-1")
	if err != nil {
		t.Fatal(err)
	}
	if device.DeviceID != "device-1" || device.Status != "ONLINE" {
		t.Fatalf("unexpected device: %+v", device)
	}
	execution, err := session.ExecuteOperation(ctx, WebSocketOperationRequest{
		DeviceID: "device-1", IdempotencyKey: "idem-1", Operation: CommandOperationRelaySet,
		Params: map[string]any{"relays": []map[string]any{{"index": 1, "action": "ON"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if execution.Command.ID != "cmd-1" {
		t.Fatalf("unexpected execution: %+v", execution)
	}
	select {
	case event := <-session.Events():
		if event.Topic != "device.connection.online" {
			t.Fatalf("unexpected event: %+v", event)
		}
		var eventData WebSocketDeviceConnectionEvent
		if err := event.Decode(&eventData); err != nil || eventData.DeviceID != "device-1" {
			t.Fatalf("unexpected event data: %+v err=%v", eventData, err)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for event")
	}
	_, err = session.Request(ctx, "test.denied", map[string]any{})
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != "AUTH_SCOPE_DENIED" || apiErr.RequestID == "" {
		t.Fatalf("unexpected WebSocket API error: %+v", err)
	}
}

func writeJSONResponse(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

func TestWebSocketDeviceStateChangedEventDecodesChanges(t *testing.T) {
	event := WebSocketEvent{Data: json.RawMessage(`{
		"deviceId":"device-1",
		"changes":[
			{"type":"relay","indexes":[1,3]},
			{"type":"digitalInput","indexes":[2]}
		]
	}`)}

	var data WebSocketDeviceStateChangedEvent
	if err := event.Decode(&data); err != nil {
		t.Fatal(err)
	}
	if len(data.Changes) != 2 || data.Changes[0].Type != WebSocketStateChangeRelay || data.Changes[1].Type != WebSocketStateChangeDigitalInput {
		t.Fatalf("unexpected changes: %+v", data.Changes)
	}
	if data.Changes[0].Indexes[1] != 3 {
		t.Fatalf("unexpected event data: %+v", data)
	}
}

func TestWebSocketEventBufferOverflowClosesSession(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wlte/v1/auth/token", func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, `{"code":"SUCCESS","message":"OK.","data":{"accessToken":"token","expiresIn":3600}}`)
	})
	mux.HandleFunc("POST /wlte/v1/ws/ticket", func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, `{"code":"SUCCESS","message":"OK.","data":{"ticket":"ticket","expiresIn":60}}`)
	})
	mux.HandleFunc("GET /wlte/v1/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		for i := 0; i < 2; i++ {
			if err := wsjson.Write(r.Context(), conn, map[string]any{
				"type": "event", "topic": WebSocketTopicDeviceStateChanged, "data": map[string]any{"deviceId": "device-1"},
			}); err != nil {
				return
			}
		}
		<-r.Context().Done()
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client, err := NewClient(ClientOptions{ClientID: "client", ClientSecret: "secret", BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	session, err := client.WebSocket.Connect(ctx, WebSocketConnectOptions{EventBuffer: 1})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-session.Done():
	case <-ctx.Done():
		t.Fatal("timed out waiting for event overflow")
	}
	if !errors.Is(session.Err(), ErrWebSocketEventBufferFull) {
		t.Fatalf("unexpected session error: %v", session.Err())
	}
	if session.DroppedEvents() != 1 {
		t.Fatalf("dropped events = %d, want 1", session.DroppedEvents())
	}
}

func TestWebSocketConnectPreservesUpgradeAPIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wlte/v1/auth/token", func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, `{"code":"SUCCESS","message":"OK.","data":{"accessToken":"token","expiresIn":3600}}`)
	})
	mux.HandleFunc("POST /wlte/v1/ws/ticket", func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, `{"code":"SUCCESS","message":"OK.","data":{"ticket":"ticket","expiresIn":60}}`)
	})
	mux.HandleFunc("GET /wlte/v1/ws", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		writeJSONResponse(w, `{"code":"RATE_LIMITED","message":"Too many requests","requestId":"req-upgrade","data":{"scope":"ws_connection"}}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client, err := NewClient(ClientOptions{ClientID: "client", ClientSecret: "secret", BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.WebSocket.Connect(context.Background(), WebSocketConnectOptions{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "RATE_LIMITED" || apiErr.RequestID != "req-upgrade" {
		t.Fatalf("unexpected upgrade error: %#v", err)
	}
}
