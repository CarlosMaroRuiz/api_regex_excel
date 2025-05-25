// config/config.go
package config

import (
	"os"
)

type Config struct {
	Port      string
	ExcelFile string
	APIURL    string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		ExcelFile: getEnv("EXCEL_FILE", "contactos.xlsx"),
		APIURL:    getEnv("API_URL", "http://localhost:8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}