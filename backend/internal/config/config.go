package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Weather  WeatherConfig  `yaml:"weather"`
	Planner  PlannerConfig  `yaml:"planner"`
	Logging  LoggingConfig  `yaml:"logging"`
	Security SecurityConfig `yaml:"security"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Path    string `yaml:"path"`
	DevPath string `yaml:"dev_path"`
}

type WeatherConfig struct {
	APIHost         string              `yaml:"api_host"`
	APIKey          string              `yaml:"api_key"`
	Endpoint        string              `yaml:"endpoint"`
	Timeout         time.Duration       `yaml:"timeout"`
	DefaultLocation DefaultLocationInfo `yaml:"default_location"`
}

type DefaultLocationInfo struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
}

type PlannerConfig struct {
	SoilOptimalMin      int     `yaml:"soil_optimal_min"`
	SoilOptimalMax      int     `yaml:"soil_optimal_max"`
	MaxIrrigationPerDay float64 `yaml:"max_irrigation_per_day"`
	BaseET              float64 `yaml:"base_et"`
	TempFactor          float64 `yaml:"temp_factor"`
	RainConversion      float64 `yaml:"rain_conversion"`
	ADCToMoisture       float64 `yaml:"adc_to_moisture"`
	CostW1              float64 `yaml:"cost_w1"`
	CostW2              float64 `yaml:"cost_w2"`
	CostW3              float64 `yaml:"cost_w3"`
}

type LoggingConfig struct {
	Level   string `yaml:"level"`
	File    string `yaml:"file"`
	DevFile string `yaml:"dev_file"`
}

type SecurityConfig struct {
	JWTSecret          string   `yaml:"jwt_secret"`
	JWTExpireHours     int      `yaml:"jwt_expire_hours"`
	AllowedOrigins     []string `yaml:"allowed_origins"`
	RateLimitPerMinute int      `yaml:"rate_limit_per_minute"`
	DeviceAPIKey       string   `yaml:"device_api_key"`
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 从环境变量覆盖敏感配置
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.Security.JWTSecret = jwtSecret
	}
	if deviceAPIKey := os.Getenv("DEVICE_API_KEY"); deviceAPIKey != "" {
		cfg.Security.DeviceAPIKey = deviceAPIKey
	}

	// 验证必要的安全配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Security.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required (set via environment variable or config file)")
	}
	if len(c.Security.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.Security.JWTExpireHours <= 0 {
		c.Security.JWTExpireHours = 24 // 默认24小时
	}
	if c.Security.RateLimitPerMinute <= 0 {
		c.Security.RateLimitPerMinute = 10 // 默认每分钟10次
	}
	return nil
}

// GetDBPath returns the appropriate database path based on mode
func (c *Config) GetDBPath() string {
	if c.Server.Mode == "debug" || c.Server.Mode == "development" {
		return c.Database.DevPath
	}
	return c.Database.Path
}

// GetLogPath returns the appropriate log path based on mode
func (c *Config) GetLogPath() string {
	if c.Server.Mode == "debug" || c.Server.Mode == "development" {
		return c.Logging.DevFile
	}
	return c.Logging.File
}
