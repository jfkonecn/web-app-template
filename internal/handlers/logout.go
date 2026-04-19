package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Handler for our logout.
func LogoutPage(auth authFlow) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		idTokenHint, _ := session.Get("id_token").(string)
		session.Clear()
		session.Options(sessions.Options{
			Path:   "/",
			MaxAge: -1,
		})
		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		if logoutURL, ok := auth.LogoutURL(idTokenHint); ok {
			ctx.Redirect(http.StatusTemporaryRedirect, logoutURL)
			return
		}

		ctx.Redirect(http.StatusTemporaryRedirect, "/")
	}
}
