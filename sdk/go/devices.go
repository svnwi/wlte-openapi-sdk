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

func (s *DevicesService) GetConfig(ctx context.Context, deviceID string) (DeviceConfig, error) {
	var result DeviceConfig
	err := s.client.request(ctx, "GET", "/wlte/v1/devices/"+urlEscape(deviceID)+"/config", nil, nil, nil, &result)
	return result, err
}

func (s *DevicesService) Add(ctx context.Context, options AddDeviceOptions) (AddDeviceResult, error) {
	var result AddDeviceResult
	err := s.client.request(ctx, "POST", "/wlte/v1/devices", nil, nil, map[string]string{
		"deviceId": options.DeviceID,
		"password": options.Password,
		"name":     options.Name,
	}, &result)
	return result, err
}

func (s *DevicesService) Remove(ctx context.Context, deviceID string) (RemoveDeviceResult, error) {
	var result RemoveDeviceResult
	err := s.client.request(ctx, "DELETE", "/wlte/v1/devices/"+urlEscape(deviceID), nil, nil, nil, &result)
	return result, err
}

func (s *DevicesService) ModifyPassword(ctx context.Context, deviceID string, options ModifyDevicePasswordOptions) (ModifyDevicePasswordResult, error) {
	var result ModifyDevicePasswordResult
	err := s.client.request(ctx, "PUT", "/wlte/v1/devices/"+urlEscape(deviceID)+"/password", nil, nil, map[string]string{
		"oldPassword": options.OldPassword,
		"newPassword": options.NewPassword,
	}, &result)
	return result, err
}
