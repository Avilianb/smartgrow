package models

import "time"

// SensorData represents a sensor data record
type SensorData struct {
	ID           int64     `json:"id"`
	DeviceID     string    `json:"device_id"`
	Timestamp    time.Time `json:"timestamp"`
	TemperatureC *float64  `json:"temperature_c"`
	HumidityPct  *float64  `json:"humidity_pct"`
	SoilRaw      *int      `json:"soil_raw"`
	RainAnalog   *int      `json:"rain_analog"`
	RainDigital  *int      `json:"rain_digital"`
	PumpState    string    `json:"pump_state"`
	ShadeState   string    `json:"shade_state"`
}

// RainForecast represents a weather forecast record
type RainForecast struct {
	ID          int64     `json:"id"`
	Date        string    `json:"date"`
	TempMax     *float64  `json:"temp_max"`
	TempMin     *float64  `json:"temp_min"`
	PrecipMm    *float64  `json:"precip_mm"`
	HumidityPct *float64  `json:"humidity_pct"`
	RawJSON     string    `json:"raw_json"`
	CreatedAt   time.Time `json:"created_at"`
}

// IrrigationPlan represents an irrigation plan record
type IrrigationPlan struct {
	ID              int64     `json:"id"`
	DeviceID        string    `json:"device_id"`
	Date            string    `json:"date"`
	PlannedVolumeL  float64   `json:"planned_volume_l"`
	CreatedAt       time.Time `json:"created_at"`
	ExecutedVolumeL float64   `json:"executed_volume_l,omitempty"` // 已执行水量
}

// DeviceLog represents a device log record
type DeviceLog struct {
	ID        int64     `json:"id"`
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // INFO, WARN, ERROR
	Message   string    `json:"message"`
	Extra     *string   `json:"extra,omitempty"`
}

// DeviceLocation represents device location information
type DeviceLocation struct {
	ID        int64     `json:"id"`
	DeviceID  string    `json:"device_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Address   *string   `json:"address,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeviceCommand represents a command to be executed by device
type DeviceCommand struct {
	ID          int64      `json:"id"`
	DeviceID    string     `json:"device_id"`
	CommandType string     `json:"command_type"` // irrigate, toggle_shade, update_config
	Parameters  *string    `json:"parameters,omitempty"`
	Status      string     `json:"status"` // pending, executing, completed, failed
	CreatedAt   time.Time  `json:"created_at"`
	ExecutedAt  *time.Time `json:"executed_at,omitempty"`
	Result      *string    `json:"result,omitempty"` // 执行结果或错误信息
}

// DeviceStatus represents the current device status (for API response)
type DeviceStatus struct {
	DeviceID     string             `json:"device_id"`
	Timestamp    time.Time          `json:"timestamp"`
	TemperatureC float64            `json:"temperature_c"`
	HumidityPct  float64            `json:"humidity_pct"`
	SoilStatus   string             `json:"soil_status"` // dry, optimal, wet
	RainStatus   string             `json:"rain_status"` // raining, no_rain
	PumpState    string             `json:"pump_state"`  // on, off
	ShadeState   string             `json:"shade_state"` // open, closed
	TodayPlan    TodayPlanInfo      `json:"today_plan"`
}

// TodayPlanInfo contains today's irrigation plan information
type TodayPlanInfo struct {
	PlannedVolumeL  float64 `json:"planned_volume_l"`
	ExecutedVolumeL float64 `json:"executed_volume_l"`
}

// DeviceDataRequest represents the device data upload request
type DeviceDataRequest struct {
	DeviceID     string   `json:"device_id" binding:"required"`
	Timestamp    string   `json:"timestamp" binding:"required"`
	TemperatureC *float64 `json:"temperature_c"`
	HumidityPct  *float64 `json:"humidity_pct"`
	SoilRaw      *int     `json:"soil_raw"`
	RainAnalog   *int     `json:"rain_analog"`
	RainDigital  *int     `json:"rain_digital"`
	PumpState    string   `json:"pump_state"`
	ShadeState   string   `json:"shade_state"`
}

// DeviceDataResponse represents the response to device data upload
type DeviceDataResponse struct {
	Success  bool              `json:"success"`
	Message  string            `json:"message,omitempty"`
	Commands []DeviceCommand   `json:"commands,omitempty"` // 待执行的命令列表
}

// CommandExecutionRequest represents ESP32 reporting command execution status
type CommandExecutionRequest struct {
	CommandID int64  `json:"command_id" binding:"required"`
	Status    string `json:"status" binding:"required"` // executing, completed, failed
	Result    string `json:"result"`                    // 执行结果或错误信息
}

// IrrigateRequest represents a manual irrigation request
type IrrigateRequest struct {
	VolumeL float64 `json:"volume_l" binding:"required,gt=0,lte=5"`
	Reason  string  `json:"reason"`
}

// UpdateLocationRequest represents a location update request
type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Address   *string `json:"address"`
}

// UpdateForecastRequest represents a forecast update request
type UpdateForecastRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// DailyPlan represents a single day's irrigation plan (for API response)
type DailyPlan struct {
	Date           string  `json:"date"`
	PlannedVolumeL float64 `json:"planned_volume_l"`
}

// ========== 用户认证相关模型 ==========

// User represents a user account
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // 不返回密码哈希
	Role         string    `json:"role"` // admin, user
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Device represents a device
type Device struct {
	ID         int64     `json:"id"`
	DeviceID   string    `json:"device_id"`
	UserID     *int64    `json:"user_id,omitempty"`
	DeviceName string    `json:"device_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
	User    *User  `json:"user,omitempty"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=20"`
	Password   string `json:"password" binding:"required,min=6"`
	DeviceID   string `json:"device_id" binding:"required"`
	DeviceName string `json:"device_name" binding:"required"`
}

// UpdateUserRequest represents a request to update user info
type UpdateUserRequest struct {
	Password   string `json:"password,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}

// ChangePasswordRequest represents a request to change password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UserWithDevice represents a user with their device info
type UserWithDevice struct {
	ID         int64     `json:"id"`
	Username   string    `json:"username"`
	Role       string    `json:"role"`
	DeviceID   *string   `json:"device_id,omitempty"`
	DeviceName *string   `json:"device_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	DeviceID string `json:"device_id,omitempty"`
}
