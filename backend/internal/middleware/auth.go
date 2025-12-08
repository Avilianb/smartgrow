package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"irrigation-system/backend/internal/models"
)

var jwtSecret []byte
var jwtExpireHours int

// InitAuth initializes the authentication middleware with config
func InitAuth(secret string, expireHours int) {
	jwtSecret = []byte(secret)
	jwtExpireHours = expireHours
}

// 简化的JWT实现（不依赖外部库）
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	DeviceID string `json:"device_id,omitempty"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

// GenerateToken 生成 JWT token
func GenerateToken(user *models.User, deviceID string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", fmt.Errorf("JWT secret not initialized")
	}

	expireDuration := time.Duration(jwtExpireHours) * time.Hour
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		DeviceID: deviceID,
		Exp:      time.Now().Add(expireDuration).Unix(),
		Iat:      time.Now().Unix(),
	}

	// 创建header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	message := headerB64 + "." + claimsB64

	// 创建签名
	h := hmac.New(sha256.New, jwtSecret)
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature, nil
}

// ParseToken 解析token
func ParseToken(tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// 验证签名
	message := parts[0] + "." + parts[1]
	h := hmac.New(sha256.New, jwtSecret)
	h.Write([]byte(message))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if parts[2] != expectedSignature {
		return nil, fmt.Errorf("invalid signature")
	}

	// 解析claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, err
	}

	// 检查过期时间
	if claims.Exp < time.Now().Unix() {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// AuthRequired 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 提取 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析 token
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "认证令牌无效或已过期: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 提取用户信息
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		if claims.DeviceID != "" {
			c.Set("device_id", claims.DeviceID)
		}

		c.Next()
	}
}

// AdminRequired 管理员权限中间件
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "需要管理员权限",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// DeviceAccessCheck 检查用户是否有权限访问该设备
func DeviceAccessCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")

		// 管理员可以访问所有设备
		if role == "admin" {
			c.Next()
			return
		}

		// 普通用户只能访问自己的设备
		deviceID := c.Param("device_id")
		userDeviceID, exists := c.Get("device_id")

		if !exists || deviceID != userDeviceID.(string) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "无权访问该设备",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// DeviceAPIAuth 设备API认证中间件（用于ESP32上报数据）
var deviceAPIKey string

// InitDeviceAuth 初始化设备API认证
func InitDeviceAuth(apiKey string) {
	deviceAPIKey = apiKey
}

// DeviceAPIAuthMiddleware 设备API认证中间件
func DeviceAPIAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-Device-API-Key")
		deviceID := c.GetHeader("X-Device-ID")

		if deviceID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "缺少设备ID",
			})
			c.Abort()
			return
		}

		if apiKey == "" || apiKey != deviceAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "设备认证失败",
			})
			c.Abort()
			return
		}

		c.Set("device_id", deviceID)
		c.Next()
	}
}
