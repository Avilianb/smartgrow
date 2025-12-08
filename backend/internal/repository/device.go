package repository

import (
	"database/sql"
	"fmt"
	"time"

	"irrigation-system/backend/internal/models"
)

type DeviceRepository struct {
	db *sql.DB
}

func NewDeviceRepository(db *sql.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// CreateDevice 创建设备
func (r *DeviceRepository) CreateDevice(deviceID, deviceName string, userID int64) (*models.Device, error) {
	now := time.Now().Format(time.RFC3339)
	query := `
		INSERT INTO devices (device_id, user_id, device_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query, deviceID, userID, deviceName, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get device id: %w", err)
	}

	return r.GetDeviceByID(id)
}

// GetDeviceByID 根据ID获取设备
func (r *DeviceRepository) GetDeviceByID(id int64) (*models.Device, error) {
	query := `
		SELECT id, device_id, user_id, device_name, created_at, updated_at
		FROM devices
		WHERE id = ?
	`

	var device models.Device
	var createdAt, updatedAt string
	var userID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&device.ID,
		&device.DeviceID,
		&userID,
		&device.DeviceName,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		return nil, err
	}

	if userID.Valid {
		device.UserID = &userID.Int64
	}

	device.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	device.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &device, nil
}

// GetDeviceByDeviceID 根据device_id获取设备
func (r *DeviceRepository) GetDeviceByDeviceID(deviceID string) (*models.Device, error) {
	query := `
		SELECT id, device_id, user_id, device_name, created_at, updated_at
		FROM devices
		WHERE device_id = ?
	`

	var device models.Device
	var createdAt, updatedAt string
	var userID sql.NullInt64

	err := r.db.QueryRow(query, deviceID).Scan(
		&device.ID,
		&device.DeviceID,
		&userID,
		&device.DeviceName,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		return nil, err
	}

	if userID.Valid {
		device.UserID = &userID.Int64
	}

	device.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	device.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &device, nil
}

// GetDeviceByUserID 根据用户ID获取设备
func (r *DeviceRepository) GetDeviceByUserID(userID int64) (*models.Device, error) {
	query := `
		SELECT id, device_id, user_id, device_name, created_at, updated_at
		FROM devices
		WHERE user_id = ?
	`

	var device models.Device
	var createdAt, updatedAt string
	var userIDVal sql.NullInt64

	err := r.db.QueryRow(query, userID).Scan(
		&device.ID,
		&device.DeviceID,
		&userIDVal,
		&device.DeviceName,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found for this user")
		}
		return nil, err
	}

	if userIDVal.Valid {
		device.UserID = &userIDVal.Int64
	}

	device.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	device.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &device, nil
}

// UpdateDeviceName 更新设备名称
func (r *DeviceRepository) UpdateDeviceName(deviceID string, deviceName string) error {
	now := time.Now().Format(time.RFC3339)
	query := `UPDATE devices SET device_name = ?, updated_at = ? WHERE device_id = ?`

	result, err := r.db.Exec(query, deviceName, now, deviceID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("device not found")
	}

	return nil
}

// DeleteDevice 删除设备
func (r *DeviceRepository) DeleteDevice(deviceID string) error {
	query := `DELETE FROM devices WHERE device_id = ?`
	result, err := r.db.Exec(query, deviceID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("device not found")
	}

	return nil
}

// GetAllDevices 获取所有设备
func (r *DeviceRepository) GetAllDevices() ([]*models.Device, error) {
	query := `
		SELECT id, device_id, user_id, device_name, created_at, updated_at
		FROM devices
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		var device models.Device
		var createdAt, updatedAt string
		var userID sql.NullInt64

		err := rows.Scan(
			&device.ID,
			&device.DeviceID,
			&userID,
			&device.DeviceName,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, err
		}

		if userID.Valid {
			device.UserID = &userID.Int64
		}

		device.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		device.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		devices = append(devices, &device)
	}

	return devices, nil
}
