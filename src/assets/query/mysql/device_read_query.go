package mysql

import (
	"database/sql"
	"time"

	"github.com/satori/go.uuid"

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
	UID         []byte
	DeviceID    string
	Name        string
	TopicName   string
	Status      string
	Description string
	CreatedDate time.Time
}

func (s DeviceReadQueryMysql) FindByID(deviceUID uuid.UUID) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		deviceRead := storage.DeviceRead{}
		rowsData := deviceReadResult{}

		err := s.DB.QueryRow("SELECT * FROM DEVICE_READ WHERE UID = ?", deviceUID.Bytes()).Scan(
			&rowsData.UID,
			&rowsData.DeviceID,
			&rowsData.Name,
			&rowsData.TopicName,
			&rowsData.Status,
			&rowsData.Description,
			&rowsData.CreatedDate,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: deviceRead}
		}

		uid, err := uuid.FromBytes(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Result: err}
		}

		deviceRead = storage.DeviceRead{
			UID:         uid,
			DeviceID:    rowsData.DeviceID,
			Name:        rowsData.Name,
			TopicName:   rowsData.TopicName,
			Status:      rowsData.Status,
			Description: rowsData.Description,
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
				&rowsData.UID,
				&rowsData.DeviceID,
				&rowsData.Name,
				&rowsData.TopicName,
				&rowsData.Status,
				&rowsData.Description,
				&rowsData.CreatedDate,
			)

			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			uid, err := uuid.FromBytes(rowsData.UID)
			if err != nil {
				result <- query.QueryResult{Result: err}
			}

			deviceReads = append(deviceReads, storage.DeviceRead{
				UID:         uid,
				DeviceID:    rowsData.DeviceID,
				Name:        rowsData.Name,
				TopicName:   rowsData.TopicName,
				Status:      rowsData.Status,
				Description: rowsData.Description,
				CreatedDate: rowsData.CreatedDate,
			})
		}

		result <- query.QueryResult{Result: deviceReads}
		close(result)
	}()

	return result
}