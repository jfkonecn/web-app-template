package server

import (
	"encoding/gob"
	"log/slog"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jfkonecn/web-app-template/internal/authenticator"
	"github.com/jfkonecn/web-app-template/internal/config"
	"github.com/jfkonecn/web-app-template/internal/handlers"
	"github.com/jfkonecn/web-app-template/internal/middleware"
)

func NewRouter(logger *slog.Logger, auth *authenticator.Authenticator, config config.Config) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestLogger(logger), gin.Recovery())

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte(config.SessionSecret))
	r.Use(sessions.Sessions("auth-session", store))

	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	r.GET("/", handlers.Index)
	r.GET("/login", handlers.LoginPage(auth))
	r.GET("/logout", handlers.LogoutPage())
	r.GET("/user", handlers.UserPage)
	r.GET("/callback", handlers.CallbackPage(auth))
	r.GET("/healthz", handlers.Health)

	return r
}
