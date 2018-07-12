package sqlite

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"

	"github.com/Tanibox/tania-core/src/assets/decoder"
	"github.com/Tanibox/tania-core/src/assets/query"
	"github.com/Tanibox/tania-core/src/assets/storage"
)

type DeviceEventQuerySqlite struct {
	DB *sql.DB
}

func NewDeviceEventQuerySqlite(db *sql.DB) query.DeviceEventQuery {
	return &DeviceEventQuerySqlite{DB: db}
}

func (f *DeviceEventQuerySqlite) FindAllByID(deviceUID uuid.UUID) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		events := []storage.DeviceEvent{}

		rows, err := f.DB.Query("SELECT * FROM DEVICE_EVENT WHERE DEVICE_UID = ? ORDER BY VERSION ASC",
			deviceUID,
		)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		rowsData := struct {
			ID          int
			DeviceUID   string
			Version     int
			CreatedDate string
			Event       []byte
		}{}

		for rows.Next() {
			rows.Scan(&rowsData.ID, &rowsData.DeviceUID, &rowsData.Version, &rowsData.CreatedDate, &rowsData.Event)

			wrapper := decoder.DeviceEventWrapper{}
			err := json.Unmarshal(rowsData.Event, &wrapper)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			deviceUID, err := uuid.FromString(rowsData.DeviceUID)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			createdDate, err := time.Parse(time.RFC3339, rowsData.CreatedDate)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			events = append(events, storage.DeviceEvent{
				DeviceUID:   deviceUID,
				Version:     rowsData.Version,
				CreatedDate: createdDate,
				Event:       wrapper.EventData,
			})
		}

		result <- query.QueryResult{Result: events}
		close(result)
	}()

	return result
}
