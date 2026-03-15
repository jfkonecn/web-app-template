package handlers

import "github.com/gin-gonic/gin"

func LoginPage(c *gin.Context) {
	c.HTML(200, "login.html", gin.H{})
}

func Login(c *gin.Context) {
	c.HTML(200, "login.html", gin.H{
		"Submitted": true,
	})
}
