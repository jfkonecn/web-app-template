package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func UserPage(c *gin.Context) {
	session := sessions.Default(c)
	profile := session.Get("profile")

	c.HTML(http.StatusOK, "user.html", profile)
}
