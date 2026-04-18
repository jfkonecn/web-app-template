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
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jfkonecn/web-app-template/internal/config"
	"golang.org/x/oauth2"
)

type stubAuthFlow struct {
	exchangeToken *oauth2.Token
	exchangeErr   error
	profile       map[string]interface{}
	verifyErr     error
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
		exchangeToken: &oauth2.Token{AccessToken: "access-token"},
		profile: map[string]interface{}{
			"name":  "Dex Admin",
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
	if !strings.Contains(body, "Dex Admin") {
		t.Fatalf("expected rendered profile name, got body %q", body)
	}
	if !strings.Contains(body, "admin@example.com") {
		t.Fatalf("expected rendered profile email, got body %q", body)
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

func TestLogoutPageClearsSessionAndRedirectsLocallyWhenNoProviderLogoutURL(t *testing.T) {
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

func newAuthTestRouter(auth authFlow) *gin.Engine {
	gin.SetMode(gin.TestMode)

	gob.Register(map[string]interface{}{})

	router := gin.New()
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("auth-session", store))
	router.SetHTMLTemplate(template.Must(template.New("user.html").Parse(`{{ .name }}|{{ .email }}`)))

	router.GET("/login", LoginPage(auth))
	router.GET("/callback", CallbackPage(auth))
	router.GET("/logout", LogoutPage(config.Config{}))
	router.GET("/user", UserPage)

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
