package wlteopenapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	defaultWebSocketEventBuffer = 64
	defaultWebSocketReadLimit   = 1 << 20

	WebSocketTopicSessionPing             = "session.ping"
	WebSocketTopicDeviceStateGet          = "device.state.get"
	WebSocketTopicDeviceOperationExecute  = "device.operation.execute"
	WebSocketTopicDeviceConnectionOnline  = "device.connection.online"
	WebSocketTopicDeviceConnectionOffline = "device.connection.offline"
	WebSocketTopicDeviceStateChanged      = "device.state.changed"
	WebSocketTopicDevicePowerLost         = "device.power.lost"
	WebSocketTopicDeviceRS485Received     = "device.rs485.received"

	WebSocketStateChangeRelay        = "relay"
	WebSocketStateChangeDigitalInput = "digitalInput"
)

var ErrWebSocketEventBufferFull = errors.New("wlte websocket event buffer is full")

type WebSocketService struct {
	client *Client
}

type WebSocketTicket struct {
	Ticket    string `json:"ticket"`
	ExpiresIn int    `json:"expiresIn"`
}

type WebSocketConnectOptions struct {
	EventBuffer int
	ReadLimit   int64
	HTTPHeader  http.Header
}

type WebSocketConnectError struct {
	Status int
}

func (e *WebSocketConnectError) Error() string {
	if e.Status == 0 {
		return "wlte websocket connection failed"
	}
	return fmt.Sprintf("wlte websocket connection failed: status=%d", e.Status)
}

func (s *WebSocketService) CreateTicket(ctx context.Context) (WebSocketTicket, error) {
	var result WebSocketTicket
	err := s.client.request(ctx, http.MethodPost, "/wlte/v1/ws/ticket", nil, nil, map[string]any{}, &result)
	if err != nil {
		return WebSocketTicket{}, err
	}
	if result.Ticket == "" || result.ExpiresIn <= 0 {
		return WebSocketTicket{}, fmt.Errorf("invalid WebSocket ticket response")
	}
	return result, nil
}

func (s *WebSocketService) Connect(ctx context.Context, options WebSocketConnectOptions) (*WebSocketSession, error) {
	if options.EventBuffer < 0 {
		return nil, fmt.Errorf("WebSocket event buffer must not be negative")
	}
	if options.EventBuffer == 0 {
		options.EventBuffer = defaultWebSocketEventBuffer
	}
	if options.ReadLimit < 0 {
		return nil, fmt.Errorf("WebSocket read limit must not be negative")
	}
	if options.ReadLimit == 0 {
		options.ReadLimit = defaultWebSocketReadLimit
	}

	ticket, err := s.CreateTicket(ctx)
	if err != nil {
		return nil, err
	}
	connectionURL, err := s.connectionURL(ticket.Ticket)
	if err != nil {
		return nil, err
	}
	conn, response, err := websocket.Dial(ctx, connectionURL, &websocket.DialOptions{
		HTTPClient: s.client.httpClient,
		HTTPHeader: options.HTTPHeader.Clone(),
	})
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode
			body, readErr := io.ReadAll(io.LimitReader(response.Body, defaultWebSocketReadLimit))
			_ = response.Body.Close()
			if readErr == nil {
				if responseErr := decodeResponse(response, body, nil); responseErr != nil {
					return nil, responseErr
				}
			}
		}
		return nil, &WebSocketConnectError{Status: status}
	}
	conn.SetReadLimit(options.ReadLimit)
	return newWebSocketSession(conn, options.EventBuffer), nil
}

func (s *WebSocketService) connectionURL(ticket string) (string, error) {
	parsed, err := url.Parse(s.client.baseURL)
	if err != nil {
		return "", err
	}
	switch parsed.Scheme {
	case "https":
		parsed.Scheme = "wss"
	case "http":
		parsed.Scheme = "ws"
	default:
		return "", fmt.Errorf("base URL must use http or https")
	}
	parsed.Path = "/wlte/v1/ws"
	parsed.RawQuery = url.Values{"ticket": []string{ticket}}.Encode()
	parsed.Fragment = ""
	return parsed.String(), nil
}

type WebSocketReply struct {
	RequestID string          `json:"requestId"`
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
}

func (r WebSocketReply) Decode(out any) error {
	return decodeWebSocketData(r.Data, out)
}

type WebSocketEvent struct {
	Topic string          `json:"topic"`
	Data  json.RawMessage `json:"data"`
}

type WebSocketDeviceConnectionEvent struct {
	DeviceID   string `json:"deviceId"`
	OccurredAt string `json:"occurredAt,omitempty"`
}

type WebSocketStateChange struct {
	Type    string `json:"type"`
	Indexes []int  `json:"indexes"`
}

type WebSocketDeviceStateChangedEvent struct {
	DeviceID       string                 `json:"deviceId"`
	OccurredAt     string                 `json:"occurredAt,omitempty"`
	Changes        []WebSocketStateChange `json:"changes,omitempty"`
	CorrelationID  string                 `json:"correlationId,omitempty"`
	Peripherals    *Peripherals           `json:"peripherals,omitempty"`
	StateUpdatedAt string                 `json:"stateUpdatedAt,omitempty"`
}

type WebSocketDevicePowerLostEvent struct {
	DeviceID   string `json:"deviceId"`
	OccurredAt string `json:"occurredAt,omitempty"`
	Message    string `json:"message,omitempty"`
}

type WebSocketRS485ReceivedEvent struct {
	DeviceID   string `json:"deviceId"`
	OccurredAt string `json:"occurredAt,omitempty"`
	RS485      struct {
		ResponseHex string `json:"responseHex"`
	} `json:"rs485"`
}

func (e WebSocketEvent) Decode(out any) error {
	return decodeWebSocketData(e.Data, out)
}

type WebSocketOperationRequest struct {
	DeviceID       string
	IdempotencyKey string
	Operation      CommandOperation
	Params         any
}

type WebSocketSession struct {
	conn *websocket.Conn

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	events chan WebSocketEvent

	pendingMu sync.Mutex
	pending   map[string]chan WebSocketReply

	errMu sync.RWMutex
	err   error

	droppedEvents atomic.Uint64
	finishOnce    sync.Once
}

func newWebSocketSession(conn *websocket.Conn, eventBuffer int) *WebSocketSession {
	ctx, cancel := context.WithCancel(context.Background())
	session := &WebSocketSession{
		conn:    conn,
		ctx:     ctx,
		cancel:  cancel,
		done:    make(chan struct{}),
		events:  make(chan WebSocketEvent, eventBuffer),
		pending: make(map[string]chan WebSocketReply),
	}
	go session.readLoop()
	return session
}

func (s *WebSocketSession) Request(ctx context.Context, topic string, data any) (WebSocketReply, error) {
	return s.RequestWithID(ctx, "sdk_"+randomRequestID(), topic, data)
}

// RequestWithID sends a request using a caller-provided correlation ID. It is
// useful for protocol observers that need to expose the outbound request before
// its reply arrives. Most callers should continue to use Request.
func (s *WebSocketSession) RequestWithID(ctx context.Context, requestID, topic string, data any) (WebSocketReply, error) {
	if strings.TrimSpace(requestID) == "" {
		return WebSocketReply{}, fmt.Errorf("WebSocket request ID is required")
	}
	if strings.TrimSpace(topic) == "" {
		return WebSocketReply{}, fmt.Errorf("WebSocket topic is required")
	}
	replyCh := make(chan WebSocketReply, 1)

	s.pendingMu.Lock()
	select {
	case <-s.done:
		s.pendingMu.Unlock()
		return WebSocketReply{}, s.closedError()
	default:
	}
	if _, exists := s.pending[requestID]; exists {
		s.pendingMu.Unlock()
		return WebSocketReply{}, fmt.Errorf("WebSocket request ID %q is already pending", requestID)
	}
	s.pending[requestID] = replyCh
	s.pendingMu.Unlock()
	defer func() {
		s.pendingMu.Lock()
		delete(s.pending, requestID)
		s.pendingMu.Unlock()
	}()

	request := map[string]any{
		"type":      "request",
		"requestId": requestID,
		"topic":     topic,
		"data":      data,
	}
	if err := wsjson.Write(ctx, s.conn, request); err != nil {
		return WebSocketReply{}, fmt.Errorf("write WebSocket request: %w", err)
	}

	select {
	case <-ctx.Done():
		return WebSocketReply{}, ctx.Err()
	case <-s.done:
		return WebSocketReply{}, s.closedError()
	case reply := <-replyCh:
		if _, ok := successCodes[reply.Code]; !ok {
			return WebSocketReply{}, &APIError{
				Code:      reply.Code,
				Message:   reply.Message,
				Data:      reply.Data,
				RequestID: reply.RequestID,
			}
		}
		return reply, nil
	}
}

func (s *WebSocketSession) ProtocolPing(ctx context.Context) error {
	return s.conn.Ping(ctx)
}

func (s *WebSocketSession) Ping(ctx context.Context) error {
	reply, err := s.Request(ctx, WebSocketTopicSessionPing, map[string]any{})
	if err != nil {
		return err
	}
	var data map[string]any
	return reply.Decode(&data)
}

func (s *WebSocketSession) GetDeviceState(ctx context.Context, deviceID string) (Device, error) {
	var device Device
	reply, err := s.Request(ctx, WebSocketTopicDeviceStateGet, map[string]string{"deviceId": deviceID})
	if err != nil {
		return Device{}, err
	}
	if err := reply.Decode(&device); err != nil {
		return Device{}, err
	}
	return device, nil
}

func (s *WebSocketSession) ExecuteOperation(ctx context.Context, request WebSocketOperationRequest) (CommandExecution, error) {
	if strings.TrimSpace(request.DeviceID) == "" {
		return CommandExecution{}, fmt.Errorf("device ID is required")
	}
	if strings.TrimSpace(request.IdempotencyKey) == "" {
		return CommandExecution{}, fmt.Errorf("idempotency key is required")
	}
	if strings.TrimSpace(string(request.Operation)) == "" {
		return CommandExecution{}, fmt.Errorf("operation is required")
	}
	data := map[string]any{
		"deviceId":       request.DeviceID,
		"idempotencyKey": request.IdempotencyKey,
		"operation": map[string]any{
			"name":   request.Operation,
			"params": request.Params,
		},
	}
	reply, err := s.Request(ctx, WebSocketTopicDeviceOperationExecute, data)
	if err != nil {
		return CommandExecution{}, err
	}
	var execution CommandExecution
	if err := reply.Decode(&execution); err != nil {
		return CommandExecution{}, err
	}
	return execution, nil
}

func (s *WebSocketSession) Events() <-chan WebSocketEvent {
	return s.events
}

func (s *WebSocketSession) Done() <-chan struct{} {
	return s.done
}

func (s *WebSocketSession) Err() error {
	s.errMu.RLock()
	defer s.errMu.RUnlock()
	return s.err
}

func (s *WebSocketSession) DroppedEvents() uint64 {
	return s.droppedEvents.Load()
}

func (s *WebSocketSession) Close() error {
	err := s.conn.Close(websocket.StatusNormalClosure, "client closed")
	select {
	case <-s.done:
	case <-time.After(5 * time.Second):
		_ = s.conn.CloseNow()
	}
	return err
}

func (s *WebSocketSession) readLoop() {
	for {
		var envelope struct {
			Type      string          `json:"type"`
			RequestID string          `json:"requestId"`
			Topic     string          `json:"topic"`
			Code      string          `json:"code"`
			Message   string          `json:"message"`
			Data      json.RawMessage `json:"data"`
		}
		if err := wsjson.Read(s.ctx, s.conn, &envelope); err != nil {
			s.finish(err)
			return
		}
		switch envelope.Type {
		case "reply":
			reply := WebSocketReply{
				RequestID: envelope.RequestID,
				Code:      envelope.Code,
				Message:   envelope.Message,
				Data:      envelope.Data,
			}
			s.pendingMu.Lock()
			replyCh := s.pending[envelope.RequestID]
			s.pendingMu.Unlock()
			if replyCh != nil {
				select {
				case replyCh <- reply:
				default:
				}
			}
		case "event":
			event := WebSocketEvent{Topic: envelope.Topic, Data: envelope.Data}
			select {
			case s.events <- event:
			default:
				s.droppedEvents.Add(1)
				_ = s.conn.CloseNow()
				s.finish(ErrWebSocketEventBufferFull)
				return
			}
		}
	}
}

func (s *WebSocketSession) finish(err error) {
	s.finishOnce.Do(func() {
		status := websocket.CloseStatus(err)
		if status == websocket.StatusNormalClosure {
			err = nil
		}
		s.errMu.Lock()
		s.err = err
		s.errMu.Unlock()
		s.cancel()
		close(s.done)
		close(s.events)
	})
}

func (s *WebSocketSession) closedError() error {
	if err := s.Err(); err != nil {
		return fmt.Errorf("WebSocket session closed: %w", err)
	}
	return fmt.Errorf("WebSocket session closed")
}

func decodeWebSocketData(data json.RawMessage, out any) error {
	if out == nil || len(data) == 0 || string(data) == "null" {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode WebSocket data: %w", err)
	}
	return nil
}

func randomRequestID() string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(raw[:])
}
