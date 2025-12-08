package repository

import (
	"database/sql"
	"time"

	"irrigation-system/backend/internal/models"
)

type LogRepository struct {
	db *sql.DB
}

func NewLogRepository(db *sql.DB) *LogRepository {
	return &LogRepository{db: db}
}

// Create inserts a new log record
func (r *LogRepository) Create(log *models.DeviceLog) error {
	query := `
		INSERT INTO device_log (device_id, timestamp, level, message, extra)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		log.DeviceID,
		log.Timestamp.Format(time.RFC3339),
		log.Level,
		log.Message,
		log.Extra,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	log.ID = id
	return nil
}

// Query retrieves logs with filters
func (r *LogRepository) Query(deviceID string, level *string, startTime *time.Time, limit, offset int) ([]*models.DeviceLog, int, error) {
	// Build query
	query := `
		SELECT id, device_id, timestamp, level, message, extra
		FROM device_log
		WHERE device_id = ?
	`
	countQuery := `SELECT COUNT(*) FROM device_log WHERE device_id = ?`
	args := []interface{}{deviceID}
	countArgs := []interface{}{deviceID}

	if level != nil && *level != "" {
		query += ` AND level = ?`
		countQuery += ` AND level = ?`
		args = append(args, *level)
		countArgs = append(countArgs, *level)
	}
	if startTime != nil {
		query += ` AND timestamp >= ?`
		countQuery += ` AND timestamp >= ?`
		timeStr := startTime.Format(time.RFC3339)
		args = append(args, timeStr)
		countArgs = append(countArgs, timeStr)
	}

	// Get total count
	var total int
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get data
	query += ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.DeviceLog
	for rows.Next() {
		var log models.DeviceLog
		var timestamp string
		var extra sql.NullString

		if err := rows.Scan(
			&log.ID,
			&log.DeviceID,
			&timestamp,
			&log.Level,
			&log.Message,
			&extra,
		); err != nil {
			return nil, 0, err
		}

		log.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
		if extra.Valid {
			log.Extra = &extra.String
		}
		logs = append(logs, &log)
	}

	return logs, total, nil
}
