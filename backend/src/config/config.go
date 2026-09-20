package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	CORSOrigin  string
	RunSeed     bool
	MigrateOnly bool
	SeedOnly    bool
	ForceSeed   bool
}

func Load() Config {
	loadDotEnv()

	migrateOnly := flag.Bool("migrate-only", false, "apply migrations and exit")
	seedOnly := flag.Bool("seed-only", false, "seed demo data if empty and exit")
	forceSeed := flag.Bool("force-seed", false, "wipe app tables and reload demo data, then exit")
	flag.Parse()

	runSeed := true
	if v := os.Getenv("RUN_SEED"); v != "" {
		runSeed, _ = strconv.ParseBool(v)
	}

	return Config{
		DatabaseURL: envOr("DATABASE_URL", "postgres://taller:taller@127.0.0.1:5432/taller_gestion?sslmode=disable"),
		JWTSecret:   envOr("JWT_SECRET", "taller-gestion-local-dev"),
		Port:        envOr("PORT", "3847"),
		CORSOrigin:  envOr("CORS_ORIGIN", "*"),
		RunSeed:     runSeed,
		MigrateOnly: *migrateOnly,
		SeedOnly:    *seedOnly,
		ForceSeed:   *forceSeed,
	}
}

func loadDotEnv() {
	candidates := []string{".env"}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "..", ".env"))
		candidates = append(candidates, filepath.Join(wd, ".env"))
	}
	candidates = append(candidates, filepath.Join("..", ".env"))

	seen := map[string]bool{}
	for _, p := range candidates {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if seen[abs] {
			continue
		}
		seen[abs] = true
		_ = godotenv.Load(abs)
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
