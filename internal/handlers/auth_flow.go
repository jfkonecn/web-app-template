package handlers

import (
	"context"

	"golang.org/x/oauth2"
)

type authFlow interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	VerifyIDTokenClaims(ctx context.Context, token *oauth2.Token) (map[string]interface{}, error)
}
