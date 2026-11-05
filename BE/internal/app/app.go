package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"sannha/be/internal/modules/booking"
	"sannha/be/internal/modules/identity"
	"sannha/be/internal/modules/venue"
	"sannha/be/internal/platform/httpx"
)

type Config struct {
	Address        string
	DatabaseURL    string
	FrontendOrigin string
	CookieSecure   bool
	OAuth          identity.OAuthConfig
	Payments       booking.PaymentConfig
}

func ConfigFromEnv() (Config, error) {
	get := func(key, fallback string) string {
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
		return fallback
	}
	c := Config{Address: get("HTTP_ADDR", "127.0.0.1:8080"), DatabaseURL: get("DATABASE_URL", ""), FrontendOrigin: get("FRONTEND_ORIGIN", "http://localhost:3000")}
	secure, err := strconv.ParseBool(get("COOKIE_SECURE", "false"))
	if err != nil {
		return c, fmt.Errorf("COOKIE_SECURE: %w", err)
	}
	c.CookieSecure = secure
	enabled, err := strconv.ParseBool(get("VIETQR_ENABLED", "false"))
	if err != nil {
		return c, fmt.Errorf("VIETQR_ENABLED must be true or false")
	}
	c.Payments = booking.PaymentConfig{Enabled: enabled, Environment: get("VIETQR_ENV", "sandbox"), Username: get("VIETQR_USERNAME", ""), Password: get("VIETQR_PASSWORD", ""), CallbackUsername: get("VIETQR_CALLBACK_USERNAME", ""), CallbackPassword: get("VIETQR_CALLBACK_PASSWORD", ""), BankCode: get("VIETQR_BANK_CODE", ""), BankAccount: get("VIETQR_BANK_ACCOUNT", ""), BankName: get("VIETQR_BANK_NAME", "")}
	if err = c.Payments.Validate(); err != nil {
		return c, err
	}
	c.OAuth = identity.OAuthConfig{APIOrigin: get("OAUTH_API_ORIGIN", "http://localhost:8080"), FrontendOrigin: c.FrontendOrigin, Secure: secure, GoogleID: get("GOOGLE_CLIENT_ID", ""), GoogleSecret: get("GOOGLE_CLIENT_SECRET", ""), GitHubID: get("GITHUB_CLIENT_ID", ""), GitHubSecret: get("GITHUB_CLIENT_SECRET", "")}
	u, err := url.Parse(c.FrontendOrigin)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.Path != "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return c, fmt.Errorf("FRONTEND_ORIGIN must be an exact http(s) origin without path")
	}
	if c.Payments.Enabled && c.Payments.Environment == "production" && (!secure || u.Scheme != "https") {
		return c, fmt.Errorf("VietQR production requires HTTPS FRONTEND_ORIGIN and COOKIE_SECURE=true")
	}
	if u.Scheme == "https" && !c.CookieSecure {
		return c, fmt.Errorf("HTTPS frontend requires COOKIE_SECURE=true")
	}
	if c.Address == "" || strings.TrimSpace(c.DatabaseURL) == "" {
		return c, fmt.Errorf("HTTP_ADDR and DATABASE_URL must not be empty")
	}
	return c, nil
}

// New is the composition root: module wiring lives here, not in domain modules.
func New(db *sql.DB, c Config) (http.Handler, error) {
	accounts, err := identity.New(db)
	if err != nil {
		return nil, err
	}
	catalog, err := venue.New(db)
	if err != nil {
		return nil, err
	}
	reservations, err := booking.New(db, catalog)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	if err = reservations.ConfigurePayments(c.Payments); err != nil {
		return nil, err
	}
	accounts.Routes(mux, c.CookieSecure)
	c.OAuth.FrontendOrigin = c.FrontendOrigin
	c.OAuth.Secure = c.CookieSecure
	if err := accounts.OAuthRoutes(mux, c.OAuth); err != nil {
		return nil, err
	}
	catalog.Routes(mux, accounts.Guard(), accounts.Guard("owner"), accounts.Guard("admin"), accounts.Optional)
	reservations.Routes(mux, accounts.Guard(), accounts.Guard("owner", "admin"))
	reservations.PaymentRoutes(mux, accounts.Guard(), accounts.Guard("admin"))
	reservations.VoucherRoutes(mux, accounts.Guard(), accounts.Guard("admin"))
	mux.HandleFunc("GET /api/v1/health", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		if err := db.PingContext(r.Context()); err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]string{"status": "ok"})
		return nil
	}))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { httpx.Fail(w, httpx.NotFound()) })
	return httpx.Middleware(c.FrontendOrigin, mux), nil
}
