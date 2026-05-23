package wlteopenapi

import (
	"context"
	"strconv"
)

type DevicesService struct {
	client *Client
}

func (s *DevicesService) List(ctx context.Context, options ListDevicesOptions) (DeviceList, error) {
	query := map[string]string{}
	if options.Page > 0 {
		query["page"] = strconv.Itoa(options.Page)
	}
	if options.PageSize > 0 {
		query["pageSize"] = strconv.Itoa(options.PageSize)
	}

	var result DeviceList
	err := s.client.request(ctx, "GET", "/wlte/v1/devices", query, nil, nil, &result)
	return result, err
}

func (s *DevicesService) Get(ctx context.Context, deviceID string) (Device, error) {
	var result Device
	err := s.client.request(ctx, "GET", "/wlte/v1/devices/"+urlEscape(deviceID), nil, nil, nil, &result)
	return result, err
}
