package service

import (
	"github.com/Tanibox/tania-core/src/assets/domain"
	"github.com/Tanibox/tania-core/src/assets/query"
	"github.com/Tanibox/tania-core/src/assets/storage"
)

type DeviceService struct {
	DeviceReadQuery query.DeviceReadQuery
}

func (s DeviceService) IsDeviceIDExists(deviceID string) (bool, error) {
	result := <-s.DeviceReadQuery.FindByDeviceID(deviceID)
	if result.Error != nil {
		return false, result.Error
	}

	device, ok := result.Result.(storage.DeviceRead)
	if !ok {
		return false, domain.DeviceError{Code: domain.DeviceErrorDeviceNotFoundCode}
	}

	if device == (storage.DeviceRead{}) {
		return false, nil
	}

	return true, nil
}
