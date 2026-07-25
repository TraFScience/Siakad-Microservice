package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"akademik-service/internal/config"
	"akademik-service/internal/response"
	"akademik-service/internal/siakad"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()

	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		slog.Error("gagal membuka koneksi database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		slog.Error("gagal ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("koneksi database berhasil")

	mhsRepo := siakad.NewMahasiswaRepository(db)
	nRepo := siakad.NewNilaiRepository(db)
	svc := siakad.NewService(db, mhsRepo, nRepo)
	handler := siakad.NewHandler(svc)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestIDMiddleware())
	router.Use(slogMiddleware())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			slog.Error("health check gagal", "error", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	api.GET("/rekap", func(c *gin.Context) {
		response.Error(c, http.StatusNotFound, fmt.Errorf("route telah dipindahkan ke rekap-service"))
	})
	api.GET("/rekap/*any", func(c *gin.Context) {
		response.Error(c, http.StatusNotFound, fmt.Errorf("route telah dipindahkan ke rekap-service"))
	})
	handler.Register(api)

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("server berjalan", "port", cfg.Port)
	if err := router.Run(addr); err != nil {
		slog.Error("gagal menjalankan server", "error", err)
		os.Exit(1)
	}
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			b := make([]byte, 16)
			rand.Read(b)
			requestID = fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func slogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		requestID, _ := c.Get("request_id")

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		slog.Info("request",
			"service_name", "akademik-service",
			"request_id", requestID,
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
		)
	}
}
