package service

import (
	"github.com/Tanibox/tania-core/src/assets/domain"
	"github.com/Tanibox/tania-core/src/assets/query"
	"github.com/Tanibox/tania-core/src/assets/storage"
	uuid "github.com/satori/go.uuid"
)

type DeviceService struct {
	DeviceReadQuery query.DeviceReadQuery
}

func (s DeviceService) FindByID(deviceUID uuid.UUID) (domain.DeviceServiceResult, error) {
	result := <-s.DeviceReadQuery.FindByID(deviceUID)

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
			UID:         device.UID,
			DeviceID:    device.DeviceID,
			Name:        device.Name,
			TopicName:   device.TopicName,
			Status:      device.Status,
			Description: device.Description,
			CreatedDate: device.CreatedDate,
		},
	}, nil
}
