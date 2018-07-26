package domain

// UserError is a custom error from Go built-in error
type UserError struct {
	Code int
}

const (
	UserErrorEmailEmptyCode = iota
	UserErrorPasswordEmptyCode
	UserErrorWrongPasswordCode
	UserErrorEmailExistsCode
	UserErrorPasswordConfirmationNotMatchCode
	UserChangePasswordErrorWrongOldPasswordCode
)

func (e UserError) Error() string {
	switch e.Code {
	case UserErrorEmailEmptyCode:
		return "Email cannot be empty"
	case UserErrorPasswordEmptyCode:
		return "Password cannot be empty"
	case UserErrorWrongPasswordCode:
		return "Wrong password"
	case UserErrorEmailExistsCode:
		return "Email already exists"
	case UserErrorPasswordConfirmationNotMatchCode:
		return "Password confirmation didn't match"
	case UserChangePasswordErrorWrongOldPasswordCode:
		return "Invalid old password"
	default:
		return "Unrecognized user error code"
	}
}
