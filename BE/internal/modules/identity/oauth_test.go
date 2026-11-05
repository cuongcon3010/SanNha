package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"sannha/be/internal/testutil"
)

func TestOAuthConfig(t *testing.T) {
	if p, err := (OAuthConfig{}).providers(); err != nil || len(p) != 0 {
		t.Fatal("disabled providers should be allowed")
	}
	for _, c := range []OAuthConfig{{GoogleID: "id"}, {GoogleID: "id", GoogleSecret: "test", APIOrigin: "http://evil.example"}, {GitHubID: "id", GitHubSecret: "test", APIOrigin: "https://api.example"}} {
		if _, err := c.providers(); err == nil {
			t.Fatal("invalid provider configuration accepted")
		}
	}
}
func TestOAuthVerifiedEmails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/google":
			json.NewEncoder(w).Encode(map[string]any{"sub": "google-user", "email": "a@example.com", "email_verified": false})
		case "/github":
			json.NewEncoder(w).Encode(map[string]any{"id": 42, "login": "player"})
		case "/emails":
			json.NewEncoder(w).Encode([]map[string]any{{"email": "player@example.com", "verified": true, "primary": true}})
		}
	}))
	defer server.Close()
	p := oauthProvider{profileURL: server.URL + "/google"}
	if _, err := p.profile(context.Background(), server.Client(), "google"); err == nil {
		t.Fatal("unverified Google email accepted")
	}
	p.profileURL = server.URL + "/github"
	p.emailURL = server.URL + "/emails"
	got, err := p.profile(context.Background(), server.Client(), "github")
	if err != nil || got.subject != "42" || got.email != "player@example.com" {
		t.Fatal("GitHub identity not resolved", err)
	}
}
func TestOAuthStateReplayAndAccountIsolation(t *testing.T) {
	db, _ := testutil.Postgres(t)
	s, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	cfg := OAuthConfig{APIOrigin: "http://localhost:8080", FrontendOrigin: "http://localhost:3000", GoogleID: "test-id", GoogleSecret: "test-secret"}
	if err = s.OAuthRoutes(mux, cfg); err != nil {
		t.Fatal(err)
	}
	start := httptest.NewRecorder()
	mux.ServeHTTP(start, httptest.NewRequest("GET", "/api/v1/auth/oauth/google/start", nil))
	if start.Code != 302 {
		t.Fatal(start.Body.String())
	}
	u, _ := url.Parse(start.Header().Get("Location"))
	state := u.Query().Get("state")
	if state == "" || u.Query().Get("code_challenge_method") != "S256" || u.Query().Get("code_challenge") == "" {
		t.Fatal("missing state/PKCE")
	}
	cookie := start.Result().Cookies()[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatal("insecure flow cookie")
	}
	callback := func(withCookie bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "/api/v1/auth/oauth/google/callback?state="+state+"&error=access_denied", nil)
		if withCookie {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	if !strings.Contains(callback(false).Header().Get("Location"), "invalid_state") {
		t.Fatal("missing browser cookie accepted")
	}
	if !strings.Contains(callback(true).Header().Get("Location"), "cancelled") {
		t.Fatal("valid denied callback not consumed")
	}
	if !strings.Contains(callback(true).Header().Get("Location"), "invalid_state") {
		t.Fatal("callback replay accepted")
	}
	profile := oauthProfile{"subject-1", "OAuth Player", "player@example.com"}
	token, err := s.oauthSession(context.Background(), "google", profile)
	if err != nil {
		t.Fatal(err)
	}
	a, err := s.Authenticate(context.Background(), token)
	if err != nil || a.Role != "user" {
		t.Fatal("OAuth user not authenticated", err)
	}
	token2, err := s.oauthSession(context.Background(), "google", profile)
	if err != nil {
		t.Fatal(err)
	}
	a2, err := s.Authenticate(context.Background(), token2)
	if err != nil || a.ID != a2.ID {
		t.Fatal("returning account duplicated")
	}
	if _, err = s.oauthSession(context.Background(), "github", oauthProfile{"different-subject", "Player", profile.email}); err == nil {
		t.Fatal("accounts silently linked by email")
	}
	if _, err = db.Exec(`UPDATE identity_users SET locked=TRUE WHERE id=$1`, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.oauthSession(context.Background(), "google", profile); err == nil {
		t.Fatal("locked OAuth user logged in")
	}
	expiredState := oauth2.GenerateVerifier()
	if _, err = db.Exec(`INSERT INTO identity_oauth_flows VALUES($1,$2,'google','verifier',$3)`, tokenHash(expiredState), tokenHash(cookie.Value), time.Now().Add(-time.Minute).Unix()); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "/api/v1/auth/oauth/google/callback?state="+expiredState, nil)
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if !strings.Contains(w.Header().Get("Location"), "invalid_state") {
		t.Fatal("expired flow accepted")
	}
}
