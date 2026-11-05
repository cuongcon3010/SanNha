package identity

import (
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"sannha/be/internal/platform/database"
	"sannha/be/internal/platform/httpx"
)

//go:embed oauth.sql
var oauthSchema string

type OAuthConfig struct {
	APIOrigin      string
	FrontendOrigin string
	GoogleID       string
	GoogleSecret   string
	GitHubID       string
	GitHubSecret   string
	Secure         bool
}
type oauthProvider struct {
	config     oauth2.Config
	profileURL string
	emailURL   string
}
type oauthProfile struct {
	subject string
	name    string
	email   string
}

func (c OAuthConfig) providers() (map[string]oauthProvider, error) {
	result := map[string]oauthProvider{}
	for _, p := range []struct {
		name, id, secret, auth, token, profile, emails string
		scopes                                         []string
	}{
		{"google", c.GoogleID, c.GoogleSecret, "https://accounts.google.com/o/oauth2/v2/auth", "https://oauth2.googleapis.com/token", "https://openidconnect.googleapis.com/v1/userinfo", "", []string{"openid", "email", "profile"}},
		{"github", c.GitHubID, c.GitHubSecret, "https://github.com/login/oauth/authorize", "https://github.com/login/oauth/access_token", "https://api.github.com/user", "https://api.github.com/user/emails", []string{"read:user", "user:email"}},
	} {
		if p.id == "" && p.secret == "" {
			continue
		}
		if p.id == "" || p.secret == "" {
			return nil, fmt.Errorf("OAuth %s requires both client ID and secret", p.name)
		}
		u, err := url.Parse(c.APIOrigin)
		if err != nil || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"))) {
			return nil, fmt.Errorf("OAUTH_API_ORIGIN must be an HTTPS origin (HTTP only on localhost)")
		}
		if u.Scheme == "https" && !c.Secure {
			return nil, fmt.Errorf("HTTPS OAuth requires COOKIE_SECURE=true")
		}
		result[p.name] = oauthProvider{config: oauth2.Config{ClientID: p.id, ClientSecret: p.secret, RedirectURL: c.APIOrigin + "/api/v1/auth/oauth/" + p.name + "/callback", Scopes: p.scopes, Endpoint: oauth2.Endpoint{AuthURL: p.auth, TokenURL: p.token}}, profileURL: p.profile, emailURL: p.emails}
	}
	return result, nil
}

func getOAuthJSON(ctx context.Context, client *http.Client, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "SanNha-OAuth")
	response, err := client.Do(req)
	if err != nil {
		return errors.New("OAuth provider unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New("OAuth provider rejected profile request")
	}
	if err = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(out); err != nil {
		return errors.New("invalid OAuth profile response")
	}
	return nil
}
func (p oauthProvider) profile(ctx context.Context, client *http.Client, provider string) (oauthProfile, error) {
	if provider == "google" {
		var v struct {
			Sub      string `json:"sub"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Verified bool   `json:"email_verified"`
		}
		if err := getOAuthJSON(ctx, client, p.profileURL, &v); err != nil {
			return oauthProfile{}, err
		}
		if v.Sub == "" || !v.Verified {
			return oauthProfile{}, errors.New("Google email is not verified")
		}
		return oauthProfile{v.Sub, v.Name, v.Email}, nil
	}
	var v struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}
	if err := getOAuthJSON(ctx, client, p.profileURL, &v); err != nil {
		return oauthProfile{}, err
	}
	if v.ID <= 0 {
		return oauthProfile{}, errors.New("invalid GitHub account")
	}
	var emails []struct {
		Email    string `json:"email"`
		Verified bool   `json:"verified"`
		Primary  bool   `json:"primary"`
	}
	if err := getOAuthJSON(ctx, client, p.emailURL, &emails); err != nil {
		return oauthProfile{}, err
	}
	for _, e := range emails {
		if e.Verified && e.Primary {
			if v.Name == "" {
				v.Name = v.Login
			}
			return oauthProfile{strconv.FormatInt(v.ID, 10), v.Name, e.Email}, nil
		}
	}
	return oauthProfile{}, errors.New("GitHub requires a verified primary email")
}

func (s *Service) oauthSession(ctx context.Context, provider string, profile oauthProfile) (string, error) {
	email, err := normalizeEmail(profile.email)
	if err != nil {
		return "", err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	// Serialize first login for the same provider identity across API instances.
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, provider+":"+profile.subject); err != nil {
		return "", err
	}
	var id string
	err = tx.QueryRowContext(ctx, `SELECT user_id FROM identity_oauth_accounts WHERE provider=$1 AND subject=$2`, provider, profile.subject).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		id = rand.Text()
		name := strings.TrimSpace(profile.name)
		if name == "" {
			name = "Người chơi"
		}
		name = string([]rune(name)[:min(len([]rune(name)), 80)])
		// A random, discarded password prevents password login until an explicit recovery/linking flow exists.
		hash, e := passwordHash(rand.Text() + rand.Text())
		if e != nil {
			return "", e
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO identity_users(id,name,email,password_hash,role,created_at) VALUES($1,$2,$3,$4,'user',$5)`, id, name, email, hash, time.Now().UTC().Format(time.RFC3339))
		if database.IsUniqueViolation(err, "identity_users_email_key") {
			return "", httpx.Conflict("Email đã có tài khoản; đăng nhập bằng phương thức ban đầu. Không tự liên kết OAuth.")
		}
		if err != nil {
			return "", err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO identity_oauth_accounts(provider,subject,user_id) VALUES($1,$2,$3)`, provider, profile.subject, id); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	var locked bool
	if err = tx.QueryRowContext(ctx, `SELECT locked FROM identity_users WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
		return "", err
	}
	if locked {
		return "", httpx.Forbidden()
	}
	token := rand.Text() + rand.Text()
	if _, err = tx.ExecContext(ctx, `DELETE FROM identity_sessions WHERE expires_at<=$1`, time.Now().Unix()); err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO identity_sessions(token_hash,user_id,expires_at) VALUES($1,$2,$3)`, tokenHash(token), id, time.Now().Add(sessionLifetime).Unix()); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) OAuthRoutes(mux *http.ServeMux, c OAuthConfig) error {
	providers, err := c.providers()
	if err != nil {
		return err
	}
	if err = database.Migrate(s.db, "identity_oauth_pg_v1", oauthSchema); err != nil {
		return err
	}
	mux.HandleFunc("GET /api/v1/auth/oauth/providers", func(w http.ResponseWriter, r *http.Request) {
		enabled := []string{}
		for _, name := range []string{"google", "github"} {
			if _, ok := providers[name]; ok {
				enabled = append(enabled, name)
			}
		}
		httpx.JSON(w, 200, enabled)
	})
	mux.HandleFunc("GET /api/v1/auth/oauth/{provider}/start", rateLimit(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		name := r.PathValue("provider")
		p, ok := providers[name]
		if !ok {
			return httpx.NotFound()
		}
		state, browser, verifier := rand.Text()+rand.Text(), rand.Text()+rand.Text(), oauth2.GenerateVerifier()
		if _, err := s.db.ExecContext(r.Context(), `DELETE FROM identity_oauth_flows WHERE expires_at<=$1`, time.Now().Unix()); err != nil {
			return err
		}
		if _, err := s.db.ExecContext(r.Context(), `INSERT INTO identity_oauth_flows(state_hash,browser_hash,provider,verifier,expires_at) VALUES($1,$2,$3,$4,$5)`, tokenHash(state), tokenHash(browser), name, verifier, time.Now().Add(10*time.Minute).Unix()); err != nil {
			return err
		}
		http.SetCookie(w, &http.Cookie{Name: "sannha_oauth_" + name, Value: browser, Path: "/api/v1/auth/oauth/" + name, HttpOnly: true, Secure: c.Secure, SameSite: http.SameSiteLaxMode, MaxAge: 600})
		http.Redirect(w, r, p.config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier)), http.StatusFound)
		return nil
	})))
	mux.HandleFunc("GET /api/v1/auth/oauth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		failure := func(code string) {
			http.Redirect(w, r, c.FrontendOrigin+"/dang-nhap?oauth_error="+code, http.StatusSeeOther)
		}
		name := r.PathValue("provider")
		p, ok := providers[name]
		if !ok {
			httpx.Fail(w, httpx.NotFound())
			return
		}
		cookie, err := r.Cookie("sannha_oauth_" + name)
		state := r.URL.Query().Get("state")
		if err != nil || len(state) < 32 || len(state) > 128 || len(cookie.Value) > 128 {
			failure("invalid_state")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		var verifier string
		err = s.db.QueryRowContext(ctx, `DELETE FROM identity_oauth_flows WHERE state_hash=$1 AND browser_hash=$2 AND provider=$3 AND expires_at>$4 RETURNING verifier`, tokenHash(state), tokenHash(cookie.Value), name, time.Now().Unix()).Scan(&verifier)
		if err != nil {
			failure("invalid_state")
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "sannha_oauth_" + name, Value: "", Path: "/api/v1/auth/oauth/" + name, HttpOnly: true, Secure: c.Secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
		code := r.URL.Query().Get("code")
		if r.URL.Query().Get("error") != "" || code == "" || len(code) > 4096 {
			failure("cancelled")
			return
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: 10 * time.Second})
		token, err := p.config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			failure("provider_failed")
			return
		}
		profile, err := p.profile(ctx, p.config.Client(ctx, token), name)
		if err != nil {
			failure("email_unverified")
			return
		}
		session, err := s.oauthSession(ctx, name, profile)
		if err != nil {
			var public *httpx.Error
			if errors.As(err, &public) && public.Status == 409 {
				failure("account_exists")
			} else {
				failure("login_failed")
			}
			return
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: session, Path: "/", HttpOnly: true, Secure: c.Secure, SameSite: http.SameSiteLaxMode, MaxAge: int(sessionLifetime.Seconds())})
		http.Redirect(w, r, c.FrontendOrigin+"/tai-khoan", http.StatusSeeOther)
	})
	return nil
}
