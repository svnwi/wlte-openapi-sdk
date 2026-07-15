package wlteopenapi

import (
	"context"
	"fmt"
)

type RelaysService struct {
	client *Client
}

func (s *RelaysService) Set(ctx context.Context, deviceID string, options RelaySetOptions) (CommandExecution, error) {
	return s.Control(ctx, deviceID, RelayCommandOptions{
		Relays:         []RelayCommand{{Index: options.Index, Action: relayActionForBool(options.On)}},
		IdempotencyKey: options.IdempotencyKey,
	})
}

func (s *RelaysService) Control(ctx context.Context, deviceID string, options RelayCommandOptions) (CommandExecution, error) {
	var result CommandExecution
	err := s.client.request(
		ctx,
		"POST",
		fmt.Sprintf("/wlte/v1/devices/%s/relays/commands", urlEscape(deviceID)),
		nil,
		map[string]string{"Idempotency-Key": options.IdempotencyKey},
		map[string][]RelayCommand{"relays": options.Relays},
		&result,
	)
	return result, err
}

func (s *RelaysService) Jog(ctx context.Context, deviceID string, options RelayJogOptions) (CommandExecution, error) {
	return s.Control(ctx, deviceID, RelayCommandOptions{
		Relays:         []RelayCommand{{Index: options.Index, Action: RelayActionJog}},
		IdempotencyKey: options.IdempotencyKey,
	})
}

func (s *RelaysService) SetJogConfig(ctx context.Context, deviceID string, options RelayJogConfigOptions) (CommandExecution, error) {
	var result CommandExecution
	err := s.client.request(
		ctx,
		"PUT",
		fmt.Sprintf("/wlte/v1/devices/%s/relays/%d/jog-config", urlEscape(deviceID), options.Index),
		nil,
		map[string]string{"Idempotency-Key": options.IdempotencyKey},
		map[string]int{"durationSec": options.DurationSec},
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
