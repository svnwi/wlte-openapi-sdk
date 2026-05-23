package wlteopenapi

import "context"

type ProfilesService struct {
	client *Client
}

func (s *ProfilesService) List(ctx context.Context) (DeviceProfileList, error) {
	var result DeviceProfileList
	err := s.client.request(ctx, "GET", "/wlte/v1/device-profiles", nil, nil, nil, &result)
	return result, err
}
