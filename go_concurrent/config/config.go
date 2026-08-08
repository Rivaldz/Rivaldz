package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Workers        int
	ChannelBuf     int
	CSVDir         string
	OutputDir      string
	ReportInterval time.Duration
	CSVFiles       int
	CSVRows        int
}

func Load() Config {
	return Config{
		Workers:        getEnvInt("WORKERS", 8),
		ChannelBuf:     getEnvInt("CHANNEL_BUF", 10000),
		CSVDir:         getEnvStr("CSV_DIR", "./data"),
		OutputDir:      getEnvStr("OUTPUT_DIR", "./output"),
		ReportInterval: getEnvDuration("REPORT_INTERVAL", 1*time.Second),
		CSVFiles:       getEnvInt("CSV_FILES", 3),
		CSVRows:        getEnvInt("CSV_ROWS", 100000),
	}
}

func getEnvStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
