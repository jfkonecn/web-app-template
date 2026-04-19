package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Handler for our callback.
func CallbackPage(auth authFlow) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		if ctx.Query("state") != session.Get("state") {
			ctx.String(http.StatusBadRequest, "Invalid state parameter.")
			return
		}

		// Exchange an authorization code for a token.
		token, err := auth.Exchange(ctx.Request.Context(), ctx.Query("code"))
		if err != nil {
			ctx.String(http.StatusUnauthorized, "Failed to convert an authorization code into a token.")
			return
		}

		profile, err := auth.VerifyIDTokenClaims(ctx.Request.Context(), token)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		if rawIDToken, ok := token.Extra("id_token").(string); ok && rawIDToken != "" {
			session.Set("id_token", rawIDToken)
		}
		session.Set("profile", sessionProfile(profile))
		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to logged in page.
		ctx.Redirect(http.StatusTemporaryRedirect, "/user")
	}
}

func sessionProfile(profile map[string]interface{}) map[string]interface{} {
	sessionProfile := map[string]interface{}{}

	if name, ok := profile["name"].(string); ok && name != "" {
		sessionProfile["name"] = name
	}
	if email, ok := profile["email"].(string); ok && email != "" {
		sessionProfile["email"] = email
	}

	return sessionProfile
}
