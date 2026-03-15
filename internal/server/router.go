package server

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jfkonecn/web-app-template/internal/handlers"
	"github.com/jfkonecn/web-app-template/internal/middleware"
)

func NewRouter(logger *slog.Logger) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestLogger(logger), gin.Recovery())
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	r.GET("/", handlers.Index)
	r.GET("/healthz", handlers.Health)

	return r
}
