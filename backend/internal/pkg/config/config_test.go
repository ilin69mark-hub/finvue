package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "enable")
	os.Setenv("SERVER_PORT", "9090")

	cfg := Load()

	if cfg.Database.Host != "testhost" {
		t.Errorf("Expected DB_HOST testhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("Expected DB_PORT 5433, got %d", cfg.Database.Port)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("Expected SERVER_PORT 9090, got %s", cfg.Server.Port)
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSLMODE")

	cfg := Load()

	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected DB_HOST localhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Expected DB_PORT 5432, got %d", cfg.Database.Port)
	}
	if cfg.Database.User != "finvue" {
		t.Errorf("Expected DB_USER finvue, got %s", cfg.Database.User)
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "finvue",
		Password: "secret",
		DBName:   "finvue",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()

	expected := "host=localhost port=5432 user=finvue password=secret dbname=finvue sslmode=disable"
	if dsn != expected {
		t.Errorf("DSN mismatch: got %s, expected %s", dsn, expected)
	}
}

func TestRedisConfig_Addr(t *testing.T) {
	cfg := RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	addr := cfg.Addr()

	if addr != "localhost:6379" {
		t.Errorf("Expected addr localhost:6379, got %s", addr)
	}
}

func TestServerConfig_Defaults(t *testing.T) {
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_READ_TIMEOUT")
	os.Unsetenv("SERVER_WRITE_TIMEOUT")

	cfg := Load()

	if cfg.Server.Port != "8080" {
		t.Errorf("Expected SERVER_PORT 8080, got %s", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 30 {
		t.Errorf("Expected ReadTimeout 30, got %d", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 30 {
		t.Errorf("Expected WriteTimeout 30, got %d", cfg.Server.WriteTimeout)
	}
}

func TestRedisConfig_Defaults(t *testing.T) {
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")

	cfg := Load()

	if cfg.Redis.Host != "localhost" {
		t.Errorf("Expected REDIS_HOST localhost, got %s", cfg.Redis.Host)
	}
	if cfg.Redis.Port != 6379 {
		t.Errorf("Expected REDIS_PORT 6379, got %d", cfg.Redis.Port)
	}
}