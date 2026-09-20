package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"taller-gestion/backend/src/config"
	"taller-gestion/backend/src/controllers"
	"taller-gestion/backend/src/db"
	"taller-gestion/backend/src/middleware"
	"taller-gestion/backend/src/repositories"
	"taller-gestion/backend/src/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	bunDB := db.Connect(cfg.DatabaseURL)
	defer bunDB.Close()

	if err := waitForDB(ctx, bunDB); err != nil {
		log.Fatalf("database: %v", err)
	}

	if cfg.MigrateOnly {
		if err := db.Migrate(ctx, bunDB); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		fmt.Println("Migrations applied.")
		return
	}

	if cfg.ForceSeed {
		if err := db.Migrate(ctx, bunDB); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		if err := db.ForceSeed(ctx, bunDB); err != nil {
			log.Fatalf("force-seed: %v", err)
		}
		fmt.Println("Demo data reloaded (force-seed).")
		return
	}

	if cfg.SeedOnly {
		if err := db.Migrate(ctx, bunDB); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		if err := db.SeedIfEmpty(ctx, bunDB); err != nil {
			log.Fatalf("seed: %v", err)
		}
		fmt.Println("Seed complete (or already present).")
		return
	}

	if err := db.Migrate(ctx, bunDB); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if cfg.RunSeed {
		if err := db.SeedIfEmpty(ctx, bunDB); err != nil {
			log.Fatalf("seed: %v", err)
		}
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(corsMiddleware(cfg.CORSOrigin))
	r.Use(middleware.Auth(cfg.JWTSecret))

	repos := repositories.New(bunDB)
	svcs := services.New(repos, cfg.JWTSecret)
	controllers.Register(r, controllers.New(svcs))

	addr := ":" + cfg.Port
	fmt.Printf("Instal Service S.A. — gestión de taller (Go) en http://localhost:%s\n", cfg.Port)
	fmt.Println("Usuarios demo: admin/admin123  coord/coord123  diego/diego123  sofia/sofia123")
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func waitForDB(ctx context.Context, bunDB interface{ PingContext(context.Context) error }) error {
	var last error
	for i := 0; i < 30; i++ {
		if err := bunDB.PingContext(ctx); err == nil {
			return nil
		} else {
			last = err
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("postgres not ready: %w", last)
}

func corsMiddleware(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if origin == "" || origin == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
