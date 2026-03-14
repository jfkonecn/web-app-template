package server

import (
	"github.com/gin-gonic/gin"
	"github.com/jfkonecn/web-app-template/internal/handlers"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", handlers.Health)

	return r
}
