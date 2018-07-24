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
	UID         []byte
	Email       string
	Password    string
	CreatedDate time.Time
	LastUpdated time.Time
}

func (s UserReadQueryMysql) FindByID(uid uuid.UUID) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		userRead := storage.UserRead{}
		rowsData := userReadResult{}

		err := s.DB.QueryRow("SELECT * FROM USER_READ WHERE UID = ?", uid.Bytes()).Scan(
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

		userRead = storage.UserRead{
			UID:         userUID,
			Email:       rowsData.Email,
			Password:    []byte(rowsData.Password),
			CreatedDate: rowsData.CreatedDate,
			LastUpdated: rowsData.LastUpdated,
		}

		result <- query.QueryResult{Result: userRead}
		close(result)
	}()

	return result
}

func (s UserReadQueryMysql) FindByUsername(email string) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		userRead := storage.UserRead{}
		rowsData := userReadResult{}

		err := s.DB.QueryRow("SELECT * FROM USER_READ WHERE USERNAME = ?", email).Scan(
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

		userRead = storage.UserRead{
			UID:         userUID,
			Email:       rowsData.Email,
			Password:    []byte(rowsData.Password),
			CreatedDate: rowsData.CreatedDate,
			LastUpdated: rowsData.LastUpdated,
		}

		result <- query.QueryResult{Result: userRead}
		close(result)
	}()

	return result
}

func (s UserReadQueryMysql) FindByUsernameAndPassword(email, password string) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		userRead := storage.UserRead{}
		rowsData := userReadResult{}

		err := s.DB.QueryRow(`SELECT * FROM USER_READ
			WHERE USERNAME = ?`, email).Scan(
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

		err = bcrypt.CompareHashAndPassword([]byte(rowsData.Password), []byte(password))
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
			Password:    []byte(rowsData.Password),
			CreatedDate: rowsData.CreatedDate,
			LastUpdated: rowsData.LastUpdated,
		}

		result <- query.QueryResult{Result: userRead}
		close(result)
	}()

	return result
}
