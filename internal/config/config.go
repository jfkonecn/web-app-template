package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

const (
	defaultHost             = "0.0.0.0"
	defaultPort             = "8080"
	defaultPostgresHost     = "localhost"
	defaultPostgresPort     = "5432"
	defaultPostgresUser     = "postgres"
	defaultPostgresPassword = "postgres"
	defaultPostgresDB       = "web_app_template"
	defaultPostgresSSLMode  = "disable"
)

type Config struct {
	Env              string
	Host             string
	Port             string
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string
}

func Load() Config {
	return Config{
		Env:              getEnv("APP_ENV", "development"),
		Host:             getEnv("APP_HOST", defaultHost),
		Port:             getEnv("APP_PORT", defaultPort),
		PostgresHost:     getEnv("POSTGRES_HOST", defaultPostgresHost),
		PostgresPort:     getEnv("POSTGRES_PORT", defaultPostgresPort),
		PostgresUser:     getEnv("POSTGRES_USER", defaultPostgresUser),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", defaultPostgresPassword),
		PostgresDB:       getEnv("POSTGRES_DB", defaultPostgresDB),
		PostgresSSLMode:  getEnv("POSTGRES_SSLMODE", defaultPostgresSSLMode),
	}
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func (c Config) PostgresDSN() string {
	query := url.Values{}
	query.Set("sslmode", c.PostgresSSLMode)

	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.PostgresUser, c.PostgresPassword),
		Host:     net.JoinHostPort(c.PostgresHost, c.PostgresPort),
		Path:     c.PostgresDB,
		RawQuery: query.Encode(),
	}).String()
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
