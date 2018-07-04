package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type DeviceServiceMock struct {
	mock.Mock
}

func (m DeviceServiceMock) FindByID(deviceID string) (DeviceServiceResult, error) {
	args := m.Called(deviceID)
	return args.Get(0).(DeviceServiceResult), nil
}

func TestCreateDevice(t *testing.T) {
	// Given
	deviceID := "my-device-id"
	deviceName := "My Device"

	deviceServiceMock := new(DeviceServiceMock)
	deviceServiceMock.On("FindByID", deviceID).Return(DeviceServiceResult{
		Device: &Device{DeviceID: deviceID},
	})

	// When
	device, err := CreateDevice(deviceServiceMock, deviceID, deviceName)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, deviceName, device.Name)

	event, ok := device.UncommittedChanges[0].(DeviceCreated)
	assert.True(t, ok)
	assert.Equal(t, device.DeviceID, event.DeviceID)
}
