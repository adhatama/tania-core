package mysql

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Tanibox/tania-core/src/assets/decoder"
	"github.com/Tanibox/tania-core/src/assets/query"
	"github.com/Tanibox/tania-core/src/assets/storage"
)

type DeviceEventQueryMysql struct {
	DB *sql.DB
}

func NewDeviceEventQueryMysql(db *sql.DB) query.DeviceEventQuery {
	return &DeviceEventQueryMysql{DB: db}
}

func (f *DeviceEventQueryMysql) FindAllByID(deviceID string) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		events := []storage.DeviceEvent{}

		rows, err := f.DB.Query("SELECT * FROM DEVICE_EVENT WHERE DEVICE_ID = ? ORDER BY VERSION ASC", deviceID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		rowsData := struct {
			ID          int
			DeviceID    string
			Version     int
			CreatedDate time.Time
			Event       []byte
		}{}

		for rows.Next() {
			rows.Scan(&rowsData.ID, &rowsData.DeviceID, &rowsData.Version, &rowsData.CreatedDate, &rowsData.Event)

			wrapper := decoder.DeviceEventWrapper{}
			err := json.Unmarshal(rowsData.Event, &wrapper)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			events = append(events, storage.DeviceEvent{
				DeviceID:    deviceID,
				Version:     rowsData.Version,
				CreatedDate: rowsData.CreatedDate,
				Event:       wrapper.EventData,
			})
		}

		result <- query.QueryResult{Result: events}
		close(result)
	}()

	return result
}
