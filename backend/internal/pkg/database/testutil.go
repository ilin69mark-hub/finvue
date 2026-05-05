package database

import (
	"os"

	"finvue/internal/pkg/config"
	"finvue/internal/pkg/logger"
)

func InitForTests() error {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "finvue")
	os.Setenv("DB_PASSWORD", "finvue_secret")
	os.Setenv("DB_NAME", "finvue")
	os.Setenv("DB_SSLMODE", "disable")

	if err := logger.Init(false); err != nil {
		return err
	}

	cfg := config.Load()
	return Connect(&cfg.Database)
}
