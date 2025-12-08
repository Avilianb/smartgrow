package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"irrigation-system/backend/internal/config"
	"irrigation-system/backend/internal/models"
	"irrigation-system/backend/internal/planner"
	"irrigation-system/backend/internal/repository"
	"irrigation-system/backend/internal/weather"
)

// roundToOneDecimal 将浮点数四舍五入到小数点后一位
func roundToOneDecimal(val *float64) *float64 {
	if val == nil {
		return nil
	}
	rounded := math.Round(*val*10) / 10
	return &rounded
}

// Service provides business logic operations
type Service struct {
	cfg           *config.Config
	sensorDataRepo *repository.SensorDataRepository
	forecastRepo   *repository.ForecastRepository
	planRepo       *repository.PlanRepository
	locationRepo   *repository.LocationRepository
	logRepo        *repository.LogRepository
	commandRepo    *repository.CommandRepository
	userRepo       *repository.UserRepository      // 新增：用户仓储
	deviceRepo     *repository.DeviceRepository    // 新增：设备仓储
	weatherClient  *weather.QWeatherClient
	planner        *planner.IrrigationPlanner
}

// NewService creates a new service instance
func NewService(
	cfg *config.Config,
	db *sql.DB,
	weatherClient *weather.QWeatherClient,
) *Service {
	// 初始化用户仓储并创建默认管理员
	userRepo := repository.NewUserRepository(db)
	userRepo.InitializeAdmin() // 初始化管理员账户

	return &Service{
		cfg:            cfg,
		sensorDataRepo: repository.NewSensorDataRepository(db),
		forecastRepo:   repository.NewForecastRepository(db),
		planRepo:       repository.NewPlanRepository(db),
		locationRepo:   repository.NewLocationRepository(db),
		logRepo:        repository.NewLogRepository(db),
		commandRepo:    repository.NewCommandRepository(db),
		userRepo:       userRepo,                             // 新增
		deviceRepo:     repository.NewDeviceRepository(db),  // 新增
		weatherClient:  weatherClient,
		planner: planner.NewIrrigationPlanner(planner.PlannerConfig{
			SoilOptimalMin:      cfg.Planner.SoilOptimalMin,
			SoilOptimalMax:      cfg.Planner.SoilOptimalMax,
			MaxIrrigationPerDay: cfg.Planner.MaxIrrigationPerDay,
			BaseET:              cfg.Planner.BaseET,
			TempFactor:          cfg.Planner.TempFactor,
			RainConversion:      cfg.Planner.RainConversion,
			ADCToMoisture:       cfg.Planner.ADCToMoisture,
			CostW1:              cfg.Planner.CostW1,
			CostW2:              cfg.Planner.CostW2,
			CostW3:              cfg.Planner.CostW3,
		}),
	}
}

// HandleDeviceData processes device data upload and returns commands
func (s *Service) HandleDeviceData(req *models.DeviceDataRequest) (*models.DeviceDataResponse, error) {
	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// 将温度和湿度四舍五入到小数点后一位
	roundedTemp := roundToOneDecimal(req.TemperatureC)
	roundedHumidity := roundToOneDecimal(req.HumidityPct)

	// Store sensor data
	sensorData := &models.SensorData{
		DeviceID:     req.DeviceID,
		Timestamp:    timestamp,
		TemperatureC: roundedTemp,
		HumidityPct:  roundedHumidity,
		SoilRaw:      req.SoilRaw,
		RainAnalog:   req.RainAnalog,
		RainDigital:  req.RainDigital,
		PumpState:    req.PumpState,
		ShadeState:   req.ShadeState,
	}

	if err := s.sensorDataRepo.Create(sensorData); err != nil {
		return nil, fmt.Errorf("failed to store sensor data: %w", err)
	}

	// Log the data reception (使用四舍五入后的值)
	tempValue := 0.0
	if roundedTemp != nil {
		tempValue = *roundedTemp
	}
	humidityValue := 0.0
	if roundedHumidity != nil {
		humidityValue = *roundedHumidity
	}
	soilValue := 0
	if req.SoilRaw != nil {
		soilValue = *req.SoilRaw
	}
	logMessage := fmt.Sprintf("接收设备数据: 温度%.1f°C, 湿度%.1f%%, 土壤%d, 水泵:%s, 遮阳:%s",
		tempValue, humidityValue, soilValue, req.PumpState, req.ShadeState)
	deviceLog := &models.DeviceLog{
		DeviceID:  req.DeviceID,
		Timestamp: timestamp,
		Level:     "INFO",
		Message:   logMessage,
	}
	s.logRepo.Create(deviceLog) // Ignore error for logging

	// Get pending commands
	commands, err := s.commandRepo.GetPendingCommands(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending commands: %w", err)
	}

	// Convert to response format (remove pointers for easier ESP32 handling)
	var commandList []models.DeviceCommand
	for _, cmd := range commands {
		commandList = append(commandList, *cmd)
	}

	return &models.DeviceDataResponse{
		Success:  true,
		Message:  "Data received",
		Commands: commandList,
	}, nil
}

// GetDeviceStatus retrieves current device status
func (s *Service) GetDeviceStatus(deviceID string) (*models.DeviceStatus, error) {
	// Get latest sensor data
	latestData, err := s.sensorDataRepo.GetLatest(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest data: %w", err)
	}

	// Determine soil status
	soilStatus := "optimal"
	if latestData.SoilRaw != nil {
		if *latestData.SoilRaw < s.cfg.Planner.SoilOptimalMin {
			soilStatus = "dry"
		} else if *latestData.SoilRaw > s.cfg.Planner.SoilOptimalMax {
			soilStatus = "wet"
		}
	}

	// Determine rain status
	rainStatus := "no_rain"
	if latestData.RainDigital != nil && *latestData.RainDigital == 0 {
		rainStatus = "raining"
	}

	// Get today's plan
	today := time.Now().Format("2006-01-02")
	todayPlan, err := s.planRepo.GetByDate(deviceID, today)
	plannedVolume := 0.0
	if err == nil && todayPlan != nil {
		plannedVolume = todayPlan.PlannedVolumeL
	}

	// Get executed volume (simplified, should track actual pump runtime)
	executedVolume, _ := s.sensorDataRepo.GetTodayIrrigationVolume(deviceID)

	return &models.DeviceStatus{
		DeviceID:     deviceID,
		Timestamp:    latestData.Timestamp,
		TemperatureC: *latestData.TemperatureC,
		HumidityPct:  *latestData.HumidityPct,
		SoilStatus:   soilStatus,
		RainStatus:   rainStatus,
		PumpState:    latestData.PumpState,
		ShadeState:   latestData.ShadeState,
		TodayPlan: models.TodayPlanInfo{
			PlannedVolumeL:  plannedVolume,
			ExecutedVolumeL: executedVolume,
		},
	}, nil
}

// GetDeviceHistory retrieves historical sensor data
func (s *Service) GetDeviceHistory(deviceID string, startTime, endTime *time.Time, limit, offset int) ([]*models.SensorData, int, error) {
	return s.sensorDataRepo.GetHistory(deviceID, startTime, endTime, limit, offset)
}

// TriggerIrrigation creates a manual irrigation command
func (s *Service) TriggerIrrigation(deviceID string, volumeL float64, reason string) (int64, error) {
	// Create command parameters
	params := map[string]interface{}{
		"volume_l": volumeL,
		"reason":   reason,
	}
	paramsJSON, _ := json.Marshal(params)
	paramsStr := string(paramsJSON)

	cmd := &models.DeviceCommand{
		DeviceID:    deviceID,
		CommandType: "irrigate",
		Parameters:  &paramsStr,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	if err := s.commandRepo.Create(cmd); err != nil {
		return 0, fmt.Errorf("failed to create command: %w", err)
	}

	// Log the action
	logMsg := fmt.Sprintf("Manual irrigation triggered: %.1fL, reason: %s", volumeL, reason)
	s.logRepo.Create(&models.DeviceLog{
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   logMsg,
	})

	return cmd.ID, nil
}

// UpdateForecast fetches and stores weather forecast
func (s *Service) UpdateForecast(latitude, longitude float64) error {
	// Fetch 15-day forecast from QWeather
	weatherData, err := s.weatherClient.Get15DayForecast(latitude, longitude)
	if err != nil {
		return fmt.Errorf("failed to fetch weather data: %w", err)
	}

	// Delete existing future forecasts
	if err := s.forecastRepo.DeleteFutureForecasts(); err != nil {
		return fmt.Errorf("failed to delete old forecasts: %w", err)
	}

	// Convert and store new forecasts
	forecasts := make([]*models.RainForecast, 0, len(weatherData.Daily))
	for _, daily := range weatherData.Daily {
		tempMax, _ := strconv.ParseFloat(daily.TempMax, 64)
		tempMin, _ := strconv.ParseFloat(daily.TempMin, 64)
		precip, _ := strconv.ParseFloat(daily.Precip, 64)
		humidity, _ := strconv.ParseFloat(daily.Humidity, 64)

		rawJSON, _ := json.Marshal(daily)

		forecasts = append(forecasts, &models.RainForecast{
			Date:        daily.FxDate,
			TempMax:     &tempMax,
			TempMin:     &tempMin,
			PrecipMm:    &precip,
			HumidityPct: &humidity,
			RawJSON:     string(rawJSON),
			CreatedAt:   time.Now(),
		})
	}

	if err := s.forecastRepo.CreateBatch(forecasts); err != nil {
		return fmt.Errorf("failed to store forecasts: %w", err)
	}

	return nil
}

// GetForecast returns weather forecast data
func (s *Service) GetForecast(days int) ([]*models.RainForecast, error) {
	return s.forecastRepo.GetForecastDays(days)
}

// RecomputePlan recalculates irrigation plan based on current data
func (s *Service) RecomputePlan(deviceID string) ([]models.IrrigationPlan, error) {
	// Get latest sensor data for initial soil moisture
	latestData, err := s.sensorDataRepo.GetLatest(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest sensor data: %w", err)
	}

	if latestData.SoilRaw == nil {
		return nil, fmt.Errorf("no soil moisture data available")
	}

	// Get 15-day forecast
	forecasts, err := s.forecastRepo.GetForecastDays(15)
	if err != nil || len(forecasts) == 0 {
		return nil, fmt.Errorf("no forecast data available")
	}

	// Convert forecasts to planner format
	plannerForecasts := make([]planner.ForecastDay, len(forecasts))
	for i, f := range forecasts {
		plannerForecasts[i] = planner.ForecastDay{
			Date:     f.Date,
			TempMax:  *f.TempMax,
			TempMin:  *f.TempMin,
			PrecipMm: *f.PrecipMm,
		}
	}

	// Run DP algorithm
	dailyPlans := s.planner.ComputePlan(*latestData.SoilRaw, plannerForecasts)

	// Delete existing future plans
	if err := s.planRepo.DeleteFuturePlans(deviceID); err != nil {
		return nil, fmt.Errorf("failed to delete old plans: %w", err)
	}

	// Convert and store new plans
	irrigationPlans := make([]*models.IrrigationPlan, len(dailyPlans))
	for i, dp := range dailyPlans {
		irrigationPlans[i] = &models.IrrigationPlan{
			DeviceID:       deviceID,
			Date:           dp.Date,
			PlannedVolumeL: dp.PlannedVolumeL,
			CreatedAt:      time.Now(),
		}
	}

	if err := s.planRepo.CreateBatch(irrigationPlans); err != nil {
		return nil, fmt.Errorf("failed to store plans: %w", err)
	}

	// Convert to response format
	result := make([]models.IrrigationPlan, len(irrigationPlans))
	for i, p := range irrigationPlans {
		result[i] = *p
	}

	// Log the recomputation
	s.logRepo.Create(&models.DeviceLog{
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   fmt.Sprintf("Irrigation plan recomputed for %d days", len(result)),
	})

	return result, nil
}

// GetLocation retrieves device location
func (s *Service) GetLocation(deviceID string) (*models.DeviceLocation, error) {
	return s.locationRepo.Get(deviceID)
}

// UpdateLocation updates device location
func (s *Service) UpdateLocation(deviceID string, latitude, longitude float64, address *string) error {
	loc := &models.DeviceLocation{
		DeviceID:  deviceID,
		Latitude:  latitude,
		Longitude: longitude,
		Address:   address,
		UpdatedAt: time.Now(),
	}

	return s.locationRepo.Upsert(loc)
}

// GetLogs retrieves device logs with filters
func (s *Service) GetLogs(deviceID string, level *string, startTime *time.Time, limit, offset int) ([]*models.DeviceLog, int, error) {
	return s.logRepo.Query(deviceID, level, startTime, limit, offset)
}

// UpdateCommandStatus updates command execution status
func (s *Service) UpdateCommandStatus(commandID int64, status string, result *string) error {
	return s.commandRepo.UpdateCommandStatus(commandID, status, result)
}

// ========== 用户认证相关服务方法 ==========

// Login 用户登录
func (s *Service) Login(username, password string) (*models.User, string, error) {
	// Trim空格
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	// 获取用户
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, "", fmt.Errorf("用户名或密码错误")
	}

	// 验证密码
	if !s.userRepo.VerifyPassword(user.PasswordHash, password) {
		return nil, "", fmt.Errorf("用户名或密码错误")
	}

	// 如果是普通用户，需要获取设备ID
	deviceID := ""
	if user.Role == "user" {
		device, err := s.deviceRepo.GetDeviceByUserID(user.ID)
		if err != nil {
			return nil, "", fmt.Errorf("该用户没有关联设备")
		}
		deviceID = device.DeviceID
	}

	// 生成token (需要在handler中调用middleware.GenerateToken)
	return user, deviceID, nil
}

// CreateUser 创建新用户（仅管理员）
func (s *Service) CreateUser(req *models.CreateUserRequest) (*models.UserWithDevice, error) {
	// 检查设备是否已存在
	existingDevice, _ := s.deviceRepo.GetDeviceByDeviceID(req.DeviceID)
	if existingDevice != nil {
		return nil, fmt.Errorf("设备ID已被使用")
	}

	// 创建用户
	user, err := s.userRepo.CreateUser(req.Username, req.Password)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 创建设备
	device, err := s.deviceRepo.CreateDevice(req.DeviceID, req.DeviceName, user.ID)
	if err != nil {
		// 如果设备创建失败，删除已创建的用户
		s.userRepo.DeleteUser(user.ID)
		return nil, fmt.Errorf("创建设备失败: %w", err)
	}

	// 返回用户和设备信息
	return &models.UserWithDevice{
		ID:         user.ID,
		Username:   user.Username,
		Role:       user.Role,
		DeviceID:   &device.DeviceID,
		DeviceName: &device.DeviceName,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

// GetAllUsers 获取所有用户（仅管理员）
func (s *Service) GetAllUsers() ([]models.UserWithDevice, error) {
	return s.userRepo.GetAllUsersWithDevices()
}

// DeleteUser 删除用户（仅管理员）
func (s *Service) DeleteUser(userID int64) error {
	// 先获取用户的设备
	device, err := s.deviceRepo.GetDeviceByUserID(userID)
	if err == nil && device != nil {
		// 删除设备
		s.deviceRepo.DeleteDevice(device.DeviceID)
	}

	// 删除用户
	return s.userRepo.DeleteUser(userID)
}

// UpdateUserPassword 更新用户密码
func (s *Service) UpdateUserPassword(userID int64, newPassword string) error {
	return s.userRepo.UpdateUserPassword(userID, newPassword)
}

// ChangePassword 修改当前用户密码（需要验证旧密码）
func (s *Service) ChangePassword(userID int64, oldPassword, newPassword string) error {
	// 获取用户
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证旧密码
	if !s.userRepo.VerifyPassword(user.PasswordHash, oldPassword) {
		return fmt.Errorf("当前密码错误")
	}

	// 更新密码
	return s.userRepo.UpdateUserPassword(userID, newPassword)
}

// UpdateDeviceName 更新设备名称
func (s *Service) UpdateDeviceName(deviceID, deviceName string) error {
	return s.deviceRepo.UpdateDeviceName(deviceID, deviceName)
}

// GetUserDevice 获取用户的设备
func (s *Service) GetUserDevice(userID int64) (*models.Device, error) {
	return s.deviceRepo.GetDeviceByUserID(userID)
}

