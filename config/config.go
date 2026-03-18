package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Env            string
	JWTSecret      string
	JWTExpireHours int
}

// Load reads configuration from environment variables.
func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "soccer"),
			Password: getEnv("DB_PASSWORD", "soccer"),
			DBName:   getEnv("DB_NAME", "soccer_manager"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		App: AppConfig{
			Env:            getEnv("APP_ENV", "development"),
			JWTSecret:      getEnv("JWT_SECRET", "ascjnajskcnlkascklalscnk"),
			JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
		},
	}
}

// DSN returns the PostgreSQL data source name.
func (c *DatabaseConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" sslmode=" + c.SSLMode
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
