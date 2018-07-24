package domain

import (
	"errors"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Organization struct {
	UID              uuid.UUID
	Name             string
	Email            string
	VerificationCode int
	Status           string
	CreatedDate      time.Time

	// Events
	Version            int
	UncommittedChanges []interface{}
}

type OrganizationService interface {
	IsEmailExists(email string) (bool, error)
	FindByName(name string) (Organization, error)
}

const (
	OrganizationStatusPendingConfirmation = "PENDING_CONFIRMATION"
	OrganizationStatusConfirmed           = "CONFIRMED"
)

func (state *Organization) TrackChange(event interface{}) {
	state.UncommittedChanges = append(state.UncommittedChanges, event)
	state.Transition(event)
}

func (state *Organization) Transition(event interface{}) {
	switch e := event.(type) {
	case OrganizationCreated:
		state.UID = e.UID
		state.Name = e.Name
		state.Email = e.Email
		state.VerificationCode = e.VerificationCode
		state.Status = e.Status
		state.CreatedDate = e.CreatedDate

	case OrganizationNameChanged:
		state.Name = e.Name

	case OrganizationVerified:
		state.Status = e.Status

	}
}

func CreateOrganization(orgService OrganizationService, name, email string) (*Organization, error) {
	if email == "" {
		return nil, errors.New("Email cannot be empty")
	}

	isExists, err := orgService.IsEmailExists(email)
	if err != nil {
		return nil, err
	}

	if isExists {
		return nil, errors.New("Email already exists")
	}

	if name == "" {
		return nil, errors.New("Name cannot be empty")
	}

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	// Generate 6 digit random number
	rand.Seed(time.Now().UnixNano())
	code := 100000 + rand.Intn(900000)

	org := &Organization{
		UID:              uid,
		Name:             name,
		Email:            email,
		VerificationCode: code,
		Status:           OrganizationStatusPendingConfirmation,
		CreatedDate:      time.Now(),
	}

	org.TrackChange(OrganizationCreated{
		UID:              org.UID,
		Name:             org.Name,
		Email:            org.Email,
		VerificationCode: org.VerificationCode,
		Status:           org.Status,
		CreatedDate:      org.CreatedDate,
	})

	return org, nil
}

func (o *Organization) ChangeName(name string) error {
	if name == "" {
		return errors.New("Name cannot be empty")
	}

	o.TrackChange(OrganizationNameChanged{
		UID:  o.UID,
		Name: name,
	})

	return nil
}

func (o *Organization) Verify() error {
	if o.Status == OrganizationStatusConfirmed {
		return errors.New("Status already confirmed")
	}

	o.TrackChange(OrganizationVerified{
		UID:    o.UID,
		Email:  o.Email,
		Status: OrganizationStatusConfirmed,
	})

	return nil
}
