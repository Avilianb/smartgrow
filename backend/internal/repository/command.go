package repository

import (
	"database/sql"
	"time"

	"irrigation-system/backend/internal/models"
)

type CommandRepository struct {
	db *sql.DB
}

func NewCommandRepository(db *sql.DB) *CommandRepository {
	return &CommandRepository{db: db}
}

// Create inserts a new command
func (r *CommandRepository) Create(cmd *models.DeviceCommand) error {
	query := `
		INSERT INTO device_commands (device_id, command_type, parameters, status, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		cmd.DeviceID,
		cmd.CommandType,
		cmd.Parameters,
		cmd.Status,
		cmd.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	cmd.ID = id
	return nil
}

// GetPendingCommands retrieves all pending commands for a device
func (r *CommandRepository) GetPendingCommands(deviceID string) ([]*models.DeviceCommand, error) {
	query := `
		SELECT id, device_id, command_type, parameters, status, created_at, executed_at, result
		FROM device_commands
		WHERE device_id = ? AND status = 'pending'
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []*models.DeviceCommand
	for rows.Next() {
		var cmd models.DeviceCommand
		var createdAt string
		var executedAt sql.NullString
		var parameters sql.NullString
		var result sql.NullString

		if err := rows.Scan(
			&cmd.ID,
			&cmd.DeviceID,
			&cmd.CommandType,
			&parameters,
			&cmd.Status,
			&createdAt,
			&executedAt,
			&result,
		); err != nil {
			return nil, err
		}

		cmd.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		if parameters.Valid {
			cmd.Parameters = &parameters.String
		}
		if executedAt.Valid {
			t, _ := time.Parse(time.RFC3339, executedAt.String)
			cmd.ExecutedAt = &t
		}
		if result.Valid {
			cmd.Result = &result.String
		}
		commands = append(commands, &cmd)
	}

	return commands, nil
}

// MarkExecuted marks a command as executed
func (r *CommandRepository) MarkExecuted(commandID int64) error {
	query := `
		UPDATE device_commands
		SET status = 'completed', executed_at = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, time.Now().Format(time.RFC3339), commandID)
	return err
}

// MarkFailed marks a command as failed
func (r *CommandRepository) MarkFailed(commandID int64) error {
	query := `
		UPDATE device_commands
		SET status = 'failed', executed_at = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, time.Now().Format(time.RFC3339), commandID)
	return err
}

// UpdateCommandStatus updates command status with result
func (r *CommandRepository) UpdateCommandStatus(commandID int64, status string, result *string) error {
	query := `
		UPDATE device_commands
		SET status = ?, executed_at = ?, result = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, status, time.Now().Format(time.RFC3339), result, commandID)
	return err
}
