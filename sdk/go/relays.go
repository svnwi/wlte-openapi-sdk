package wlteopenapi

import (
	"context"
	"fmt"
)

type RelaysService struct {
	client *Client
}

func (s *RelaysService) Set(ctx context.Context, deviceID string, options RelaySetOptions) (Command, error) {
	var result Command
	err := s.client.request(
		ctx,
		"POST",
		fmt.Sprintf("/wlte/v1/devices/%s/relays/%d/commands", urlEscape(deviceID), options.Index),
		nil,
		map[string]string{"Idempotency-Key": options.IdempotencyKey},
		map[string]RelayAction{"action": relayActionForBool(options.On)},
		&result,
	)
	return result, err
}

func (s *RelaysService) Jog(ctx context.Context, deviceID string, options RelayJogOptions) (Command, error) {
	var result Command
	err := s.client.request(
		ctx,
		"POST",
		fmt.Sprintf("/wlte/v1/devices/%s/relays/%d/commands", urlEscape(deviceID), options.Index),
		nil,
		map[string]string{"Idempotency-Key": options.IdempotencyKey},
		map[string]RelayAction{"action": RelayActionJog},
		&result,
	)
	return result, err
}

func relayActionForBool(on bool) RelayAction {
	if on {
		return RelayActionOn
	}
	return RelayActionOff
}
