package domain

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type UserCreated struct {
	UID             uuid.UUID
	Email           string
	Password        []byte
	OrganizationUID uuid.UUID
	InvitationCode  int
	Role            string
	Status          string
	CreatedDate     time.Time
	LastUpdated     time.Time
}

type PasswordChanged struct {
	UID         uuid.UUID
	NewPassword []byte
	DateChanged time.Time
}

type UserProfileChanged struct {
	UID       uuid.UUID
	Name      string
	Gender    string
	BirthDate time.Time
}

type UserVerified struct {
	UID    uuid.UUID
	Email  string
	Status string
}

type ResetPasswordRequested struct {
	UID               uuid.UUID
	Email             string
	ResetPasswordCode int
}

type InitialUserProfileSet struct {
	UID         uuid.UUID
	Name        string
	Gender      string
	BirthDate   time.Time
	Password    []byte
	Status      string
	DateChanged time.Time
}
