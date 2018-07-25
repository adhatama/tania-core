package server

import (
	"database/sql"
	"errors"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/Tanibox/tania-core/config"
	"github.com/Tanibox/tania-core/src/eventbus"
	"github.com/Tanibox/tania-core/src/helper/structhelper"
	"github.com/Tanibox/tania-core/src/user/domain"
	"github.com/Tanibox/tania-core/src/user/domain/service"
	"github.com/Tanibox/tania-core/src/user/query"
	queryMysql "github.com/Tanibox/tania-core/src/user/query/mysql"
	querySqlite "github.com/Tanibox/tania-core/src/user/query/sqlite"
	"github.com/Tanibox/tania-core/src/user/repository"
	repoMysql "github.com/Tanibox/tania-core/src/user/repository/mysql"
	repoSqlite "github.com/Tanibox/tania-core/src/user/repository/sqlite"
	"github.com/Tanibox/tania-core/src/user/storage"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	uuid "github.com/satori/go.uuid"
)

// AuthServer ties the routes and handlers with injected dependencies
type AuthServer struct {
	UserEventRepo  repository.UserEventRepository
	UserReadRepo   repository.UserReadRepository
	UserEventQuery query.UserEventQuery
	UserReadQuery  query.UserReadQuery
	UserAuthRepo   repository.UserAuthRepository
	UserAuthQuery  query.UserAuthQuery
	UserService    domain.UserService

	OrganizationEventRepo  repository.OrganizationEventRepository
	OrganizationReadRepo   repository.OrganizationReadRepository
	OrganizationEventQuery query.OrganizationEventQuery
	OrganizationReadQuery  query.OrganizationReadQuery
	OrganizationService    domain.OrganizationService

	EventBus eventbus.TaniaEventBus
}

// NewAuthServer initializes AuthServer's dependencies and create new AuthServer struct
func NewAuthServer(
	db *sql.DB,
	eventBus eventbus.TaniaEventBus,
) (*AuthServer, error) {

	authServer := &AuthServer{
		EventBus: eventBus,
	}

	switch *config.Config.TaniaPersistenceEngine {
	case config.DB_SQLITE:
		authServer.UserEventRepo = repoSqlite.NewUserEventRepositorySqlite(db)
		authServer.UserReadRepo = repoSqlite.NewUserReadRepositorySqlite(db)
		authServer.UserEventQuery = querySqlite.NewUserEventQuerySqlite(db)
		authServer.UserReadQuery = querySqlite.NewUserReadQuerySqlite(db)

		authServer.UserAuthRepo = repoSqlite.NewUserAuthRepositorySqlite(db)
		authServer.UserAuthQuery = querySqlite.NewUserAuthQuerySqlite(db)

		authServer.UserService = service.UserServiceImpl{UserReadQuery: authServer.UserReadQuery}

	case config.DB_MYSQL:
		authServer.UserEventRepo = repoMysql.NewUserEventRepositoryMysql(db)
		authServer.UserReadRepo = repoMysql.NewUserReadRepositoryMysql(db)
		authServer.UserEventQuery = queryMysql.NewUserEventQueryMysql(db)
		authServer.UserReadQuery = queryMysql.NewUserReadQueryMysql(db)

		authServer.OrganizationEventRepo = repoMysql.NewOrganizationEventRepositoryMysql(db)
		authServer.OrganizationReadRepo = repoMysql.NewOrganizationReadRepositoryMysql(db)
		authServer.OrganizationEventQuery = queryMysql.NewOrganizationEventQueryMysql(db)
		authServer.OrganizationReadQuery = queryMysql.NewOrganizationReadQueryMysql(db)

		authServer.UserAuthRepo = repoMysql.NewUserAuthRepositoryMysql(db)
		authServer.UserAuthQuery = queryMysql.NewUserAuthQueryMysql(db)

		authServer.UserService = service.UserServiceImpl{
			UserReadQuery: authServer.UserReadQuery,
		}
		authServer.OrganizationService = service.OrganizationServiceImpl{
			OrganizationReadQuery: authServer.OrganizationReadQuery,
		}

	}

	authServer.InitSubscriber()

	return authServer, nil
}

// InitSubscriber defines the mapping of which event this domain listen with their handler
func (s *AuthServer) InitSubscriber() {
	s.EventBus.Subscribe("OrganizationCreated", s.SendEmailSubscriber)
}

// Mount defines the AuthServer's endpoints with its handlers
func (s *AuthServer) Mount(g *echo.Group) {
	g.POST("authorize", s.Authorize)
	g.POST("register/organization", s.RegisterOrganization)
	g.POST("organization/verification", s.VerifyOrganization)
	g.POST("register/user", s.RegisterUser)
	g.POST("user/verification", s.VerifyUser)
	g.POST("forgot_password", s.ForgotPassword)
	g.POST("reset_password", s.ResetPassword)
}

func (s *AuthServer) Authorize(c echo.Context) error {
	responseType := "token"
	redirectURI := config.Config.RedirectURI
	clientID := *config.Config.ClientID

	reqEmail := c.FormValue("email")
	reqPassword := c.FormValue("password")
	reqClientID := c.FormValue("client_id")
	reqResponseType := c.FormValue("response_type")
	reqRedirectURI := c.FormValue("redirect_uri")
	reqState := c.FormValue("state")

	queryResult := <-s.UserReadQuery.FindByEmailAndPassword(reqEmail, reqPassword)
	if queryResult.Error != nil {
		return Error(c, queryResult.Error)
	}

	userRead, ok := queryResult.Result.(storage.UserRead)
	if !ok {
		return Error(c, errors.New("Error type assertion"))
	}

	queryResult = <-s.UserAuthQuery.FindByUserID(userRead.UID)
	if queryResult.Error != nil {
		return Error(c, queryResult.Error)
	}

	userAuth, ok := queryResult.Result.(storage.UserAuth)
	if !ok {
		return Error(c, errors.New("Error type assertion"))
	}

	if userRead.UID == (uuid.UUID{}) {
		return Error(c, NewRequestValidationError(INVALID, "username"))
	}

	if reqClientID != clientID {
		return Error(c, NewRequestValidationError(INVALID, "client_id"))
	}

	selectedRedirectURI := ""
	for _, v := range redirectURI {
		if reqRedirectURI == *v {
			selectedRedirectURI = *v
			break
		}
	}

	if selectedRedirectURI == "" {
		return Error(c, NewRequestValidationError(INVALID, "redirect_uri"))
	}

	if reqResponseType != responseType {
		return Error(c, NewRequestValidationError(INVALID, "response_type"))
	}

	// Generate access token here
	// We use uuid method temporarily until we find better method
	uidAccessToken, err := uuid.NewV4()
	if err != nil {
		return Error(c, err)
	}

	accessToken := uidAccessToken.String()

	// We don't expire token because it's complicating things
	// Also Google recommend it. https://developers.google.com/actions/identity/oauth2-implicit-flow
	expiresIn := 0

	userAuth.AccessToken = accessToken
	userAuth.TokenExpires = expiresIn

	err = <-s.UserAuthRepo.Save(&userAuth)
	if err != nil {
		return Error(c, err)
	}

	selectedRedirectURI += "?" + "access_token=" + accessToken + "&state=" + reqState + "&expires_in=" + strconv.Itoa(expiresIn)

	c.Response().Header().Set(echo.HeaderAuthorization, "Bearer "+accessToken)

	return c.Redirect(302, selectedRedirectURI)
}

func (s *AuthServer) RegisterOrganization(c echo.Context) error {
	email := c.FormValue("email")

	if email == "" {
		return Error(c, errors.New("Email is required"))
	}

	org, err := domain.CreateOrganization(s.OrganizationService, email)
	if err != nil {
		return Error(c, err)
	}

	err = <-s.OrganizationEventRepo.Save(org.UID, org.Version, org.UncommittedChanges)
	if err != nil {
		return Error(c, err)
	}

	s.publishUncommittedEvents(org)

	data := make(map[string]interface{})
	data["data"] = org

	return c.JSON(http.StatusOK, data)
}

func (s *AuthServer) VerifyOrganization(c echo.Context) error {
	// Validate
	organizationUID, err := uuid.FromString(c.FormValue("organization_id"))
	if err != nil {
		return Error(c, err)
	}

	verificationCode := c.FormValue("verification_code")
	code, err := strconv.Atoi(verificationCode)
	if err != nil {
		return Error(c, err)
	}

	if verificationCode == "" {
		return Error(c, errors.New("Verification code is required"))
	}

	queryResult := <-s.OrganizationReadQuery.FindByIDAndVerificationCode(organizationUID, code)
	if queryResult.Error != nil {
		return Error(c, queryResult.Error)
	}

	orgRead, ok := queryResult.Result.(storage.OrganizationRead)
	if !ok {
		return Error(c, errors.New("Error type assertion"))
	}

	if orgRead.UID == (uuid.UUID{}) {
		return Error(c, errors.New("Verification code not found"))
	}

	if orgRead.Status == domain.OrganizationStatusConfirmed {
		return Error(c, errors.New("Organization is already confirmed"))
	}

	// Process
	eventQueryResult := <-s.OrganizationEventQuery.FindAllByID(orgRead.UID)
	if eventQueryResult.Error != nil {
		return Error(c, eventQueryResult.Error)
	}

	events := eventQueryResult.Result.([]storage.OrganizationEvent)
	org := repository.NewOrganizationFromHistory(events)

	err = org.Verify()
	if err != nil {
		return Error(c, err)
	}

	err = <-s.OrganizationEventRepo.Save(org.UID, org.Version, org.UncommittedChanges)
	if err != nil {
		return Error(c, err)
	}

	s.publishUncommittedEvents(org)

	data := make(map[string]interface{})
	data["data"] = organizationUID

	return c.JSON(http.StatusOK, data)
}

func (s *AuthServer) RegisterUser(c echo.Context) error {
	organizationUID, err := uuid.FromString(c.FormValue("organization_id"))
	if err != nil {
		return Error(c, err)
	}

	email := c.FormValue("email")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")
	role := c.FormValue("role")

	if password != confirmPassword {
		return Error(c, errors.New("Confirm password didn't match"))
	}

	user, err := domain.CreateUser(s.UserService, organizationUID, email, password, role)
	if err != nil {
		return Error(c, err)
	}

	err = <-s.UserEventRepo.Save(user.UID, user.Version, user.UncommittedChanges)
	if err != nil {
		return Error(c, err)
	}

	s.publishUncommittedEvents(user)

	data := make(map[string]storage.UserRead)
	data["data"] = MapToUserRead(user)

	return c.JSON(http.StatusOK, data)
}

func (s *AuthServer) VerifyUser(c echo.Context) error {
	// Validate
	organizationUID, err := uuid.FromString(c.FormValue("organization_id"))
	if err != nil {
		return Error(c, err)
	}

	invitationCode := c.FormValue("invitation_code")
	code, err := strconv.Atoi(invitationCode)
	if err != nil {
		return Error(c, err)
	}

	if invitationCode == "" {
		return Error(c, errors.New("Invitation code is required"))
	}

	queryResult := <-s.UserReadQuery.FindByOrganizationIDAndInvitationCode(organizationUID, code)
	if queryResult.Error != nil {
		return Error(c, queryResult.Error)
	}

	userRead, ok := queryResult.Result.(storage.UserRead)
	if !ok {
		return Error(c, errors.New("Error type assertion"))
	}

	if userRead.UID == (uuid.UUID{}) {
		return Error(c, errors.New("Invitation code not found"))
	}

	if userRead.Status == domain.UserStatusConfirmed {
		return Error(c, errors.New("User is already confirmed"))
	}

	// Process
	eventQueryResult := <-s.UserEventQuery.FindAllByID(userRead.UID)
	if eventQueryResult.Error != nil {
		return Error(c, eventQueryResult.Error)
	}

	events := eventQueryResult.Result.([]storage.UserEvent)
	user := repository.NewUserFromHistory(events)

	err = user.VerifyInvitation()
	if err != nil {
		return Error(c, err)
	}

	err = <-s.UserEventRepo.Save(user.UID, user.Version, user.UncommittedChanges)
	if err != nil {
		return Error(c, err)
	}

	s.publishUncommittedEvents(user)

	data := make(map[string]interface{})
	data["data"] = MapToUserRead(user)

	return c.JSON(http.StatusOK, data)
}

func (s *AuthServer) ForgotPassword(c echo.Context) error {
	email := c.FormValue("email")

	queryResult := <-s.UserReadQuery.FindByEmail(email)
	if queryResult.Error != nil {
		return Error(c, queryResult.Error)
	}

	userRead, ok := queryResult.Result.(storage.UserRead)
	if !ok {
		return Error(c, errors.New("Error type assertion"))
	}

	if userRead.UID == (uuid.UUID{}) {
		return Error(c, errors.New("Email has not registered"))
	}

	// Process
	eventQueryResult := <-s.UserEventQuery.FindAllByID(userRead.UID)
	if eventQueryResult.Error != nil {
		return Error(c, eventQueryResult.Error)
	}

	events := eventQueryResult.Result.([]storage.UserEvent)
	user := repository.NewUserFromHistory(events)

	err := user.RequestResetPassword()
	if err != nil {
		return Error(c, err)
	}

	err = <-s.UserEventRepo.Save(user.UID, user.Version, user.UncommittedChanges)
	if err != nil {
		return Error(c, err)
	}

	s.publishUncommittedEvents(user)

	data := make(map[string]interface{})
	data["data"] = MapToUserRead(user)

	return c.JSON(http.StatusOK, data)
}

func (s *AuthServer) ResetPassword(c echo.Context) error {
	email := c.FormValue("email")
	resetPwdCode := c.FormValue("reset_password_code")
	newPassword := c.FormValue("new_password")
	confirmNewPassword := c.FormValue("confirm_new_password")

	if resetPwdCode == "" {
		return Error(c, errors.New("Reset password code is required"))
	}

	code, err := strconv.Atoi(resetPwdCode)
	if err != nil {
		return Error(c, err)
	}

	queryResult := <-s.UserReadQuery.FindByEmailAndResetPasswordCode(email, code)
	if queryResult.Error != nil {
		return Error(c, queryResult.Error)
	}

	userRead, ok := queryResult.Result.(storage.UserRead)
	if !ok {
		return Error(c, errors.New("Error type assertion"))
	}

	if userRead.UID == (uuid.UUID{}) {
		return Error(c, errors.New("Invitation code not found"))
	}

	if newPassword != confirmNewPassword {
		return Error(c, NewRequestValidationError(NOT_MATCH, "password"))
	}

	// Process
	eventQueryResult := <-s.UserEventQuery.FindAllByID(userRead.UID)
	if eventQueryResult.Error != nil {
		return Error(c, eventQueryResult.Error)
	}

	events := eventQueryResult.Result.([]storage.UserEvent)
	user := repository.NewUserFromHistory(events)

	err = user.ResetPassword(newPassword)
	if err != nil {
		return Error(c, err)
	}

	// Persists //
	resultSave := <-s.UserEventRepo.Save(user.UID, user.Version, user.UncommittedChanges)
	if resultSave != nil {
		return Error(c, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error"))
	}

	// Publish //
	s.publishUncommittedEvents(user)

	data := make(map[string]storage.UserRead)
	data["data"] = MapToUserRead(user)

	return c.JSON(http.StatusOK, data)
}

func (s *AuthServer) SendEmailSubscriber(event interface{}) error {
	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		*config.Config.MailUsername,
		*config.Config.MailPassword,
		*config.Config.MailHost,
	)

	recipients := []string{}
	switch e := event.(type) {
	case domain.OrganizationCreated:
		recipients = append(recipients, e.Email)
	}

	composedMsg := "From: " + *config.Config.MailSender + "\r\n" +
		"To: " + strings.Join(recipients, ",") + "\r\n" +
		"Subject: Tania Verification Code" + "\r\n\r\n" +
		"Your verification code is 123546"

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

func (s *AuthServer) publishUncommittedEvents(entity interface{}) error {
	switch e := entity.(type) {
	case *domain.User:
		for _, v := range e.UncommittedChanges {
			name := structhelper.GetName(v)
			s.EventBus.Publish(name, v)
		}

	case *domain.Organization:
		for _, v := range e.UncommittedChanges {
			name := structhelper.GetName(v)
			s.EventBus.Publish(name, v)
		}

	}

	return nil
}
