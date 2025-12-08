package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"irrigation-system/backend/internal/config"
	"irrigation-system/backend/internal/database"
	"irrigation-system/backend/internal/handler"
	"irrigation-system/backend/internal/middleware"
	"irrigation-system/backend/internal/service"
	"irrigation-system/backend/internal/weather"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database
	dbPath := cfg.GetDBPath()
	log.Printf("Opening database at: %s", dbPath)
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	schemaPath := filepath.Join("configs", "schema.sql")
	if _, err := os.Stat(schemaPath); err == nil {
		log.Println("Initializing database schema...")
		if err := db.InitSchema(schemaPath); err != nil {
			log.Fatalf("Failed to initialize schema: %v", err)
		}
		log.Println("Database schema initialized successfully")
	}

	// Initialize weather client
	weatherClient := weather.NewQWeatherClient(
		cfg.Weather.APIHost,
		cfg.Weather.APIKey,
		cfg.Weather.Timeout,
	)

	// Initialize service
	svc := service.NewService(cfg, db.DB, weatherClient)

	// Initialize JWT auth with config
	middleware.InitAuth(cfg.Security.JWTSecret, cfg.Security.JWTExpireHours)
	log.Printf("JWT auth initialized (expire: %dh)", cfg.Security.JWTExpireHours)

	// Initialize device API auth
	middleware.InitDeviceAuth(cfg.Security.DeviceAPIKey)
	log.Printf("Device API auth initialized")

	// Initialize handler
	h := handler.NewHandler(svc)

	// Setup Gin router
	r := gin.Default()

	// Add CORS middleware with restricted origins
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否是允许的源
		allowed := false
		for _, allowedOrigin := range cfg.Security.AllowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Add security headers
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// CSP header
		c.Writer.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")
		c.Next()
	})

	// Add audit logging
	auditLogger := middleware.NewAuditLogger()
	r.Use(auditLogger.AuditMiddleware())

	// Setup routes
	h.SetupRoutes(r)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "irrigation-system",
		})
	})

	// 服务前端静态文件
	// 优先检查 ./dist 目录（部署时），如果不存在则使用 ./frontend/dist（开发时）
	staticDir := "./dist"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = "./frontend/dist"
	}

	if _, err := os.Stat(staticDir); err == nil {
		log.Printf("Serving static files from: %s", staticDir)

		// 服务静态资源文件（CSS, JS, 图片等）
		r.Static("/assets", filepath.Join(staticDir, "assets"))

		// 服务其他静态文件
		r.StaticFile("/vite.svg", filepath.Join(staticDir, "vite.svg"))
		r.StaticFile("/diagnostic.html", filepath.Join(staticDir, "diagnostic.html"))

		// 所有未匹配的路由返回 index.html（用于 React Router）
		r.NoRoute(func(c *gin.Context) {
			// 如果是 API 请求，返回 404
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(404, gin.H{"error": "API endpoint not found"})
				return
			}
			// 其他请求返回前端首页
			c.File(filepath.Join(staticDir, "index.html"))
		})
	} else {
		log.Printf("Warning: Static files directory not found at %s", staticDir)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Mode: %s", cfg.Server.Mode)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
