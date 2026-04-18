package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

type Config struct {
	Env              string
	Host             string
	Port             string
	SessionSecret    string
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string
	OIDCBaseURL      string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCCallbackURL  string
}

func Load() Config {
	return Config{
		Env:              mustGetEnv("APP_ENV"),
		Host:             mustGetEnv("APP_HOST"),
		Port:             mustGetEnv("APP_PORT"),
		SessionSecret:    mustGetEnv("SESSION_SECRET"),
		PostgresHost:     mustGetEnv("POSTGRES_HOST"),
		PostgresPort:     mustGetEnv("POSTGRES_PORT"),
		PostgresUser:     mustGetEnv("POSTGRES_USER"),
		PostgresPassword: mustGetEnv("POSTGRES_PASSWORD"),
		PostgresDB:       mustGetEnv("POSTGRES_DB"),
		PostgresSSLMode:  mustGetEnv("POSTGRES_SSLMODE"),
		OIDCBaseURL:      mustGetEnv("OIDC_BASE_URL"),
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

func mustGetEnv(key string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}

	fmt.Fprintf(os.Stderr, "missing required environment variable %s; please set %s\n", key, key)
	os.Exit(1)
	return ""
}
