package mysql

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Tanibox/tania-core/src/user/decoder"
	"github.com/Tanibox/tania-core/src/user/query"
	"github.com/Tanibox/tania-core/src/user/storage"
	uuid "github.com/satori/go.uuid"
)

type OrganizationEventQueryMysql struct {
	DB *sql.DB
}

func NewOrganizationEventQueryMysql(db *sql.DB) query.OrganizationEventQuery {
	return &OrganizationEventQueryMysql{DB: db}
}

func (f *OrganizationEventQueryMysql) FindAllByID(uid uuid.UUID) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		events := []storage.OrganizationEvent{}

		rows, err := f.DB.Query("SELECT * FROM ORGANIZATION_EVENT WHERE ORGANIZATION_UID = ? ORDER BY VERSION ASC", uid.Bytes())
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		rowsData := struct {
			ID              int
			OrganizationUID []byte
			Version         int
			CreatedDate     time.Time
			Event           []byte
		}{}

		for rows.Next() {
			rows.Scan(&rowsData.ID, &rowsData.OrganizationUID, &rowsData.Version, &rowsData.CreatedDate, &rowsData.Event)

			wrapper := decoder.OrganizationEventWrapper{}
			err := json.Unmarshal(rowsData.Event, &wrapper)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			orgUID, err := uuid.FromBytes(rowsData.OrganizationUID)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			events = append(events, storage.OrganizationEvent{
				OrganizationUID: orgUID,
				Version:         rowsData.Version,
				CreatedDate:     rowsData.CreatedDate,
				Event:           wrapper.EventData,
			})
		}

		result <- query.QueryResult{Result: events}
		close(result)
	}()

	return result
}
