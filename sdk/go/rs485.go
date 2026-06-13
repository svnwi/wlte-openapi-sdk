package wlteopenapi

import "context"

type RS485Service struct {
	client *Client
}

func (s *RS485Service) Transceive(ctx context.Context, deviceID string, options RS485TransceiveOptions) (Command, error) {
	var result Command
	err := s.client.request(
		ctx,
		"POST",
		"/wlte/v1/devices/"+urlEscape(deviceID)+"/rs485/transceive",
		nil,
		map[string]string{"Idempotency-Key": options.IdempotencyKey},
		map[string]string{"requestHex": options.RequestHex},
		&result,
	)
	return result, err
}

func (s *RS485Service) SetBaudRate(ctx context.Context, deviceID string, options RS485BaudRateOptions) (Command, error) {
	var result Command
	err := s.client.request(
		ctx,
		"PUT",
		"/wlte/v1/devices/"+urlEscape(deviceID)+"/rs485/baud-rate",
		nil,
		map[string]string{"Idempotency-Key": options.IdempotencyKey},
		map[string]int{"baudRate": options.BaudRate},
		&result,
	)
	return result, err
}
