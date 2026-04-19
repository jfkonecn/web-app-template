package authenticator

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jfkonecn/web-app-template/internal/config"
	"golang.org/x/oauth2"
)

// Save this file in ./platform/authenticator/auth.go

// Authenticator is used to authenticate our users.
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
	endSessionEndpoint   string
	postLogoutRedirectURL string
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

	var metadata struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := provider.Claims(&metadata); err != nil {
		return nil, fmt.Errorf("read provider discovery metadata: %w", err)
	}

	postLogoutRedirectURL, err := postLogoutRedirectURLFromCallback(config.OIDCCallbackURL)
	if err != nil {
		return nil, err
	}

	endSessionEndpoint := ""
	if metadata.EndSessionEndpoint != "" {
		endSessionEndpointURL, err := url.Parse(metadata.EndSessionEndpoint)
		if err != nil {
			return nil, fmt.Errorf("parse end_session_endpoint: %w", err)
		}
		if !endSessionEndpointURL.IsAbs() {
			return nil, fmt.Errorf("parse end_session_endpoint: endpoint must be absolute")
		}
		endSessionEndpoint = endSessionEndpointURL.String()
	}

	conf := oauth2.Config{
		ClientID:     config.OIDCClientID,
		ClientSecret: config.OIDCClientSecret,
		RedirectURL:  config.OIDCCallbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &Authenticator{
		Provider:              provider,
		Config:                conf,
		endSessionEndpoint:    endSessionEndpoint,
		postLogoutRedirectURL: postLogoutRedirectURL,
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

func (a *Authenticator) LogoutURL(idTokenHint string) (string, bool) {
	if a.endSessionEndpoint == "" {
		return "", false
	}

	logoutURL, err := url.Parse(a.endSessionEndpoint)
	if err != nil {
		return "", false
	}

	query := logoutURL.Query()
	query.Set("post_logout_redirect_uri", a.postLogoutRedirectURL)
	if idTokenHint != "" {
		query.Set("id_token_hint", idTokenHint)
	}
	logoutURL.RawQuery = query.Encode()

	return logoutURL.String(), true
}

func postLogoutRedirectURLFromCallback(callbackURL string) (string, error) {
	parsed, err := url.Parse(callbackURL)
	if err != nil {
		return "", fmt.Errorf("parse OIDC_CALLBACK_URL: %w", err)
	}
	if !parsed.IsAbs() {
		return "", fmt.Errorf("parse OIDC_CALLBACK_URL: callback URL must be absolute")
	}

	return (&url.URL{
		Scheme: parsed.Scheme,
		Host:   parsed.Host,
		Path:   "/",
	}).String(), nil
}
