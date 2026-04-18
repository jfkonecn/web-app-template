package handlers

import (
	"net/http"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jfkonecn/web-app-template/internal/config"
)

// Handler for our logout.
func LogoutPage(config config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		session.Clear()
		session.Options(sessions.Options{
			Path:   "/",
			MaxAge: -1,
		})
		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		if config.OIDCLogoutURL == "" {
			ctx.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}

		logoutUrl, err := url.Parse(config.OIDCLogoutURL)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		scheme := "http"
		if ctx.Request.TLS != nil {
			scheme = "https"
		}

		returnTo, err := url.Parse(scheme + "://" + ctx.Request.Host)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		parameters := url.Values{}
		parameters.Add("returnTo", returnTo.String())
		parameters.Add("client_id", config.OIDCClientID)
		logoutUrl.RawQuery = parameters.Encode()

		ctx.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
	}
}
