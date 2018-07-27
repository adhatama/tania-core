package server

import (
	"errors"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/Tanibox/tania-core/config"
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

	case domain.OrganizationProfileChanged:
		queryResult := <-s.OrganizationReadQuery.FindByID(e.UID)
		if queryResult.Error != nil {
			log.Error(queryResult.Error)
		}

		org, ok := queryResult.Result.(storage.OrganizationRead)
		if !ok {
			log.Error(errors.New("Internal server error. Error type assertion"))
		}

		orgRead = &org

		orgRead.Name = e.Name
		orgRead.Type = e.Type
		orgRead.TotalMember = e.TotalMember
		orgRead.Province = e.Province
		orgRead.City = e.City

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
	case domain.UserAdminCreated:
		userRead.UID = e.UID
		userRead.Email = e.Email
		userRead.Role = e.Role
		userRead.Status = e.Status
		userRead.OrganizationUID = e.OrganizationUID
		userRead.CreatedDate = e.CreatedDate
		userRead.LastUpdated = e.LastUpdated

	case domain.UserInvited:
		userRead.UID = e.UID
		userRead.Email = e.Email
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

	case domain.InitialUserProfileSet:
		queryResult := <-s.UserReadQuery.FindByID(e.UID)
		if queryResult.Error != nil {
			log.Error(queryResult.Error)
		}

		u, ok := queryResult.Result.(storage.UserRead)
		if !ok {
			log.Error(errors.New("Internal server error. Error type assertion"))
		}

		userRead = &u

		userRead.Name = &e.Name
		userRead.Gender = &e.Gender
		userRead.BirthDate = &e.BirthDate
		userRead.Status = e.Status
		userRead.LastUpdated = e.DateChanged
		userRead.Password = e.Password

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
	case domain.InitialUserProfileSet:
		userAuth.UserUID = e.UID
		userAuth.CreatedDate = e.DateChanged
		userAuth.LastUpdated = e.DateChanged

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

func (s *UserServer) SendEmailSubscriber(event interface{}) error {
	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		*config.Config.MailUsername,
		*config.Config.MailPassword,
		*config.Config.MailHost,
	)

	recipients := []string{}
	code := ""
	subject := ""
	switch e := event.(type) {
	case domain.UserInvited:
		subject = "Tania Kode Verifikasi untuk Pendaftaran Pengguna Baru"
		recipients = append(recipients, e.Email)
		code = strconv.Itoa(e.InvitationCode)
	}

	composedMsg := "From: " + *config.Config.MailSender + "\r\n" +
		"To: " + strings.Join(recipients, ",") + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		"Kode undangan Anda adalah " + code

	err := smtp.SendMail(
		*config.Config.MailHost+":"+strconv.Itoa(*config.Config.MailPort),
		auth,
		*config.Config.MailSender,
		recipients,
		[]byte(composedMsg),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
