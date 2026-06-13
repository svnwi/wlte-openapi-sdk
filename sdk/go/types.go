package wlteopenapi

import (
	"encoding/json"
	"time"
)

type Envelope[T any] struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
	Data      T      `json:"data"`
}

type TokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int    `json:"expiresIn"`
}

type Pagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"pageSize"`
	Total      int  `json:"total"`
	TotalPages int  `json:"totalPages"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
}

type DeviceStats struct {
	Total   int `json:"total"`
	Online  int `json:"online"`
	Offline int `json:"offline"`
}

type RelayState struct {
	Index int   `json:"index"`
	On    *bool `json:"on"`
}

type DigitalInputState struct {
	Index  int   `json:"index"`
	Active *bool `json:"active"`
}

type AnalogMeasurement struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type AnalogInputState struct {
	Index       int                `json:"index"`
	Type        string             `json:"type"`
	Value       *float64           `json:"value,omitempty"`
	Unit        string             `json:"unit,omitempty"`
	Status      string             `json:"status"`
	Measurement *AnalogMeasurement `json:"measurement,omitempty"`
}

type SensorState struct {
	Index  int     `json:"index"`
	Type   string  `json:"type"`
	Value  float64 `json:"value"`
	Unit   string  `json:"unit"`
	Status string  `json:"status"`
}

type Peripherals struct {
	Relays        []RelayState        `json:"relays,omitempty"`
	DigitalInputs []DigitalInputState `json:"digitalInputs,omitempty"`
	AnalogInputs  []AnalogInputState  `json:"analogInputs,omitempty"`
	Sensors       []SensorState       `json:"sensors,omitempty"`
}

type Device struct {
	DeviceID       string       `json:"deviceId"`
	Name           string       `json:"name"`
	Status         string       `json:"status"`
	DeviceType     string       `json:"deviceType,omitempty"`
	StateUpdatedAt string       `json:"stateUpdatedAt,omitempty"`
	Peripherals    *Peripherals `json:"peripherals,omitempty"`
}

type DeviceList struct {
	Devices    []Device    `json:"devices"`
	Stats      DeviceStats `json:"stats"`
	Pagination Pagination  `json:"pagination"`
}

type RelayChannelConfig struct {
	Index          int `json:"index"`
	JogTimeSeconds int `json:"jogTimeSeconds,omitempty"`
}

type RelayConfig struct {
	Channels []RelayChannelConfig `json:"channels"`
}

type RS485Config struct {
	BaudRate int `json:"baudRate,omitempty"`
}

type DeviceConfig struct {
	Relay     *RelayConfig `json:"relay,omitempty"`
	RS485     *RS485Config `json:"rs485,omitempty"`
	UpdatedAt string       `json:"updatedAt,omitempty"`
}

type SensorInterface struct {
	Index          int      `json:"index"`
	SupportedTypes []string `json:"supportedTypes"`
}

type RelayOperationSpec struct {
	Actions []RelayAction `json:"actions"`
}

type OperationSpecs struct {
	Relay *RelayOperationSpec `json:"relay,omitempty"`
}

type DeviceProfileCapabilities struct {
	RelayCount        *int              `json:"relayCount,omitempty"`
	DigitalInputCount *int              `json:"digitalInputCount,omitempty"`
	AnalogInputCount  *int              `json:"analogInputCount,omitempty"`
	SensorInterfaces  []SensorInterface `json:"sensorInterfaces,omitempty"`
	OperationSpecs    *OperationSpecs   `json:"operationSpecs,omitempty"`
}

type DeviceProfile struct {
	DeviceType   string                    `json:"deviceType"`
	Capabilities DeviceProfileCapabilities `json:"capabilities"`
}

type DeviceProfileList struct {
	Profiles []DeviceProfile `json:"profiles"`
}

type ListDevicesOptions struct {
	Page     int
	PageSize int
}

type RelayAction string

const (
	RelayActionOn  RelayAction = "ON"
	RelayActionOff RelayAction = "OFF"
	RelayActionJog RelayAction = "JOG"
)

type CommandStatus string

const (
	CommandStatusSent    CommandStatus = "SENT"
	CommandStatusSuccess CommandStatus = "SUCCESS"
	CommandStatusFailed  CommandStatus = "FAILED"
	CommandStatusTimeout CommandStatus = "TIMEOUT"
)

type CommandType string

const (
	CommandTypeRelaySet          CommandType = "RELAY_SET"
	CommandTypeRS485Transceive   CommandType = "RS485_TRANSCEIVE"
	CommandTypeRS485BaudRateSet  CommandType = "RS485_BAUD_RATE_SET"
	CommandTypeRelayJogConfigSet CommandType = "RELAY_JOG_CONFIG_SET"
)

type Command struct {
	ID         string          `json:"id"`
	DeviceID   string          `json:"deviceId"`
	Type       CommandType     `json:"type,omitempty"`
	RelayIndex int             `json:"relayIndex,omitempty"`
	Action     RelayAction     `json:"action,omitempty"`
	Status     CommandStatus   `json:"status,omitempty"`
	Result     json.RawMessage `json:"result,omitempty"`
	CreatedAt  *time.Time      `json:"createdAt,omitempty"`
}

type CommandResult = Command

type RS485TransceiveResult struct {
	RequestHex  string `json:"requestHex"`
	ResponseHex string `json:"responseHex,omitempty"`
}

type RS485BaudRateResult struct {
	BaudRate int `json:"baudRate"`
}

type RelayJogConfigResult struct {
	RelayIndex  int `json:"relayIndex"`
	DurationSec int `json:"durationSec"`
}

type RelaySetOptions struct {
	Index          int
	On             bool
	IdempotencyKey string
}

type RelayJogOptions struct {
	Index          int
	IdempotencyKey string
}

type RelayJogConfigOptions struct {
	Index          int
	DurationSec    int
	IdempotencyKey string
}

type RS485TransceiveOptions struct {
	RequestHex     string
	IdempotencyKey string
}

type RS485BaudRateOptions struct {
	BaudRate       int
	IdempotencyKey string
}
