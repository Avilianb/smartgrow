package repository

import (
	"database/sql"
	"time"

	"irrigation-system/backend/internal/models"
)

type SensorDataRepository struct {
	db *sql.DB
}

func NewSensorDataRepository(db *sql.DB) *SensorDataRepository {
	return &SensorDataRepository{db: db}
}

// Create inserts a new sensor data record
func (r *SensorDataRepository) Create(data *models.SensorData) error {
	query := `
		INSERT INTO sensor_data
		(device_id, timestamp, temperature_c, humidity_pct, soil_raw, rain_analog, rain_digital, pump_state, shade_state)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		data.DeviceID,
		data.Timestamp.Format(time.RFC3339),
		data.TemperatureC,
		data.HumidityPct,
		data.SoilRaw,
		data.RainAnalog,
		data.RainDigital,
		data.PumpState,
		data.ShadeState,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	data.ID = id
	return nil
}

// GetLatest retrieves the latest sensor data for a device
func (r *SensorDataRepository) GetLatest(deviceID string) (*models.SensorData, error) {
	query := `
		SELECT id, device_id, timestamp, temperature_c, humidity_pct, soil_raw, rain_analog, rain_digital, pump_state, shade_state
		FROM sensor_data
		WHERE device_id = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var data models.SensorData
	var timestamp string
	err := r.db.QueryRow(query, deviceID).Scan(
		&data.ID,
		&data.DeviceID,
		&timestamp,
		&data.TemperatureC,
		&data.HumidityPct,
		&data.SoilRaw,
		&data.RainAnalog,
		&data.RainDigital,
		&data.PumpState,
		&data.ShadeState,
	)
	if err != nil {
		return nil, err
	}

	data.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
	return &data, nil
}

// GetHistory retrieves historical sensor data
func (r *SensorDataRepository) GetHistory(deviceID string, startTime, endTime *time.Time, limit, offset int) ([]*models.SensorData, int, error) {
	// Build query
	query := `
		SELECT id, device_id, timestamp, temperature_c, humidity_pct, soil_raw, rain_analog, rain_digital, pump_state, shade_state
		FROM sensor_data
		WHERE device_id = ?
	`
	countQuery := `SELECT COUNT(*) FROM sensor_data WHERE device_id = ?`
	args := []interface{}{deviceID}
	countArgs := []interface{}{deviceID}

	if startTime != nil {
		query += ` AND timestamp >= ?`
		countQuery += ` AND timestamp >= ?`
		timeStr := startTime.Format(time.RFC3339)
		args = append(args, timeStr)
		countArgs = append(countArgs, timeStr)
	}
	if endTime != nil {
		query += ` AND timestamp <= ?`
		countQuery += ` AND timestamp <= ?`
		timeStr := endTime.Format(time.RFC3339)
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

	var dataList []*models.SensorData
	for rows.Next() {
		var data models.SensorData
		var timestamp string
		if err := rows.Scan(
			&data.ID,
			&data.DeviceID,
			&timestamp,
			&data.TemperatureC,
			&data.HumidityPct,
			&data.SoilRaw,
			&data.RainAnalog,
			&data.RainDigital,
			&data.PumpState,
			&data.ShadeState,
		); err != nil {
			return nil, 0, err
		}
		data.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
		dataList = append(dataList, &data)
	}

	return dataList, total, nil
}

// GetTodayIrrigationVolume calculates the total irrigation volume for today
func (r *SensorDataRepository) GetTodayIrrigationVolume(deviceID string) (float64, error) {
	// 这里简化处理，实际应该根据水泵开启时间和流量计算
	// 目前返回 0，由后续任务完善
	return 0, nil
}
