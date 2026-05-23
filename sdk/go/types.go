package wlteopenapi

import "time"

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

type Command struct {
	ID         string        `json:"id"`
	DeviceID   string        `json:"deviceId"`
	RelayIndex int           `json:"relayIndex"`
	Action     RelayAction   `json:"action"`
	Status     CommandStatus `json:"status,omitempty"`
	CreatedAt  *time.Time    `json:"createdAt,omitempty"`
}

type CommandResult = Command

type RelaySetOptions struct {
	Index          int
	On             bool
	IdempotencyKey string
}

type RelayJogOptions struct {
	Index          int
	IdempotencyKey string
}
