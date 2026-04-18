package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func UserPage(c *gin.Context) {
	session := sessions.Default(c)
	profile, _ := session.Get("profile").(map[string]interface{})

	name, _ := profile["name"].(string)
	email, _ := profile["email"].(string)

	c.HTML(http.StatusOK, "user.html", gin.H{
		"name":  name,
		"email": email,
	})
}
