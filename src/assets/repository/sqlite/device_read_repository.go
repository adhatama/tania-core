package sqlite

import (
	"database/sql"
	"time"

	"github.com/Tanibox/tania-core/src/assets/repository"
	"github.com/Tanibox/tania-core/src/assets/storage"
)

type DeviceReadRepositorySqlite struct {
	DB *sql.DB
}

func NewDeviceReadRepositorySqlite(db *sql.DB) repository.DeviceReadRepository {
	return &DeviceReadRepositorySqlite{DB: db}
}

func (f *DeviceReadRepositorySqlite) Save(deviceRead *storage.DeviceRead) <-chan error {
	result := make(chan error)

	go func() {
		count := 0
		err := f.DB.QueryRow(`SELECT COUNT(*) FROM DEVICE_READ WHERE UID = ?`, deviceRead.UID).Scan(&count)
		if err != nil {
			result <- err
		}

		if count > 0 {
			_, err := f.DB.Exec(`UPDATE DEVICE_READ SET
				DEVICE_ID = ?, NAME = ?, TOPIC_NAME = ?, STATUS = ?, DESCRIPTION = ?, CREATED_DATE = ?
				WHERE UID = ?`,
				deviceRead.DeviceID, deviceRead.Name, deviceRead.TopicName, deviceRead.Status,
				deviceRead.Description, deviceRead.CreatedDate.Format(time.RFC3339),
				deviceRead.UID,
			)

			if err != nil {
				result <- err
			}
		} else {
			_, err := f.DB.Exec(`INSERT INTO DEVICE_READ
				(UID, DEVICE_ID, NAME, TOPIC_NAME, STATUS, DESCRIPTION, CREATED_DATE)
				VALUES (?, ?, ?, ?, ?, ?, ?)`,
				deviceRead.UID, deviceRead.DeviceID, deviceRead.Name, deviceRead.TopicName,
				deviceRead.Status, deviceRead.Description, deviceRead.CreatedDate.Format(time.RFC3339),
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
