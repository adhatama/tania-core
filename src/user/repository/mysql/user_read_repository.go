package mysql

import (
	"database/sql"

	"github.com/Tanibox/tania-core/src/user/repository"
	"github.com/Tanibox/tania-core/src/user/storage"
)

type UserReadRepositoryMysql struct {
	DB *sql.DB
}

func NewUserReadRepositoryMysql(db *sql.DB) repository.UserReadRepository {
	return &UserReadRepositoryMysql{DB: db}
}

func (f *UserReadRepositoryMysql) Save(userRead *storage.UserRead) <-chan error {
	result := make(chan error)

	go func() {
		count := 0
		err := f.DB.QueryRow(`SELECT COUNT(*) FROM USER_READ WHERE UID = ?`, userRead.UID.Bytes()).Scan(&count)
		if err != nil {
			result <- err
		}

		if count > 0 {
			_, err := f.DB.Exec(`UPDATE USER_READ SET
				EMAIL = ?, PASSWORD = ?, ROLE = ?, STATUS = ?, ORGANIZATION_UID = ?,
				NAME = ?, GENDER = ?, BIRTH_DATE = ?,
				INVITATION_CODE = ?, RESET_PASSWORD_CODE = ?, CREATED_DATE = ?, LAST_UPDATED = ?
				WHERE UID = ?`,
				userRead.Email, userRead.Password, userRead.Role, userRead.Status, userRead.OrganizationUID.Bytes(),
				userRead.Name, userRead.Gender, userRead.BirthDate,
				userRead.InvitationCode, userRead.ResetPasswordCode,
				userRead.CreatedDate, userRead.LastUpdated,
				userRead.UID.Bytes())

			if err != nil {
				result <- err
			}
		} else {
			_, err := f.DB.Exec(`INSERT INTO USER_READ
				(UID, EMAIL, PASSWORD, ROLE, STATUS, ORGANIZATION_UID, INVITATION_CODE,
					NAME, GENDER, BIRTH_DATE, RESET_PASSWORD_CODE,
					CREATED_DATE, LAST_UPDATED)
				VALUES (?, ?, ?, ?, ?, ?, ?, ? ,?, ?, ?, ?, ?)`,
				userRead.UID.Bytes(), userRead.Email, userRead.Password,
				userRead.Role, userRead.Status, userRead.OrganizationUID.Bytes(), userRead.InvitationCode,
				&userRead.Name, &userRead.Gender, &userRead.BirthDate,
				userRead.ResetPasswordCode, userRead.CreatedDate, userRead.LastUpdated)

			if err != nil {
				result <- err
			}
		}

		result <- nil
		close(result)
	}()

	return result
}
