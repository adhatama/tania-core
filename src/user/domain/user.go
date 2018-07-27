package domain

import (
	"errors"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UID               uuid.UUID
	Email             string
	Password          []byte
	Role              string
	Status            string
	OrganizationUID   uuid.UUID
	InvitationCode    int
	ResetPasswordCode int

	Name      *string
	Gender    *string
	BirthDate *time.Time

	CreatedDate time.Time
	LastUpdated time.Time

	// Events
	Version            int
	UncommittedChanges []interface{}
}

type UserService interface {
	FindUserByEmail(email string) (UserServiceResult, error)
}

type UserServiceResult struct {
	UID   uuid.UUID
	Email string
}

const (
	UserRoleAdmin = "ADMIN"
	UserRoleUser  = "USER"
)

const (
	UserStatusPendingConfirmation = "PENDING_CONFIRMATION"
	UserStatusConfirmed           = "CONFIRMED"
	UserStatusCompleted           = "COMPLETED"
)

func (state *User) TrackChange(event interface{}) {
	state.UncommittedChanges = append(state.UncommittedChanges, event)
	state.Transition(event)
}

func (state *User) Transition(event interface{}) {
	switch e := event.(type) {
	case UserCreated:
		state.UID = e.UID
		state.Email = e.Email
		state.Password = e.Password
		state.Role = e.Role
		state.Status = e.Status
		state.InvitationCode = e.InvitationCode
		state.OrganizationUID = e.OrganizationUID
		state.CreatedDate = e.CreatedDate
		state.LastUpdated = e.LastUpdated

	case PasswordChanged:
		state.Password = e.NewPassword
		state.LastUpdated = e.DateChanged

	case UserProfileChanged:
		state.Name = &e.Name
		state.Gender = &e.Gender
		state.BirthDate = &e.BirthDate

	case UserVerified:
		state.Status = e.Status

	case ResetPasswordRequested:
		state.ResetPasswordCode = e.ResetPasswordCode

	case InitialUserProfileSet:
		state.Name = &e.Name
		state.Gender = &e.Gender
		state.BirthDate = &e.BirthDate
		state.Password = e.Password
		state.Status = e.Status
		state.LastUpdated = e.DateChanged
	}
}

func CreateUser(userService UserService, organizationUID uuid.UUID, email, role string) (*User, error) {
	if email == "" {
		return nil, UserError{UserErrorEmailEmptyCode}
	}

	userResult, err := userService.FindUserByEmail(email)
	if err != nil {
		return nil, err
	}

	if userResult.UID != (uuid.UUID{}) {
		return nil, UserError{UserErrorEmailExistsCode}
	}

	// hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// if err != nil {
	// 	return nil, err
	// }

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	// Generate 6 digit random number
	rand.Seed(time.Now().UnixNano())
	code := 100000 + rand.Intn(900000)

	status := ""
	switch role {
	case UserRoleAdmin:
		status = UserStatusConfirmed
	case UserRoleUser:
		status = UserStatusPendingConfirmation
	default:
		return nil, errors.New("Role not found")
	}

	user := &User{
		UID:             uid,
		Email:           email,
		OrganizationUID: organizationUID,
		InvitationCode:  code,
		Role:            role,
		Status:          status,
	}

	now := time.Now()

	user.TrackChange(UserCreated{
		UID:             user.UID,
		Email:           user.Email,
		Password:        user.Password,
		OrganizationUID: user.OrganizationUID,
		InvitationCode:  user.InvitationCode,
		Role:            user.Role,
		Status:          user.Status,
		CreatedDate:     now,
		LastUpdated:     now,
	})

	return user, nil
}

func (u *User) ChangePassword(oldPassword, newPassword, newConfirmPassword string) error {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(oldPassword))
	if err != nil {
		return UserError{UserChangePasswordErrorWrongOldPasswordCode}
	}

	err = validatePassword(newPassword, newConfirmPassword)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.TrackChange(PasswordChanged{
		UID:         u.UID,
		NewPassword: hash,
		DateChanged: time.Now(),
	})

	return nil
}

func (u *User) IsPasswordValid(password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	if err != nil {
		return false, UserError{UserErrorWrongPasswordCode}
	}

	return true, nil
}

func (u *User) ChangeProfile(name, gender string, birthDate time.Time) error {
	if name == "" {
		return errors.New("Name cannot be empty")
	}

	if gender == "" {
		return errors.New("Gender cannot be empty")
	}

	if birthDate == (time.Time{}) {
		return errors.New("Birth Date cannot be empty")
	}

	u.TrackChange(UserProfileChanged{
		UID:       u.UID,
		Name:      name,
		Gender:    gender,
		BirthDate: birthDate,
	})

	return nil
}

func (u *User) SetInitialUserProfile(name, gender string, birthDate time.Time, password string) error {
	if u.Status == UserStatusCompleted {
		return errors.New("User profile cannot be changed from here because the values are already filled")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if name == "" {
		return errors.New("Name cannot be empty")
	}

	if gender == "" {
		return errors.New("Gender cannot be empty")
	}

	if birthDate == (time.Time{}) {
		return errors.New("Birth Date cannot be empty")
	}

	u.TrackChange(InitialUserProfileSet{
		UID:         u.UID,
		Name:        name,
		Gender:      gender,
		BirthDate:   birthDate,
		Password:    hash,
		Status:      UserStatusCompleted,
		DateChanged: time.Now(),
	})

	return nil
}

func (u *User) VerifyInvitation() error {
	if u.Status == UserStatusConfirmed {
		return errors.New("Status already confirmed")
	}

	u.TrackChange(UserVerified{
		UID:    u.UID,
		Email:  u.Email,
		Status: UserStatusConfirmed,
	})

	return nil
}

func (u *User) RequestResetPassword() error {
	// Generate 6 digit random number
	rand.Seed(time.Now().UnixNano())
	code := 100000 + rand.Intn(900000)

	u.TrackChange(ResetPasswordRequested{
		UID:               u.UID,
		Email:             u.Email,
		ResetPasswordCode: code,
	})

	return nil
}

func (u *User) ResetPassword(newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.TrackChange(PasswordChanged{
		UID:         u.UID,
		NewPassword: hash,
		DateChanged: time.Now(),
	})

	return nil
}

func validatePassword(password, confirmPassword string) error {
	if password == "" {
		return UserError{UserErrorPasswordEmptyCode}
	}

	if password != confirmPassword {
		return UserError{UserErrorPasswordConfirmationNotMatchCode}
	}

	return nil
}
