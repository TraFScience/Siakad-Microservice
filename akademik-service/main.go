package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "time"

"akademik-service/internal/config"
"akademik-service/internal/siakad"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    cfg := config.Load()

    db, err := sql.Open("pgx", cfg.DSN())
    if err != nil {
        log.Fatalf("gagal membuka koneksi database: %v", err)
    }
    defer db.Close()

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    ctx, cancel := timeout(10 * time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        log.Fatalf("gagal ping database: %v", err)
    }
    log.Println("koneksi database berhasil")

    mhsRepo := siakad.NewMahasiswaRepository(db)
    nRepo := siakad.NewNilaiRepository(db)
    svc := siakad.NewService(db, mhsRepo, nRepo)
    handler := siakad.NewHandler(svc)

    router := gin.Default()

    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    api := router.Group("/api/v1")
    api.GET("/rekap", func(c *gin.Context) {
        c.JSON(http.StatusNotFound, gin.H{"sukses": false, "error": "route telah dipindahkan ke rekap-service"})
    })
    api.GET("/rekap/*any", func(c *gin.Context) {
        c.JSON(http.StatusNotFound, gin.H{"sukses": false, "error": "route telah dipindahkan ke rekap-service"})
    })
    handler.Register(api)

    addr := fmt.Sprintf(":%s", cfg.AppPort)
    log.Printf("server berjalan di %s", addr)
    if err := router.Run(addr); err != nil {
        log.Fatalf("gagal menjalankan server: %v", err)
    }
}

func timeout(d time.Duration) (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), d)
}
