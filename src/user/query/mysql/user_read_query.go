package mysql

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Tanibox/tania-core/src/user/query"
	"github.com/Tanibox/tania-core/src/user/storage"
	uuid "github.com/satori/go.uuid"
)

type UserReadQueryMysql struct {
	DB *sql.DB
}

func NewUserReadQueryMysql(db *sql.DB) query.UserReadQuery {
	return UserReadQueryMysql{DB: db}
}

type userReadResult struct {
	UID             []byte
	Email           string
	Password        sql.NullString
	Role            string
	Status          string
	Name            sql.NullString
	Gender          sql.NullString
	BirthDate       sql.NullString
	InvitationCode  int
	OrganizationUID []byte
	CreatedDate     time.Time
	LastUpdated     time.Time
}

func (s UserReadQueryMysql) FindByID(uid uuid.UUID) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		userRead := storage.UserRead{}
		rowsData := userReadResult{}

		err := s.DB.QueryRow(`SELECT UID, EMAIL, PASSWORD, ROLE, STATUS,
			NAME, GENDER, BIRTH_DATE,
			INVITATION_CODE, ORGANIZATION_UID, CREATED_DATE, LAST_UPDATED
			FROM USER_READ WHERE UID = ?`, uid.Bytes()).Scan(
			&rowsData.UID,
			&rowsData.Email,
			&rowsData.Password,
			&rowsData.Role,
			&rowsData.Status,
			&rowsData.Name,
			&rowsData.Gender,
			&rowsData.BirthDate,
			&rowsData.InvitationCode,
			&rowsData.OrganizationUID,
			&rowsData.CreatedDate,
			&rowsData.LastUpdated,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: userRead}
		}

		userUID, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		orgUID, err := uuid.FromBytes(rowsData.OrganizationUID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		var password []byte
		if rowsData.Password.Valid {
			password = []byte(rowsData.Password.String)
		}

		var name *string
		if rowsData.Name.Valid {
			name = &rowsData.Name.String
		}

		var gender *string
		if rowsData.Gender.Valid {
			gender = &rowsData.Gender.String
		}

		var birthDate *time.Time
		if rowsData.BirthDate.Valid {
			bd := rowsData.BirthDate.String
			bdTime, err := time.Parse(time.RFC3339, bd)
			birthDate = &bdTime
			if err != nil {
				result <- query.QueryResult{Error: err}
			}
		}

		userRead = storage.UserRead{
			UID:             userUID,
			Email:           rowsData.Email,
			Password:        []byte(password),
			Role:            rowsData.Role,
			Status:          rowsData.Status,
			Name:            name,
			Gender:          gender,
			BirthDate:       birthDate,
			InvitationCode:  rowsData.InvitationCode,
			OrganizationUID: orgUID,
			CreatedDate:     rowsData.CreatedDate,
			LastUpdated:     rowsData.LastUpdated,
		}

		result <- query.QueryResult{Result: userRead}
		close(result)
	}()

	return result
}

func (s UserReadQueryMysql) FindByEmail(email string) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		userRead := storage.UserRead{}
		rowsData := userReadResult{}

		err := s.DB.QueryRow(`SELECT UID, EMAIL, PASSWORD, CREATED_DATE, LAST_UPDATED
			FROM USER_READ WHERE EMAIL = ?`, email).Scan(
			&rowsData.UID,
			&rowsData.Email,
			&rowsData.Password,
			&rowsData.CreatedDate,
			&rowsData.LastUpdated,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: userRead}
		}

		userUID, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		password := ""
		if rowsData.Password.Valid {
			password = rowsData.Password.String
		}

		userRead = storage.UserRead{
			UID:         userUID,
			Email:       rowsData.Email,
			Password:    []byte(password),
			CreatedDate: rowsData.CreatedDate,
			LastUpdated: rowsData.LastUpdated,
		}

		result <- query.QueryResult{Result: userRead}
		close(result)
	}()

	return result
}

func (s UserReadQueryMysql) FindByEmailAndPassword(email, password string) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		userRead := storage.UserRead{}
		rowsData := userReadResult{}

		err := s.DB.QueryRow(`SELECT UID, EMAIL, PASSWORD, CREATED_DATE, LAST_UPDATED
			FROM USER_READ WHERE EMAIL = ?`, email).Scan(
			&rowsData.UID,
			&rowsData.Email,
			&rowsData.Password,
			&rowsData.CreatedDate,
			&rowsData.LastUpdated,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: userRead}
		}

		pwd := ""
		if rowsData.Password.Valid {
			pwd = rowsData.Password.String
		}

		err = bcrypt.CompareHashAndPassword([]byte(pwd), []byte(password))
		if err != nil {
			result <- query.QueryResult{Result: userRead}
		}

		userUID, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		userRead = storage.UserRead{
			UID:         userUID,
			Email:       rowsData.Email,
			Password:    []byte(pwd),
			CreatedDate: rowsData.CreatedDate,
			LastUpdated: rowsData.LastUpdated,
		}

		result <- query.QueryResult{Result: userRead}
		close(result)
	}()

	return result
}

func (s UserReadQueryMysql) FindByOrganizationIDAndInvitationCode(orgUID uuid.UUID, invitationCode int) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		userRead := storage.UserRead{}
		rowsData := userReadResult{}

		err := s.DB.QueryRow(`SELECT UID, EMAIL, CREATED_DATE, LAST_UPDATED
			FROM USER_READ WHERE ORGANIZATION_UID = ? AND INVITATION_CODE = ?`,
			orgUID.Bytes(), invitationCode).Scan(
			&rowsData.UID,
			&rowsData.Email,
			&rowsData.CreatedDate,
			&rowsData.LastUpdated,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: userRead}
		}

		userUID, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		userRead = storage.UserRead{
			UID:         userUID,
			Email:       rowsData.Email,
			CreatedDate: rowsData.CreatedDate,
			LastUpdated: rowsData.LastUpdated,
		}

		result <- query.QueryResult{Result: userRead}
		close(result)
	}()

	return result
}

func (s UserReadQueryMysql) FindByEmailAndResetPasswordCode(email string, code int) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		userRead := storage.UserRead{}
		rowsData := userReadResult{}

		err := s.DB.QueryRow(`SELECT UID, EMAIL, CREATED_DATE, LAST_UPDATED
			FROM USER_READ WHERE EMAIL = ? AND RESET_PASSWORD_CODE = ?`,
			email, code).Scan(
			&rowsData.UID,
			&rowsData.Email,
			&rowsData.CreatedDate,
			&rowsData.LastUpdated,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: userRead}
		}

		userUID, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		userRead = storage.UserRead{
			UID:         userUID,
			Email:       rowsData.Email,
			CreatedDate: rowsData.CreatedDate,
			LastUpdated: rowsData.LastUpdated,
		}

		result <- query.QueryResult{Result: userRead}
		close(result)
	}()

	return result
}
