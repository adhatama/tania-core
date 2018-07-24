package mysql

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Tanibox/tania-core/src/helper/structhelper"
	"github.com/Tanibox/tania-core/src/user/decoder"
	"github.com/Tanibox/tania-core/src/user/repository"
	uuid "github.com/satori/go.uuid"
)

type OrganizationEventRepositoryMysql struct {
	DB *sql.DB
}

func NewOrganizationEventRepositoryMysql(db *sql.DB) repository.OrganizationEventRepository {
	return &OrganizationEventRepositoryMysql{DB: db}
}

func (f *OrganizationEventRepositoryMysql) Save(uid uuid.UUID, latestVersion int, events []interface{}) <-chan error {
	result := make(chan error)

	go func() {
		for _, v := range events {
			stmt, err := f.DB.Prepare(`INSERT INTO ORGANIZATION_EVENT
				(ORGANIZATION_UID, VERSION, CREATED_DATE, EVENT) VALUES (?, ?, ?, ?)`)
			if err != nil {
				result <- err
			}

			latestVersion++

			e, err := json.Marshal(decoder.EventWrapper{
				EventName: structhelper.GetName(v),
				EventData: v,
			})

			_, err = stmt.Exec(uid.Bytes(), latestVersion, time.Now(), e)
			if err != nil {
				result <- err
			}
		}

		result <- nil
		close(result)
	}()

	return result
}
