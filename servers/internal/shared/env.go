package shared

import (
	"strconv"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload" // Automatically load .env file
)

func EnvString(k, fallback string) string {
	if v, ok := syscall.Getenv(k); ok {
		return v
	}
	return fallback
}

func EnvInt(k string, fallback int) int {
	if v, ok := syscall.Getenv(k); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func EnvDuration(k string, fallback time.Duration) time.Duration {
	if v, ok := syscall.Getenv(k); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return time.Duration(i) * time.Second
		}
	}
	return fallback
}

func EnvFloat64(k string, fallback float64) float64 {
	if v, ok := syscall.Getenv(k); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
