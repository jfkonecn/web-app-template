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

func AdminExamplePage(c *gin.Context) {
	session := sessions.Default(c)
	profile, _ := session.Get("profile").(map[string]interface{})
	permissions := stringClaims(profile["permissions"])

	if !hasPermission(permissions, "read:admin") {
		c.HTML(http.StatusForbidden, "403.html", gin.H{
			"requiredPermission": "read:admin",
		})
		return
	}

	name, _ := profile["name"].(string)
	email, _ := profile["email"].(string)

	c.HTML(http.StatusOK, "admin-example.html", gin.H{
		"name":               name,
		"email":              email,
		"requiredPermission": "read:admin",
	})
}

func hasPermission(permissions []string, required string) bool {
	for _, permission := range permissions {
		if permission == required {
			return true
		}
	}

	return false
}
