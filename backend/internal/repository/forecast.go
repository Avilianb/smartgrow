package repository

import (
	"database/sql"
	"time"

	"irrigation-system/backend/internal/models"
)

type ForecastRepository struct {
	db *sql.DB
}

func NewForecastRepository(db *sql.DB) *ForecastRepository {
	return &ForecastRepository{db: db}
}

// DeleteFutureForecasts deletes all forecasts from today onwards
func (r *ForecastRepository) DeleteFutureForecasts() error {
	query := `DELETE FROM rain_forecast WHERE date >= date('now')`
	_, err := r.db.Exec(query)
	return err
}

// CreateBatch inserts multiple forecast records
func (r *ForecastRepository) CreateBatch(forecasts []*models.RainForecast) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO rain_forecast (date, temp_max, temp_min, precip_mm, humidity_pct, raw_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, forecast := range forecasts {
		_, err := stmt.Exec(
			forecast.Date,
			forecast.TempMax,
			forecast.TempMin,
			forecast.PrecipMm,
			forecast.HumidityPct,
			forecast.RawJSON,
			forecast.CreatedAt.Format(time.RFC3339),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetForecastDays retrieves forecast data for the next N days
func (r *ForecastRepository) GetForecastDays(days int) ([]*models.RainForecast, error) {
	query := `
		SELECT id, date, temp_max, temp_min, precip_mm, humidity_pct, raw_json, created_at
		FROM rain_forecast
		WHERE date >= date('now')
		ORDER BY date ASC
		LIMIT ?
	`
	rows, err := r.db.Query(query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var forecasts []*models.RainForecast
	for rows.Next() {
		var f models.RainForecast
		var createdAt string
		if err := rows.Scan(
			&f.ID,
			&f.Date,
			&f.TempMax,
			&f.TempMin,
			&f.PrecipMm,
			&f.HumidityPct,
			&f.RawJSON,
			&createdAt,
		); err != nil {
			return nil, err
		}
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		forecasts = append(forecasts, &f)
	}

	return forecasts, nil
}
