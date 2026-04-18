package authenticator

import (
	"context"
	"errors"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jfkonecn/web-app-template/internal/config"
	"golang.org/x/oauth2"
)

// Save this file in ./platform/authenticator/auth.go

// Authenticator is used to authenticate our users.
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

// New instantiates the *Authenticator.
func New(config config.Config) (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		config.OIDCBaseURL,
	)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     config.OIDCClientID,
		ClientSecret: config.OIDCClientSecret,
		RedirectURL:  config.OIDCCallbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
	}, nil
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func (a *Authenticator) VerifyIDTokenClaims(ctx context.Context, token *oauth2.Token) (map[string]interface{}, error) {
	idToken, err := a.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, err
	}

	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		return nil, err
	}

	return profile, nil
}
