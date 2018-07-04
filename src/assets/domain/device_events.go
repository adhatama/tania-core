package domain

import "time"

type DeviceCreated struct {
	DeviceID    string
	Name        string
	TopicName   string
	Status      string
	CreatedDate time.Time
}
