package repository

import (
	"database/sql"
	"time"

	"irrigation-system/backend/internal/models"
)

type PlanRepository struct {
	db *sql.DB
}

func NewPlanRepository(db *sql.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

// GetByDate retrieves irrigation plan for a specific date
func (r *PlanRepository) GetByDate(deviceID, date string) (*models.IrrigationPlan, error) {
	query := `
		SELECT id, device_id, date, planned_volume_l, created_at
		FROM irrigation_plan
		WHERE device_id = ? AND date = ?
	`
	var plan models.IrrigationPlan
	var createdAt string

	err := r.db.QueryRow(query, deviceID, date).Scan(
		&plan.ID,
		&plan.DeviceID,
		&plan.Date,
		&plan.PlannedVolumeL,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	plan.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &plan, nil
}

// GetFuturePlans retrieves all future irrigation plans for a device
func (r *PlanRepository) GetFuturePlans(deviceID string, days int) ([]*models.IrrigationPlan, error) {
	query := `
		SELECT id, device_id, date, planned_volume_l, created_at
		FROM irrigation_plan
		WHERE device_id = ? AND date >= date('now')
		ORDER BY date ASC
		LIMIT ?
	`
	rows, err := r.db.Query(query, deviceID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*models.IrrigationPlan
	for rows.Next() {
		var plan models.IrrigationPlan
		var createdAt string
		if err := rows.Scan(
			&plan.ID,
			&plan.DeviceID,
			&plan.Date,
			&plan.PlannedVolumeL,
			&createdAt,
		); err != nil {
			return nil, err
		}
		plan.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		plans = append(plans, &plan)
	}

	return plans, nil
}

// DeleteFuturePlans deletes all future plans for a device
func (r *PlanRepository) DeleteFuturePlans(deviceID string) error {
	query := `DELETE FROM irrigation_plan WHERE device_id = ? AND date >= date('now')`
	_, err := r.db.Exec(query, deviceID)
	return err
}

// CreateBatch inserts multiple irrigation plans
func (r *PlanRepository) CreateBatch(plans []*models.IrrigationPlan) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO irrigation_plan (device_id, date, planned_volume_l, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(device_id, date) DO UPDATE SET
			planned_volume_l = excluded.planned_volume_l,
			created_at = excluded.created_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, plan := range plans {
		_, err := stmt.Exec(
			plan.DeviceID,
			plan.Date,
			plan.PlannedVolumeL,
			plan.CreatedAt.Format(time.RFC3339),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
