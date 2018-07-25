package mysql

import (
	"database/sql"

	"github.com/Tanibox/tania-core/src/user/repository"
	"github.com/Tanibox/tania-core/src/user/storage"
)

type OrganizationReadRepositoryMysql struct {
	DB *sql.DB
}

func NewOrganizationReadRepositoryMysql(db *sql.DB) repository.OrganizationReadRepository {
	return &OrganizationReadRepositoryMysql{DB: db}
}

func (f *OrganizationReadRepositoryMysql) Save(organizationRead *storage.OrganizationRead) <-chan error {
	result := make(chan error)

	go func() {
		count := 0
		err := f.DB.QueryRow(`SELECT COUNT(*) FROM ORGANIZATION_READ WHERE UID = ?`, organizationRead.UID.Bytes()).Scan(&count)
		if err != nil {
			result <- err
		}

		if count > 0 {
			_, err := f.DB.Exec(`UPDATE ORGANIZATION_READ SET
				NAME = ?, EMAIL = ?, VERIFICATION_CODE = ?, STATUS = ?, TYPE = ?, TOTAL_MEMBER = ?,
				PROVINCE = ?, CITY = ?, CREATED_DATE = ?
				WHERE UID = ?`,
				organizationRead.Name, organizationRead.Email, organizationRead.VerificationCode,
				organizationRead.Status, organizationRead.Type, organizationRead.TotalMember,
				organizationRead.Province, organizationRead.City, organizationRead.CreatedDate,
				organizationRead.UID.Bytes(),
			)

			if err != nil {
				result <- err
			}
		} else {
			_, err := f.DB.Exec(`INSERT INTO ORGANIZATION_READ
				(UID, NAME, EMAIL, VERIFICATION_CODE, STATUS, TYPE, TOTAL_MEMBER,
					PROVINCE, CITY, CREATED_DATE)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				organizationRead.UID.Bytes(), organizationRead.Name, organizationRead.Email,
				organizationRead.VerificationCode, organizationRead.Status, organizationRead.Type,
				organizationRead.TotalMember, organizationRead.Province, organizationRead.City,
				organizationRead.CreatedDate,
			)

			if err != nil {
				result <- err
			}
		}

		result <- nil
		close(result)
	}()

	return result
}
