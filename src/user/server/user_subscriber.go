package server

import (
	"errors"

	"github.com/Tanibox/tania-core/src/user/domain"
	"github.com/Tanibox/tania-core/src/user/storage"
	"github.com/labstack/gommon/log"
	uuid "github.com/satori/go.uuid"
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
		userRead.Role = e.Role
		userRead.Status = e.Status
		userRead.InvitationCode = e.InvitationCode
		userRead.OrganizationUID = e.OrganizationUID
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

	case domain.UserVerified:
		queryResult := <-s.UserReadQuery.FindByID(e.UID)
		if queryResult.Error != nil {
			log.Error(queryResult.Error)
		}

		u, ok := queryResult.Result.(storage.UserRead)
		if !ok {
			log.Error(errors.New("Internal server error. Error type assertion"))
		}

		userRead = &u

		userRead.Status = e.Status

	case domain.ResetPasswordRequested:
		queryResult := <-s.UserReadQuery.FindByID(e.UID)
		if queryResult.Error != nil {
			log.Error(queryResult.Error)
		}

		u, ok := queryResult.Result.(storage.UserRead)
		if !ok {
			log.Error(errors.New("Internal server error. Error type assertion"))
		}

		userRead = &u

		userRead.ResetPasswordCode = e.ResetPasswordCode

	}

	err := <-s.UserReadRepo.Save(userRead)
	if err != nil {
		log.Error(err)
	}

	return nil
}

func (s *UserServer) SaveToAuthModel(event interface{}) error {
	userAuth := &storage.UserAuth{}

	switch e := event.(type) {
	case domain.UserCreated:
		userAuth.UserUID = e.UID
		userAuth.CreatedDate = e.CreatedDate
		userAuth.LastUpdated = e.LastUpdated

		// Generate access token here
		// We use uuid method temporarily until we find better method
		uidAccessToken, err := uuid.NewV4()
		if err != nil {
			log.Error(err)
		}

		userAuth.AccessToken = uidAccessToken.String()
	}

	err := <-s.UserAuthRepo.Save(userAuth)
	if err != nil {
		log.Error(err)
	}

	return nil
}
