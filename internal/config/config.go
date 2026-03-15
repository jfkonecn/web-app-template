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
	OIDCBaseURL      string
	OIDCLogoutURL    string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCCallbackURL  string
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
		OIDCBaseURL:      mustGetEnv("OIDC_BASE_URL"),
		OIDCLogoutURL:    getEnv("OIDC_LOGOUT_URL", ""),
		OIDCClientID:     mustGetEnv("OIDC_CLIENT_ID"),
		OIDCClientSecret: mustGetEnv("OIDC_CLIENT_SECRET"),
		OIDCCallbackURL:  mustGetEnv("OIDC_CALLBACK_URL"),
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

func mustGetEnv(key string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}

	fmt.Fprintf(os.Stderr, "missing required environment variable %s; please set %s\n", key, key)
	os.Exit(1)
	return ""
}
