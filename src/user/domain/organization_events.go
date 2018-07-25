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

type OrganizationProfileChanged struct {
	UID         uuid.UUID
	Name        string
	Type        string
	TotalMember string
	Province    string
	City        string
}

type OrganizationVerified struct {
	UID    uuid.UUID
	Email  string
	Status string
}
