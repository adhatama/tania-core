package domain

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type DeviceCreated struct {
	UID         uuid.UUID
	DeviceID    string
	Name        string
	TopicName   string
	Status      string
	Description string
	CreatedDate time.Time
}

type DeviceIDChanged struct {
	UID          uuid.UUID
	LastDeviceID string
	NewDeviceID  string
	TopicName    string
}

type DeviceNameChanged struct {
	UID  uuid.UUID
	Name string
}

type DeviceDescriptionChanged struct {
	UID         uuid.UUID
	Description string
}

type DeviceStatusChanged struct {
	UID    uuid.UUID
	Status string
}

type DeviceRemoved struct {
	UID    uuid.UUID
	Status string
}
