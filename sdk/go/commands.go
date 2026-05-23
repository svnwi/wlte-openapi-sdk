package wlteopenapi

import "context"

type CommandsService struct {
	client *Client
}

func (s *CommandsService) GetResult(ctx context.Context, commandID string) (CommandResult, error) {
	var result CommandResult
	err := s.client.request(ctx, "GET", "/wlte/v1/commands/"+urlEscape(commandID), nil, nil, nil, &result)
	return result, err
}
