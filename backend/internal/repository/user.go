package repository

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"irrigation-system/backend/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser 创建新用户
func (r *UserRepository) CreateUser(username, password string) (*models.User, error) {
	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().Format(time.RFC3339)
	query := `
		INSERT INTO users (username, password_hash, role, created_at, updated_at)
		VALUES (?, ?, 'user', ?, ?)
	`

	result, err := r.db.Exec(query, username, string(passwordHash), now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user id: %w", err)
	}

	return r.GetUserByID(id)
}

// GetUserByID 根据ID获取用户
func (r *UserRepository) GetUserByID(id int64) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	var user models.User
	var createdAt, updatedAt string

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = ?
	`

	var user models.User
	var createdAt, updatedAt string

	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &user, nil
}

// VerifyPassword 验证密码
func (r *UserRepository) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetAllUsersWithDevices 获取所有用户及其设备信息（仅管理员）
func (r *UserRepository) GetAllUsersWithDevices() ([]models.UserWithDevice, error) {
	query := `
		SELECT
			u.id,
			u.username,
			u.role,
			d.device_id,
			d.device_name,
			u.created_at,
			u.updated_at
		FROM users u
		LEFT JOIN devices d ON u.id = d.user_id
		WHERE u.role = 'user'
		ORDER BY u.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.UserWithDevice
	for rows.Next() {
		var user models.UserWithDevice
		var createdAt, updatedAt string
		var deviceID, deviceName sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Role,
			&deviceID,
			&deviceName,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, err
		}

		if deviceID.Valid {
			user.DeviceID = &deviceID.String
		}
		if deviceName.Valid {
			user.DeviceName = &deviceName.String
		}

		user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		users = append(users, user)
	}

	return users, nil
}

// DeleteUser 删除用户
func (r *UserRepository) DeleteUser(userID int64) error {
	query := `DELETE FROM users WHERE id = ? AND role = 'user'` // 不能删除管理员
	result, err := r.db.Exec(query, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("user not found or cannot delete admin")
	}

	return nil
}

// UpdateUserPassword 更新用户密码
func (r *UserRepository) UpdateUserPassword(userID int64, newPassword string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().Format(time.RFC3339)
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`

	_, err = r.db.Exec(query, string(passwordHash), now, userID)
	return err
}

// InitializeAdmin 初始化管理员账户
func (r *UserRepository) InitializeAdmin() error {
	// 检查是否已存在管理员
	query := `SELECT COUNT(*) FROM users WHERE role = 'admin'`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // 管理员已存在
	}

	// 创建默认管理员 (密码: admin123)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	insertQuery := `
		INSERT INTO users (username, password_hash, role, created_at, updated_at)
		VALUES ('admin', ?, 'admin', ?, ?)
	`

	_, err = r.db.Exec(insertQuery, string(passwordHash), now, now)
	return err
}
