package handlers

import (
	"context"
	"encoding/gob"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	cookiepkg "github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type stubAuthFlow struct {
	exchangeToken *oauth2.Token
	exchangeErr   error
	profile       map[string]interface{}
	verifyErr     error
	logoutURL     string
	logoutHint    string
}

func (s *stubAuthFlow) AuthCodeURL(state string, _ ...oauth2.AuthCodeOption) string {
	values := url.Values{}
	values.Set("state", state)
	return "https://issuer.example/authorize?" + values.Encode()
}

func (s *stubAuthFlow) Exchange(_ context.Context, _ string, _ ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	if s.exchangeErr != nil {
		return nil, s.exchangeErr
	}
	return s.exchangeToken, nil
}

func (s *stubAuthFlow) VerifyIDTokenClaims(_ context.Context, _ *oauth2.Token) (map[string]interface{}, error) {
	if s.verifyErr != nil {
		return nil, s.verifyErr
	}
	return s.profile, nil
}

func (s *stubAuthFlow) LogoutURL(idTokenHint string) (string, bool) {
	s.logoutHint = idTokenHint
	if s.logoutURL == "" {
		return "", false
	}
	return s.logoutURL, true
}

func TestLoginPageRedirectsAndStoresState(t *testing.T) {
	t.Parallel()

	router := newAuthTestRouter(&stubAuthFlow{})

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}

	location := rec.Header().Get("Location")
	if !strings.HasPrefix(location, "https://issuer.example/authorize?") {
		t.Fatalf("expected auth redirect, got %q", location)
	}

	redirectURL, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}

	state := redirectURL.Query().Get("state")
	if state == "" {
		t.Fatal("expected state query parameter")
	}

	assertSessionCookiePresent(t, rec.Result().Cookies())
}

func TestCallbackPageSuccessStoresProfileAndRedirects(t *testing.T) {
	t.Parallel()

	auth := &stubAuthFlow{
		exchangeToken: (&oauth2.Token{AccessToken: "access-token"}).WithExtra(map[string]interface{}{
			"id_token": "raw-id-token",
		}),
		profile: map[string]interface{}{
			"name":  "Example Admin",
			"email": "admin@example.com",
		},
	}
	router := newAuthTestRouter(auth)

	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, httptest.NewRequest(http.MethodGet, "/login", nil))

	state := mustQueryValue(t, loginRec.Header().Get("Location"), "state")
	sessionCookie := mustCookie(t, loginRec.Result().Cookies(), "auth-session")

	callbackReq := httptest.NewRequest(http.MethodGet, "/callback?state="+url.QueryEscape(state)+"&code=abc123", nil)
	callbackReq.AddCookie(sessionCookie)
	callbackRec := httptest.NewRecorder()

	router.ServeHTTP(callbackRec, callbackReq)

	if callbackRec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, callbackRec.Code)
	}
	if location := callbackRec.Header().Get("Location"); location != "/user" {
		t.Fatalf("expected redirect to /user, got %q", location)
	}

	updatedSession := mustCookie(t, callbackRec.Result().Cookies(), "auth-session")
	userReq := httptest.NewRequest(http.MethodGet, "/user", nil)
	userReq.AddCookie(updatedSession)
	userRec := httptest.NewRecorder()

	router.ServeHTTP(userRec, userReq)

	if userRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, userRec.Code)
	}
	body := userRec.Body.String()
	if !strings.Contains(body, "Example Admin") {
		t.Fatalf("expected rendered profile name, got body %q", body)
	}
	if !strings.Contains(body, "admin@example.com") {
		t.Fatalf("expected rendered profile email, got body %q", body)
	}
}

func TestCallbackPageStoresOnlyMinimalProfileInSession(t *testing.T) {
	t.Parallel()

	auth := &stubAuthFlow{
		exchangeToken: (&oauth2.Token{AccessToken: "access-token"}).WithExtra(map[string]interface{}{
			"id_token": "raw-id-token",
		}),
		profile: map[string]interface{}{
			"name":               "Example Admin",
			"email":              "admin@example.com",
			"permissions":        []interface{}{"read:admin"},
			"preferred_username": "admin-user",
			"realm_access": map[string]interface{}{
				"roles": []string{"admin"},
			},
		},
	}
	router := newAuthTestRouter(auth)

	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, httptest.NewRequest(http.MethodGet, "/login", nil))

	state := mustQueryValue(t, loginRec.Header().Get("Location"), "state")
	sessionCookie := mustCookie(t, loginRec.Result().Cookies(), "auth-session")
	callbackReq := httptest.NewRequest(http.MethodGet, "/callback?state="+url.QueryEscape(state)+"&code=abc123", nil)
	callbackReq.AddCookie(sessionCookie)
	callbackRec := httptest.NewRecorder()

	router.ServeHTTP(callbackRec, callbackReq)

	updatedSession := mustCookie(t, callbackRec.Result().Cookies(), "auth-session")
	sessionValues := decodeSessionValuesFromSessionCookie(t, updatedSession)
	profile, _ := sessionValues["profile"].(map[string]interface{})

	if len(profile) != 3 {
		t.Fatalf("expected only name and email in session profile, got %#v", profile)
	}
	if profile["name"] != "Example Admin" {
		t.Fatalf("expected name claim in session profile, got %#v", profile["name"])
	}
	if profile["email"] != "admin@example.com" {
		t.Fatalf("expected email claim in session profile, got %#v", profile["email"])
	}
	permissions, _ := profile["permissions"].([]string)
	if len(permissions) != 1 || permissions[0] != "read:admin" {
		t.Fatalf("expected permissions claim in session profile, got %#v", profile["permissions"])
	}
	if sessionValues["id_token"] != "raw-id-token" {
		t.Fatalf("expected id_token to be stored for logout, got %#v", sessionValues["id_token"])
	}
}

func TestUserPageRendersNameFromSessionProfile(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	gob.Register(map[string]interface{}{})

	router := gin.New()
	store := cookiepkg.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("auth-session", store))
	router.SetHTMLTemplate(template.Must(template.New("user.html").Parse(`{{ .name }}|{{ .email }}`)))
	router.GET("/user", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("profile", map[string]interface{}{
			"name":  "Example Admin",
			"email": "admin@example.com",
		})
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		UserPage(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if body := rec.Body.String(); body != "Example Admin|admin@example.com" {
		t.Fatalf("expected rendered profile details, got %q", body)
	}
}

func TestAdminExamplePageReturnsForbiddenWithoutPermission(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	gob.Register(map[string]interface{}{})

	router := gin.New()
	store := cookiepkg.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("auth-session", store))
	router.SetHTMLTemplate(template.Must(template.New("403.html").Parse(`403|{{ .requiredPermission }}`)))
	router.GET("/admin-example", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("profile", map[string]interface{}{
			"name":  "Plain User",
			"email": "user@example.com",
		})
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		AdminExamplePage(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin-example", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if body := rec.Body.String(); body != "403|read:admin" {
		t.Fatalf("expected 403 page body, got %q", body)
	}
}

func TestAdminExamplePageReturnsProtectedPageWithPermission(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	gob.Register(map[string]interface{}{})

	router := gin.New()
	store := cookiepkg.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("auth-session", store))
	router.SetHTMLTemplate(template.Must(template.New("").
		New("admin-example.html").Parse(`OK|{{ .name }}|{{ .requiredPermission }}`)))
	router.GET("/admin-example", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("profile", map[string]interface{}{
			"name":        "Admin User",
			"email":       "admin@example.com",
			"permissions": []string{"read:admin"},
		})
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		AdminExamplePage(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin-example", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if body := rec.Body.String(); body != "OK|Admin User|read:admin" {
		t.Fatalf("expected protected page body, got %q", body)
	}
}

func TestCallbackPageRejectsInvalidState(t *testing.T) {
	t.Parallel()

	router := newAuthTestRouter(&stubAuthFlow{})

	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, httptest.NewRequest(http.MethodGet, "/login", nil))

	sessionCookie := mustCookie(t, loginRec.Result().Cookies(), "auth-session")
	req := httptest.NewRequest(http.MethodGet, "/callback?state=wrong&code=abc123", nil)
	req.AddCookie(sessionCookie)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Invalid state parameter.") {
		t.Fatalf("expected invalid state error, got %q", body)
	}
}

func TestCallbackPageReturnsUnauthorizedWhenCodeExchangeFails(t *testing.T) {
	t.Parallel()

	router := newAuthTestRouter(&stubAuthFlow{
		exchangeErr: errors.New("exchange failed"),
	})

	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, httptest.NewRequest(http.MethodGet, "/login", nil))

	state := mustQueryValue(t, loginRec.Header().Get("Location"), "state")
	sessionCookie := mustCookie(t, loginRec.Result().Cookies(), "auth-session")
	req := httptest.NewRequest(http.MethodGet, "/callback?state="+url.QueryEscape(state)+"&code=bad-code", nil)
	req.AddCookie(sessionCookie)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Failed to convert an authorization code into a token.") {
		t.Fatalf("expected code exchange error, got %q", body)
	}
}

func TestCallbackPageReturnsInternalServerErrorWhenTokenCannotBeVerified(t *testing.T) {
	t.Parallel()

	router := newAuthTestRouter(&stubAuthFlow{
		exchangeToken: &oauth2.Token{AccessToken: "access-token"},
		verifyErr:     errors.New("verify failed"),
	})

	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, httptest.NewRequest(http.MethodGet, "/login", nil))

	state := mustQueryValue(t, loginRec.Header().Get("Location"), "state")
	sessionCookie := mustCookie(t, loginRec.Result().Cookies(), "auth-session")
	req := httptest.NewRequest(http.MethodGet, "/callback?state="+url.QueryEscape(state)+"&code=bad-code", nil)
	req.AddCookie(sessionCookie)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Failed to verify ID Token.") {
		t.Fatalf("expected verify error, got %q", body)
	}
}

func TestLogoutPageClearsSessionAndRedirectsHome(t *testing.T) {
	t.Parallel()

	router := newAuthTestRouter(&stubAuthFlow{})

	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, httptest.NewRequest(http.MethodGet, "/login", nil))

	sessionCookie := mustCookie(t, loginRec.Result().Cookies(), "auth-session")
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	req.AddCookie(sessionCookie)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}
	if location := rec.Header().Get("Location"); location != "/" {
		t.Fatalf("expected redirect to /, got %q", location)
	}

	clearedCookie := mustCookie(t, rec.Result().Cookies(), "auth-session")
	if clearedCookie.MaxAge >= 0 {
		t.Fatalf("expected cleared session cookie MaxAge < 0, got %d", clearedCookie.MaxAge)
	}
}

func TestLogoutPageRedirectsToProviderLogoutWhenSupported(t *testing.T) {
	t.Parallel()

	auth := &stubAuthFlow{
		exchangeToken: (&oauth2.Token{AccessToken: "access-token"}).WithExtra(map[string]interface{}{
			"id_token": "raw-id-token",
		}),
		profile: map[string]interface{}{
			"name":  "Example Admin",
			"email": "admin@example.com",
		},
		logoutURL: "https://issuer.example/logout?id_token_hint=raw-id-token",
	}
	router := newAuthTestRouter(auth)

	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, httptest.NewRequest(http.MethodGet, "/login", nil))

	state := mustQueryValue(t, loginRec.Header().Get("Location"), "state")
	sessionCookie := mustCookie(t, loginRec.Result().Cookies(), "auth-session")

	callbackReq := httptest.NewRequest(http.MethodGet, "/callback?state="+url.QueryEscape(state)+"&code=abc123", nil)
	callbackReq.AddCookie(sessionCookie)
	callbackRec := httptest.NewRecorder()
	router.ServeHTTP(callbackRec, callbackReq)

	updatedSession := mustCookie(t, callbackRec.Result().Cookies(), "auth-session")
	logoutReq := httptest.NewRequest(http.MethodGet, "/logout", nil)
	logoutReq.AddCookie(updatedSession)
	logoutRec := httptest.NewRecorder()

	router.ServeHTTP(logoutRec, logoutReq)

	if logoutRec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, logoutRec.Code)
	}
	if location := logoutRec.Header().Get("Location"); location != auth.logoutURL {
		t.Fatalf("expected redirect to provider logout URL, got %q", location)
	}
	if auth.logoutHint != "raw-id-token" {
		t.Fatalf("expected id token hint %q, got %q", "raw-id-token", auth.logoutHint)
	}

	clearedCookie := mustCookie(t, logoutRec.Result().Cookies(), "auth-session")
	if clearedCookie.MaxAge >= 0 {
		t.Fatalf("expected cleared session cookie MaxAge < 0, got %d", clearedCookie.MaxAge)
	}
}

func newAuthTestRouter(auth authFlow) *gin.Engine {
	gin.SetMode(gin.TestMode)

	gob.Register(map[string]interface{}{})

	router := gin.New()
	store := cookiepkg.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("auth-session", store))
	router.SetHTMLTemplate(template.Must(template.New("user.html").Parse(`{{ .name }}|{{ .email }}`)))

	router.GET("/login", LoginPage(auth))
	router.GET("/callback", CallbackPage(auth))
	router.GET("/logout", LogoutPage(auth))
	router.GET("/user", UserPage)
	router.GET("/admin-example", AdminExamplePage)

	return router
}

func mustQueryValue(t *testing.T, rawURL string, key string) string {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url %q: %v", rawURL, err)
	}
	value := parsed.Query().Get(key)
	if value == "" {
		t.Fatalf("missing query parameter %q in %q", key, rawURL)
	}
	return value
}

func mustCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("missing cookie %q", name)
	return nil
}

func assertSessionCookiePresent(t *testing.T, cookies []*http.Cookie) {
	t.Helper()

	cookie := mustCookie(t, cookies, "auth-session")
	if cookie.Value == "" {
		t.Fatal("expected auth-session cookie value")
	}
}

func decodeSessionValuesFromSessionCookie(t *testing.T, cookie *http.Cookie) map[interface{}]interface{} {
	t.Helper()

	store := cookiepkg.NewStore([]byte("test-secret"))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	session, err := store.Get(req, "auth-session")
	if err != nil {
		t.Fatalf("decode session cookie: %v", err)
	}

	return session.Values
}
