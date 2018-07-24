package server

import (
	"errors"

	"github.com/Tanibox/tania-core/src/user/domain"
	"github.com/Tanibox/tania-core/src/user/storage"
	"github.com/labstack/gommon/log"
)

func (s *UserServer) SaveToOrganizationReadModel(event interface{}) error {
	orgRead := &storage.OrganizationRead{}

	switch e := event.(type) {
	case domain.OrganizationCreated:
		orgRead.UID = e.UID
		orgRead.Email = e.Email
		orgRead.Name = e.Name
		orgRead.VerificationCode = e.VerificationCode
		orgRead.Status = e.Status
		orgRead.CreatedDate = e.CreatedDate

	case domain.OrganizationVerified:
		queryResult := <-s.OrganizationReadQuery.FindByID(e.UID)
		if queryResult.Error != nil {
			log.Error(queryResult.Error)
		}

		org, ok := queryResult.Result.(storage.OrganizationRead)
		if !ok {
			log.Error(errors.New("Internal server error. Error type assertion"))
		}

		orgRead = &org

		orgRead.Status = e.Status

	}

	err := <-s.OrganizationReadRepo.Save(orgRead)
	if err != nil {
		log.Error(err)
	}

	return nil
}

func (s *UserServer) SaveToUserReadModel(event interface{}) error {
	userRead := &storage.UserRead{}

	switch e := event.(type) {
	case domain.UserCreated:
		userRead.UID = e.UID
		userRead.Email = e.Email
		userRead.Password = e.Password
		userRead.CreatedDate = e.CreatedDate
		userRead.LastUpdated = e.LastUpdated

	case domain.PasswordChanged:
		queryResult := <-s.UserReadQuery.FindByID(e.UID)
		if queryResult.Error != nil {
			log.Error(queryResult.Error)
		}

		u, ok := queryResult.Result.(storage.UserRead)
		if !ok {
			log.Error(errors.New("Internal server error. Error type assertion"))
		}

		userRead = &u

		userRead.Password = e.NewPassword
		userRead.LastUpdated = e.DateChanged

	}

	err := <-s.UserReadRepo.Save(userRead)
	if err != nil {
		log.Error(err)
	}

	return nil
}
