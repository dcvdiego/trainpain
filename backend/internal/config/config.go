package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	HSP      HSPConfig
	TfL      TfLConfig
}

type ServerConfig struct {
	Port string
	Host string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type HSPConfig struct {
	Username string
	Password string
	BaseURL  string
}

type TfLConfig struct {
	AppID  string
	AppKey string
}

func Load() (*Config, error) {
	// Try to read config file first
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.env", "development")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "trainpain_user")
	viper.SetDefault("database.password", "trainpain_dev_password")
	viper.SetDefault("database.dbname", "trainpain")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("hsp.base_url", "https://hsp-prod.rockshore.net")

	// Read config file if exists (optional)
	_ = viper.ReadInConfig() // Ignore error if file not found

	// Override with environment variables
	// This reads from SERVER_PORT, DATABASE_HOST, etc.
	viper.SetEnvPrefix("") // No prefix
	viper.AutomaticEnv()

	// Build config manually to support env vars
	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", viper.GetString("server.port")),
			Host: getEnv("SERVER_HOST", viper.GetString("server.host")),
			Env:  getEnv("SERVER_ENV", viper.GetString("server.env")),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", viper.GetString("database.host")),
			Port:     getEnvInt("DATABASE_PORT", viper.GetInt("database.port")),
			User:     getEnv("DATABASE_USER", viper.GetString("database.user")),
			Password: getEnv("DATABASE_PASSWORD", viper.GetString("database.password")),
			DBName:   getEnv("DATABASE_DBNAME", viper.GetString("database.dbname")),
			SSLMode:  getEnv("DATABASE_SSLMODE", viper.GetString("database.sslmode")),
		},
		HSP: HSPConfig{
			Username: getEnv("HSP_USERNAME", viper.GetString("hsp.username")),
			Password: getEnv("HSP_PASSWORD", viper.GetString("hsp.password")),
			BaseURL:  getEnv("HSP_BASE_URL", viper.GetString("hsp.base_url")),
		},
		TfL: TfLConfig{
			AppID:  getEnv("TFL_APP_ID", viper.GetString("tfl.app_id")),
			AppKey: getEnv("TFL_APP_KEY", viper.GetString("tfl.app_key")),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}
