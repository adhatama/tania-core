package service

import (
	"errors"

	uuid "github.com/satori/go.uuid"

	"github.com/Tanibox/tania-core/src/user/domain"
	"github.com/Tanibox/tania-core/src/user/query"
	"github.com/Tanibox/tania-core/src/user/storage"
)

type OrganizationServiceImpl struct {
	OrganizationReadQuery query.OrganizationReadQuery
}

func (s OrganizationServiceImpl) IsEmailExists(email string) (bool, error) {
	result := <-s.OrganizationReadQuery.FindByEmail(email)
	if result.Error != nil {
		return false, result.Error
	}

	org, ok := result.Result.(storage.OrganizationRead)
	if !ok {
		return false, errors.New("Error type assertion")
	}

	if org.UID == (uuid.UUID{}) {
		return false, nil
	}

	return true, nil
}

func (s OrganizationServiceImpl) FindByName(name string) (domain.Organization, error) {
	result := <-s.OrganizationReadQuery.FindByName(name)

	if result.Error != nil {
		return domain.Organization{}, result.Error
	}

	orgRead, ok := result.Result.(storage.OrganizationRead)
	if !ok {
		return domain.Organization{}, errors.New("Error type assertion")
	}

	if orgRead.UID == (uuid.UUID{}) {
		return domain.Organization{}, errors.New("Email not found")
	}

	org := domain.Organization{
		UID:              orgRead.UID,
		Name:             orgRead.Name,
		Email:            orgRead.Email,
		VerificationCode: orgRead.VerificationCode,
		Status:           orgRead.Status,
		CreatedDate:      orgRead.CreatedDate,
	}

	return org, nil
}
