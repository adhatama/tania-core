package domain

import "time"

type Device struct {
	DeviceID    string
	Name        string
	TopicName   string
	Status      string
	CreatedDate time.Time

	// Events
	Version            int
	UncommittedChanges []interface{}
}

type DeviceService interface {
	FindByID(deviceID string) (DeviceServiceResult, error)
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
		state.DeviceID = e.DeviceID
		state.Name = e.Name
		state.TopicName = e.TopicName
		state.Status = e.Status
		state.CreatedDate = e.CreatedDate

	}
}

func CreateDevice(deviceService DeviceService, deviceID, name string) (*Device, error) {
	// validate device ID, name

	// create topic name
	topicName := "topic-" + deviceID

	device := &Device{
		DeviceID: deviceID,
		Name:     name,
		Status:   DeviceMetadataCreated,
	}

	device.TrackChange(DeviceCreated{
		DeviceID:    device.DeviceID,
		Name:        device.Name,
		TopicName:   topicName,
		Status:      device.Status,
		CreatedDate: time.Now(),
	})

	return device, nil
}

func (d *Device) ChangeID(newDeviceID string) error {
	// validate new device ID

	// create topic name

	d.DeviceID = newDeviceID

	return nil
}

func (d *Device) ChangeName(name string) error {
	// validate name

	d.Name = name

	return nil
}

func (d *Device) ChangeStatus(status string) error {
	// validate status

	d.Status = status

	return nil
}
