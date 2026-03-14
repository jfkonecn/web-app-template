package logging

import (
	"log/slog"
	"os"
	"strings"
)

const prodEnv = "production"

func New(env string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if strings.EqualFold(env, prodEnv) {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

func IsProduction(env string) bool {
	return strings.EqualFold(env, prodEnv) || strings.EqualFold(env, "prod")
}
