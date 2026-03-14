package config

import (
	"fmt"
	"os"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "8080"
)

type Config struct {
	Env  string
	Host string
	Port string
}

func Load() Config {
	return Config{
		Env:  getEnv("APP_ENV", "development"),
		Host: getEnv("APP_HOST", defaultHost),
		Port: getEnv("APP_PORT", defaultPort),
	}
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
