package domain

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Device struct {
	UID         uuid.UUID
	DeviceID    string
	Name        string
	TopicName   string
	Status      string
	Description string
	CreatedDate time.Time

	// Events
	Version            int
	UncommittedChanges []interface{}
}

type DeviceService interface {
	IsDeviceIDExists(deviceID string) (bool, error)
}

const (
	DeviceStatusMetadataCreated = "METADATA_CREATED"
	DeviceStatusMetadataUpdated = "METADATA_UPDATED"
	DeviceStatusNodeRedCreated  = "NODERED_CREATED"
	DeviceStatusRemoved         = "REMOVED"
)

func (state *Device) TrackChange(event interface{}) {
	state.UncommittedChanges = append(state.UncommittedChanges, event)
	state.Transition(event)
}

func (state *Device) Transition(event interface{}) {
	switch e := event.(type) {
	case DeviceCreated:
		state.UID = e.UID
		state.DeviceID = e.DeviceID
		state.Name = e.Name
		state.TopicName = e.TopicName
		state.Status = e.Status
		state.Description = e.Description
		state.CreatedDate = e.CreatedDate

	case DeviceIDChanged:
		state.UID = e.UID
		state.DeviceID = e.NewDeviceID
		state.TopicName = e.TopicName

	case DeviceNameChanged:
		state.UID = e.UID
		state.Name = e.Name

	case DeviceDescriptionChanged:
		state.UID = e.UID
		state.Description = e.Description

	case DeviceStatusChanged:
		state.UID = e.UID
		state.Status = e.Status

	case DeviceRemoved:
		state.UID = e.UID
		state.Status = e.Status

	}
}

func CreateDevice(deviceService DeviceService, deviceID, name, description string) (*Device, error) {
	isDeviceIDExists, err := deviceService.IsDeviceIDExists(deviceID)
	if err != nil {
		return nil, err
	}

	if isDeviceIDExists {
		return nil, DeviceError{DeviceErrorDeviceIDAlreadyExistsCode}
	}

	if name == "" {
		return nil, DeviceError{DeviceErrorNameEmptyCode}
	}

	// create topic name
	topicName := "topic-" + deviceID

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	device := &Device{
		UID:         uid,
		DeviceID:    deviceID,
		Name:        name,
		Description: description,
		Status:      DeviceStatusMetadataCreated,
	}

	device.TrackChange(DeviceCreated{
		UID:         device.UID,
		DeviceID:    device.DeviceID,
		Name:        device.Name,
		TopicName:   topicName,
		Status:      device.Status,
		Description: device.Description,
		CreatedDate: time.Now(),
	})

	return device, nil
}

func (d *Device) ChangeID(deviceService DeviceService, newDeviceID string) error {
	isDeviceIDExists, err := deviceService.IsDeviceIDExists(newDeviceID)
	if err != nil {
		return err
	}

	if isDeviceIDExists {
		return DeviceError{DeviceErrorDeviceIDAlreadyExistsCode}
	}

	newTopicName := "topic-" + newDeviceID

	d.TrackChange(DeviceIDChanged{
		UID:          d.UID,
		LastDeviceID: d.DeviceID,
		NewDeviceID:  newDeviceID,
		TopicName:    newTopicName,
	})

	return nil
}

func (d *Device) ChangeName(name string) error {
	if name == "" {
		return DeviceError{DeviceErrorNameEmptyCode}
	}

	d.TrackChange(DeviceNameChanged{
		UID:  d.UID,
		Name: name,
	})

	return nil
}

func (d *Device) ChangeDescription(description string) error {
	// validate description

	d.TrackChange(DeviceDescriptionChanged{
		UID:         d.UID,
		Description: description,
	})

	return nil
}

func (d *Device) ChangeStatus(status string) error {
	d.TrackChange(DeviceStatusChanged{
		UID:    d.UID,
		Status: status,
	})

	return nil
}

func (d *Device) Remove() error {
	// validate description

	d.TrackChange(DeviceRemoved{
		UID:    d.UID,
		Status: DeviceStatusRemoved,
	})

	return nil
}
