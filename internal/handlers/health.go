package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jfkonecn/web-app-template/pkg/response"
)

func Health(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{
		"status": "ok",
	})
}
