package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"irrigation-system/backend/internal/middleware"
	"irrigation-system/backend/internal/models"
	"irrigation-system/backend/internal/service"
)

// Handler handles HTTP requests
type Handler struct {
	service          *service.Service
	loginRateLimiter *middleware.LoginRateLimiter
}

// NewHandler creates a new handler instance
func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		service:          svc,
		loginRateLimiter: middleware.NewLoginRateLimiter(),
	}
}

// SetupRoutes configures all API routes
func (h *Handler) SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 认证API（添加登录速率限制）
		auth := api.Group("/auth")
		auth.Use(h.loginRateLimiter.LoginRateLimitMiddleware())
		{
			auth.POST("/login", h.Login)              // 普通用户登录
			auth.POST("/admin/login", h.AdminLogin) // 管理员登录
		}

		// 设备数据API（ESP32上报数据需要设备API认证）
		device := api.Group("/device")
		device.Use(middleware.DeviceAPIAuthMiddleware())
		{
			device.POST("/data", h.HandleDeviceData)
			device.POST("/command/status", h.UpdateCommandStatus)
		}

		// 需要认证的API
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// 设备API（需要设备访问权限检查）
			protected.GET("/device/:device_id/status", middleware.DeviceAccessCheck(), h.GetDeviceStatus)
			protected.GET("/device/:device_id/history", middleware.DeviceAccessCheck(), h.GetDeviceHistory)
			protected.POST("/device/:device_id/irrigate", middleware.DeviceAccessCheck(), h.TriggerIrrigation)
			protected.GET("/device/:device_id/logs", middleware.DeviceAccessCheck(), h.GetLogs)

			// 位置API
			protected.GET("/location/:device_id", middleware.DeviceAccessCheck(), h.GetLocation)
			protected.POST("/location/:device_id", middleware.DeviceAccessCheck(), h.UpdateLocation)

			// 天气和计划API
			protected.POST("/forecast/update", h.UpdateForecast)
			protected.GET("/forecast", h.GetForecast)
			protected.POST("/plan/recompute", h.RecomputePlan)

			// 管理员专用API
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminRequired())
			{
				admin.GET("/users", h.GetAllUsers)           // 获取所有用户
				admin.POST("/users", h.CreateUser)           // 创建用户
				admin.DELETE("/users/:user_id", h.DeleteUser) // 删除用户
			}

			// 用户个人操作（所有登录用户可用）
			protected.POST("/user/change-password", h.ChangePassword) // 修改密码
		}
	}
}

// HandleDeviceData handles device data upload
func (h *Handler) HandleDeviceData(c *gin.Context) {
	var req models.DeviceDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request format: " + err.Error(),
		})
		return
	}

	resp, err := h.service.HandleDeviceData(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to process device data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDeviceStatus retrieves device status
func (h *Handler) GetDeviceStatus(c *gin.Context) {
	deviceID := c.Param("device_id")

	status, err := h.service.GetDeviceStatus(deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get device status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetDeviceHistory retrieves historical data
func (h *Handler) GetDeviceHistory(c *gin.Context) {
	deviceID := c.Param("device_id")

	// Parse query parameters
	var startTime, endTime *time.Time
	if startStr := c.Query("start_time"); startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			startTime = &t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			endTime = &t
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	data, total, err := h.service.GetDeviceHistory(deviceID, startTime, endTime, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get history: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  data,
		"total": total,
	})
}

// TriggerIrrigation handles manual irrigation request
func (h *Handler) TriggerIrrigation(c *gin.Context) {
	deviceID := c.Param("device_id")

	var req models.IrrigateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	commandID, err := h.service.TriggerIrrigation(deviceID, req.VolumeL, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to trigger irrigation: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"command_id": strconv.FormatInt(commandID, 10),
	})
}

// GetLogs retrieves device logs
func (h *Handler) GetLogs(c *gin.Context) {
	deviceID := c.Param("device_id")

	// Parse query parameters
	level := c.Query("level")
	var levelPtr *string
	if level != "" {
		levelPtr = &level
	}

	var startTime *time.Time
	if startStr := c.Query("start_time"); startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			startTime = &t
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, total, err := h.service.GetLogs(deviceID, levelPtr, startTime, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get logs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
	})
}

// GetLocation retrieves device location
func (h *Handler) GetLocation(c *gin.Context) {
	deviceID := c.Param("device_id")

	location, err := h.service.GetLocation(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Location not found: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, location)
}

// UpdateLocation updates device location
func (h *Handler) UpdateLocation(c *gin.Context) {
	deviceID := c.Param("device_id")

	var req models.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateLocation(deviceID, req.Latitude, req.Longitude, req.Address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update location: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// UpdateForecast fetches and stores weather forecast
func (h *Handler) UpdateForecast(c *gin.Context) {
	var req models.UpdateForecastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateForecast(req.Latitude, req.Longitude); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update forecast: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"updated_days": 15,
	})
}

// GetForecast retrieves weather forecast data
func (h *Handler) GetForecast(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "5"))

	forecasts, err := h.service.GetForecast(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get forecast: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    forecasts,
	})
}

// RecomputePlan recalculates irrigation plan
func (h *Handler) RecomputePlan(c *gin.Context) {
	deviceID := c.Query("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "device_id is required",
		})
		return
	}

	plans, err := h.service.RecomputePlan(deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to recompute plan: " + err.Error(),
		})
		return
	}

	// Convert to simple format
	simplePlans := make([]models.DailyPlan, len(plans))
	for i, p := range plans {
		simplePlans[i] = models.DailyPlan{
			Date:           p.Date,
			PlannedVolumeL: p.PlannedVolumeL,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plan":    simplePlans,
	})
}

// UpdateCommandStatus handles ESP32 reporting command execution status
func (h *Handler) UpdateCommandStatus(c *gin.Context) {
	var req models.CommandExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateCommandStatus(req.CommandID, req.Status, &req.Result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update command status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// ========== 用户认证处理器 ==========

// Login 普通用户登录
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求格式错误: " + err.Error(),
		})
		return
	}

	// 调用服务层登录
	user, deviceID, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 只允许普通用户从此接口登录
	if user.Role != "user" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "请使用管理员登录入口",
		})
		return
	}

	// 生成token
	token, err := middleware.GenerateToken(user, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Token生成失败",
		})
		return
	}

	// 登录成功，重置速率限制
	h.loginRateLimiter.ResetAttempts(c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user":    user,
	})
}

// AdminLogin 管理员登录
func (h *Handler) AdminLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AdminLogin] 请求格式错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求格式错误: " + err.Error(),
		})
		return
	}

	log.Printf("[AdminLogin] 尝试登录: username=%s", req.Username)

	// 调用服务层登录
	user, deviceID, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		log.Printf("[AdminLogin] 登录失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	log.Printf("[AdminLogin] 登录成功: username=%s, role=%s", user.Username, user.Role)

	// 只允许管理员从此接口登录
	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "您不是管理员",
		})
		return
	}

	// 生成token
	token, err := middleware.GenerateToken(user, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Token生成失败",
		})
		return
	}

	// 登录成功，重置速率限制
	h.loginRateLimiter.ResetAttempts(c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user":    user,
	})
}

// ========== 用户管理处理器（管理员专用） ==========

// GetAllUsers 获取所有用户
func (h *Handler) GetAllUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"users":   users,
	})
}

// CreateUser 创建新用户
func (h *Handler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求格式错误: " + err.Error(),
		})
		return
	}

	user, err := h.service.CreateUser(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户创建成功",
		"user":    user,
	})
}

// DeleteUser 删除用户
func (h *Handler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的用户ID",
		})
		return
	}

	if err := h.service.DeleteUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除用户失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户删除成功",
	})
}

// ChangePassword 修改当前用户密码
func (h *Handler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求格式错误: " + err.Error(),
		})
		return
	}

	// 从token中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}

	// 调用service修改密码
	err := h.service.ChangePassword(userID.(int64), req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "密码修改成功",
	})
}

