package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jfkonecn/web-app-template/internal/authenticator"
	"github.com/jfkonecn/web-app-template/internal/config"
	"github.com/jfkonecn/web-app-template/internal/logging"
	"github.com/jfkonecn/web-app-template/internal/server"
)

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.Env)

	if logging.IsProduction(cfg.Env) {
		gin.SetMode(gin.ReleaseMode)
	}

	auth, err := authenticator.New(cfg)

	if err != nil {
		logger.Error("auth config failed to load", slog.String("error", err.Error()))
		os.Exit(1)
	}

	r := server.NewRouter(logger, auth, cfg)

	logger.Info("starting server",
		slog.String("env", cfg.Env),
		slog.String("address", cfg.Address()),
	)

	if err := r.Run(cfg.Address()); err != nil {
		logger.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
