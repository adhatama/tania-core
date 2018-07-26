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
	Type             string
	TotalMember      string
	Province         string
	City             string
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
	OrganizationStatusCompleted           = "COMPLETED"
)

func (state *Organization) TrackChange(event interface{}) {
	state.UncommittedChanges = append(state.UncommittedChanges, event)
	state.Transition(event)
}

func (state *Organization) Transition(event interface{}) {
	switch e := event.(type) {
	case OrganizationCreated:
		state.UID = e.UID
		state.Email = e.Email
		state.VerificationCode = e.VerificationCode
		state.Status = e.Status
		state.CreatedDate = e.CreatedDate

	case OrganizationProfileChanged:
		state.Name = e.Name
		state.Type = e.Type
		state.TotalMember = e.TotalMember
		state.Province = e.Province
		state.City = e.City

	case OrganizationVerified:
		state.Status = e.Status

	}
}

func CreateOrganization(orgService OrganizationService, email string) (*Organization, error) {
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

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	// Generate 6 digit random number
	rand.Seed(time.Now().UnixNano())
	code := 100000 + rand.Intn(900000)

	org := &Organization{
		UID:              uid,
		Email:            email,
		VerificationCode: code,
		Status:           OrganizationStatusPendingConfirmation,
		CreatedDate:      time.Now(),
	}

	org.TrackChange(OrganizationCreated{
		UID:              org.UID,
		Email:            org.Email,
		VerificationCode: org.VerificationCode,
		Status:           org.Status,
		CreatedDate:      org.CreatedDate,
	})

	return org, nil
}

func (o *Organization) ChangeProfile(orgService OrganizationService, name, orgType, totalMember, province, city string) error {
	if name == "" {
		return errors.New("Name cannot be empty")
	}

	org, err := orgService.FindByName(name)
	if err != nil {
		return err
	}

	if org.UID != (uuid.UUID{}) && org.UID != o.UID {
		return errors.New("Organization name is already used")
	}

	if orgType == "" {
		return errors.New("Organization type cannot be empty")
	}

	if totalMember == "" {
		return errors.New("Total member cannot be empty")
	}

	if province == "" {
		return errors.New("Province cannot be empty")
	}

	if city == "" {
		return errors.New("City cannot be empty")
	}

	o.TrackChange(OrganizationProfileChanged{
		UID:         o.UID,
		Name:        name,
		Type:        orgType,
		TotalMember: totalMember,
		Province:    province,
		City:        city,
		Status:      OrganizationStatusCompleted,
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
