package mysql

import (
	"database/sql"
	"time"

	"github.com/Tanibox/tania-core/src/user/query"
	"github.com/Tanibox/tania-core/src/user/storage"
	uuid "github.com/satori/go.uuid"
)

type OrganizationReadQueryMysql struct {
	DB *sql.DB
}

func NewOrganizationReadQueryMysql(db *sql.DB) query.OrganizationReadQuery {
	return OrganizationReadQueryMysql{DB: db}
}

type organizationReadResult struct {
	UID              []byte
	Name             string
	Email            string
	VerificationCode int
	Status           string
	CreatedDate      time.Time
}

func (s OrganizationReadQueryMysql) FindByID(uid uuid.UUID) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		orgRead := storage.OrganizationRead{}
		rowsData := organizationReadResult{}

		err := s.DB.QueryRow("SELECT * FROM ORGANIZATION_READ WHERE UID = ?", uid.Bytes()).Scan(
			&rowsData.UID,
			&rowsData.Name,
			&rowsData.Email,
			&rowsData.VerificationCode,
			&rowsData.Status,
			&rowsData.CreatedDate,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: orgRead}
		}

		orgUID, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		orgRead = storage.OrganizationRead{
			UID:              orgUID,
			Name:             rowsData.Name,
			Email:            rowsData.Email,
			VerificationCode: rowsData.VerificationCode,
			Status:           rowsData.Status,
			CreatedDate:      rowsData.CreatedDate,
		}

		result <- query.QueryResult{Result: orgRead}
		close(result)
	}()

	return result
}

func (s OrganizationReadQueryMysql) FindByEmail(email string) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		orgRead := storage.OrganizationRead{}
		rowsData := organizationReadResult{}

		err := s.DB.QueryRow("SELECT * FROM ORGANIZATION_READ WHERE EMAIL = ?", email).Scan(
			&rowsData.UID,
			&rowsData.Name,
			&rowsData.Email,
			&rowsData.VerificationCode,
			&rowsData.Status,
			&rowsData.CreatedDate,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: orgRead}
		}

		orgUID, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		orgRead = storage.OrganizationRead{
			UID:              orgUID,
			Name:             rowsData.Name,
			Email:            rowsData.Email,
			VerificationCode: rowsData.VerificationCode,
			Status:           rowsData.Status,
			CreatedDate:      rowsData.CreatedDate,
		}

		result <- query.QueryResult{Result: orgRead}
		close(result)
	}()

	return result
}

func (s OrganizationReadQueryMysql) FindByName(name string) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		orgRead := storage.OrganizationRead{}
		rowsData := organizationReadResult{}

		err := s.DB.QueryRow("SELECT * FROM ORGANIZATION_READ WHERE NAME = ?", name).Scan(
			&rowsData.UID,
			&rowsData.Name,
			&rowsData.Email,
			&rowsData.VerificationCode,
			&rowsData.Status,
			&rowsData.CreatedDate,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: orgRead}
		}

		orgUID, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		orgRead = storage.OrganizationRead{
			UID:              orgUID,
			Name:             rowsData.Name,
			Email:            rowsData.Email,
			VerificationCode: rowsData.VerificationCode,
			Status:           rowsData.Status,
			CreatedDate:      rowsData.CreatedDate,
		}

		result <- query.QueryResult{Result: orgRead}
		close(result)
	}()

	return result
}

func (s OrganizationReadQueryMysql) FindByIDAndVerificationCode(uid uuid.UUID, verificationCode int) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		orgRead := storage.OrganizationRead{}
		rowsData := organizationReadResult{}

		err := s.DB.QueryRow("SELECT * FROM ORGANIZATION_READ WHERE UID = ?", uid.Bytes()).Scan(
			&rowsData.UID,
			&rowsData.Name,
			&rowsData.Email,
			&rowsData.VerificationCode,
			&rowsData.Status,
			&rowsData.CreatedDate,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: orgRead}
		}

		orgUID, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		orgRead = storage.OrganizationRead{
			UID:              orgUID,
			Name:             rowsData.Name,
			Email:            rowsData.Email,
			VerificationCode: rowsData.VerificationCode,
			Status:           rowsData.Status,
			CreatedDate:      rowsData.CreatedDate,
		}

		result <- query.QueryResult{Result: orgRead}
		close(result)
	}()

	return result
}

func (s OrganizationReadQueryMysql) FindAll(name string) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		orgRead := storage.OrganizationRead{}
		rowsData := organizationReadResult{}

		err := s.DB.QueryRow(`SELECT * FROM ORGANIZATION_READ
			WHERE NAME LIKE ?`, "%"+name+"%").Scan(
			&rowsData.UID,
			&rowsData.Name,
			&rowsData.Email,
			&rowsData.VerificationCode,
			&rowsData.Status,
			&rowsData.CreatedDate,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: orgRead}
		}

		orgUID, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		orgRead = storage.OrganizationRead{
			UID:              orgUID,
			Name:             rowsData.Name,
			Email:            rowsData.Email,
			VerificationCode: rowsData.VerificationCode,
			Status:           rowsData.Status,
			CreatedDate:      rowsData.CreatedDate,
		}

		result <- query.QueryResult{Result: orgRead}
		close(result)
	}()

	return result
}
