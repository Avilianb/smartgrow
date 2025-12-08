package repository

import (
	"database/sql"
	"time"

	"irrigation-system/backend/internal/models"
)

type LocationRepository struct {
	db *sql.DB
}

func NewLocationRepository(db *sql.DB) *LocationRepository {
	return &LocationRepository{db: db}
}

// Get retrieves device location
func (r *LocationRepository) Get(deviceID string) (*models.DeviceLocation, error) {
	query := `
		SELECT id, device_id, latitude, longitude, address, updated_at
		FROM device_locations
		WHERE device_id = ?
	`
	var loc models.DeviceLocation
	var updatedAt string
	var address sql.NullString

	err := r.db.QueryRow(query, deviceID).Scan(
		&loc.ID,
		&loc.DeviceID,
		&loc.Latitude,
		&loc.Longitude,
		&address,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if address.Valid {
		loc.Address = &address.String
	}
	loc.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &loc, nil
}

// Upsert inserts or updates device location
func (r *LocationRepository) Upsert(loc *models.DeviceLocation) error {
	query := `
		INSERT INTO device_locations (device_id, latitude, longitude, address, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(device_id) DO UPDATE SET
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			address = excluded.address,
			updated_at = excluded.updated_at
	`
	_, err := r.db.Exec(query,
		loc.DeviceID,
		loc.Latitude,
		loc.Longitude,
		loc.Address,
		loc.UpdatedAt.Format(time.RFC3339),
	)
	return err
}
