package service

import (
	"github.com/Tanibox/tania-core/src/assets/domain"
	"github.com/Tanibox/tania-core/src/assets/query"
	"github.com/Tanibox/tania-core/src/assets/storage"
)

type DeviceService struct {
	DeviceReadQuery query.DeviceReadQuery
}

func (s DeviceService) FindByID(deviceID string) (domain.DeviceServiceResult, error) {
	result := <-s.DeviceReadQuery.FindByID(deviceID)

	if result.Error != nil {
		return domain.DeviceServiceResult{}, result.Error
	}

	device, ok := result.Result.(storage.DeviceRead)

	if !ok {
		return domain.DeviceServiceResult{}, domain.ReservoirError{Code: domain.ReservoirErrorFarmNotFound}
	}

	if device == (storage.DeviceRead{}) {
		return domain.DeviceServiceResult{}, domain.ReservoirError{Code: domain.ReservoirErrorFarmNotFound}
	}

	return domain.DeviceServiceResult{
		Device: &domain.Device{
			DeviceID:    device.DeviceID,
			Name:        device.Name,
			TopicName:   device.TopicName,
			Status:      device.Status,
			CreatedDate: device.CreatedDate,
		},
	}, nil
}
