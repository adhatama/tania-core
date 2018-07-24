package query

import uuid "github.com/satori/go.uuid"

type UserEventQuery interface {
	FindAllByID(userUID uuid.UUID) <-chan QueryResult
}

type UserReadQuery interface {
	FindByID(userUID uuid.UUID) <-chan QueryResult
	FindByUsername(username string) <-chan QueryResult
	FindByUsernameAndPassword(username, password string) <-chan QueryResult
}

type UserAuthQuery interface {
	FindByUserID(userUID uuid.UUID) <-chan QueryResult
}

type OrganizationEventQuery interface {
	FindAllByID(organizationUID uuid.UUID) <-chan QueryResult
}

type OrganizationReadQuery interface {
	FindByID(organizationUID uuid.UUID) <-chan QueryResult
	FindByIDAndVerificationCode(organizationUID uuid.UUID, verificationCode int) <-chan QueryResult
	FindByEmail(email string) <-chan QueryResult
	FindByName(name string) <-chan QueryResult
	FindAll(name string) <-chan QueryResult
}

type QueryResult struct {
	Result interface{}
	Error  error
}
