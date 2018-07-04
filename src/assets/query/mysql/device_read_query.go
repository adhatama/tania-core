package mysql

import (
	"database/sql"
	"time"

	"github.com/Tanibox/tania-core/src/assets/query"
	"github.com/Tanibox/tania-core/src/assets/storage"
)

type DeviceReadQueryMysql struct {
	DB *sql.DB
}

func NewDeviceReadQueryMysql(db *sql.DB) query.DeviceReadQuery {
	return DeviceReadQueryMysql{DB: db}
}

type deviceReadResult struct {
	DeviceID    string
	Name        string
	TopicName   string
	Status      string
	CreatedDate time.Time
}

func (s DeviceReadQueryMysql) FindByID(deviceID string) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		deviceRead := storage.DeviceRead{}
		rowsData := deviceReadResult{}

		err := s.DB.QueryRow("SELECT * FROM DEVICE_READ WHERE DEVICE_ID = ?", deviceID).Scan(
			&rowsData.DeviceID,
			&rowsData.Name,
			&rowsData.TopicName,
			&rowsData.Status,
			&rowsData.CreatedDate,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: deviceRead}
		}

		deviceRead = storage.DeviceRead{
			DeviceID:    rowsData.DeviceID,
			Name:        rowsData.Name,
			TopicName:   rowsData.TopicName,
			Status:      rowsData.Status,
			CreatedDate: rowsData.CreatedDate,
		}

		result <- query.QueryResult{Result: deviceRead}
		close(result)
	}()

	return result
}

func (s DeviceReadQueryMysql) FindAll() <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		deviceReads := []storage.DeviceRead{}
		rowsData := deviceReadResult{}

		rows, err := s.DB.Query("SELECT * FROM DEVICE_READ ORDER BY CREATED_DATE ASC")
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		for rows.Next() {
			err = rows.Scan(
				&rowsData.DeviceID,
				&rowsData.Name,
				&rowsData.TopicName,
				&rowsData.Status,
				&rowsData.CreatedDate,
			)

			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			deviceReads = append(deviceReads, storage.DeviceRead{
				DeviceID:    rowsData.DeviceID,
				Name:        rowsData.Name,
				TopicName:   rowsData.TopicName,
				Status:      rowsData.Status,
				CreatedDate: rowsData.CreatedDate,
			})
		}

		result <- query.QueryResult{Result: deviceReads}
		close(result)
	}()

	return result
}
