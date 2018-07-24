package storage

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type UserEvent struct {
	UserUID     uuid.UUID
	Version     int
	CreatedDate time.Time
	Event       interface{}
}

type UserRead struct {
	UID         uuid.UUID `json:"uid"`
	Email       string    `json:"email"`
	Password    []byte    `json:"-"`
	CreatedDate time.Time `json:"created_date"`
	LastUpdated time.Time `json:"last_updated"`
}

type UserAuth struct {
	UserUID      uuid.UUID `json:"uid"`
	AccessToken  string    `json:"access_token"`
	TokenExpires int       `json:"token_expires"`
	CreatedDate  time.Time `json:"created_date"`
	LastUpdated  time.Time `json:"last_updated"`
}

type OrganizationEvent struct {
	OrganizationUID uuid.UUID
	Version         int
	CreatedDate     time.Time
	Event           interface{}
}

type OrganizationRead struct {
	UID              uuid.UUID `json:"uid"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	VerificationCode int       `json:"verification_code"`
	Status           string    `json:"status"`
	CreatedDate      time.Time `json:"created_date"`
}
