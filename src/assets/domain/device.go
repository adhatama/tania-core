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
	FindByID(deviceUID uuid.UUID) (DeviceServiceResult, error)
}

type DeviceServiceResult struct {
	*Device
}

const (
	DeviceMetadataCreated = "METADATA_CREATED"
	DeviceMetadataUpdated = "METADATA_UPDATED"
	DeviceNodeRedCreated  = "NODERED_CREATED"
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
		state.DeviceID = e.DeviceID
		state.TopicName = e.TopicName

	case DeviceNameChanged:
		state.UID = e.UID
		state.Name = e.Name

	case DeviceDescriptionChanged:
		state.UID = e.UID
		state.Description = e.Description

	}
}

func CreateDevice(deviceService DeviceService, deviceID, name, description string) (*Device, error) {
	// validate device ID, name

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
		Status:      DeviceMetadataCreated,
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

func (d *Device) ChangeID(newDeviceID string) error {
	// validate new device ID

	// create topic name

	d.TrackChange(DeviceIDChanged{
		UID:       d.UID,
		DeviceID:  newDeviceID,
		TopicName: d.TopicName,
	})

	return nil
}

func (d *Device) ChangeName(name string) error {
	// validate name

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
