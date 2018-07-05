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
