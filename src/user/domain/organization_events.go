package domain

import (
	"time"

	"github.com/satori/go.uuid"
)

type OrganizationCreated struct {
	UID              uuid.UUID
	Email            string
	VerificationCode int
	Status           string
	CreatedDate      time.Time
}

type OrganizationNameChanged struct {
	UID  uuid.UUID
	Name string
}

type OrganizationVerified struct {
	UID    uuid.UUID
	Email  string
	Status string
}
