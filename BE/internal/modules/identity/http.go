package identity

import (
	"net"
	"net/http"
	"sync"
	"time"

	"sannha/be/internal/platform/httpx"
)

const cookieName = "sannha_session"

func (s *Service) Guard(roles ...string) httpx.Guard {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				return httpx.Unauthorized()
			}
			a, err := s.Authenticate(r.Context(), cookie.Value)
			if err != nil {
				return err
			}
			if len(roles) > 0 {
				allowed := false
				for _, role := range roles {
					if a.Role == role {
						allowed = true
					}
				}
				if !allowed {
					return httpx.Forbidden()
				}
			}
			next(w, httpx.WithActor(r, a))
			return nil
		})
	}
}
func (s *Service) Optional(next http.HandlerFunc) http.HandlerFunc {
	return httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		if cookie, err := r.Cookie(cookieName); err == nil {
			a, err := s.Authenticate(r.Context(), cookie.Value)
			if err == nil {
				r = httpx.WithActor(r, a)
			} else if e, ok := err.(*httpx.Error); !ok || e.Status != 401 {
				return err
			}
		}
		next(w, r)
		return nil
	})
}

type attempts struct {
	count   int
	expires time.Time
}

func rateLimit(next http.HandlerFunc) http.HandlerFunc {
	var mu sync.Mutex
	entries := map[string]attempts{}
	return func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		now := time.Now()
		mu.Lock()
		for ip, a := range entries {
			if now.After(a.expires) {
				delete(entries, ip)
			}
		}
		a := entries[host]
		if a.expires.IsZero() {
			a.expires = now.Add(15 * time.Minute)
		}
		if a.count >= 30 || len(entries) >= 2048 && a.count == 0 {
			mu.Unlock()
			w.Header().Set("Retry-After", "900")
			httpx.Fail(w, &httpx.Error{Status: 429, Message: "Quá nhiều yêu cầu. Thử lại sau 15 phút."})
			return
		}
		a.count++
		entries[host] = a
		mu.Unlock()
		next(w, r)
	}
}
func (s *Service) Routes(mux *http.ServeMux, secure bool) {
	mux.HandleFunc("POST /api/v1/auth/register", rateLimit(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in Registration
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		u, err := s.Register(r.Context(), in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, u)
		return nil
	})))
	mux.HandleFunc("POST /api/v1/auth/login", rateLimit(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		u, token, err := s.Login(r.Context(), in.Email, in.Password)
		if err != nil {
			return err
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: token, Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: int(sessionLifetime.Seconds())})
		httpx.JSON(w, 200, u)
		return nil
	})))
	mux.HandleFunc("POST /api/v1/auth/logout", s.Guard()(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in struct{}
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		cookie, _ := r.Cookie(cookieName)
		if err := s.Logout(r.Context(), cookie.Value); err != nil {
			return err
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
		w.WriteHeader(204)
		return nil
	})))
	mux.HandleFunc("GET /api/v1/me", s.Guard()(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		u, err := s.Get(r.Context(), httpx.Current(r).ID)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, u)
		return nil
	})))
	mux.HandleFunc("PATCH /api/v1/me", s.Guard()(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in Profile
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		u, err := s.UpdateProfile(r.Context(), httpx.Current(r).ID, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, u)
		return nil
	})))
	mux.HandleFunc("GET /api/v1/admin/users", s.Guard("admin")(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		users, err := s.Users(r.Context())
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, users)
		return nil
	})))
	mux.HandleFunc("PATCH /api/v1/admin/users/{id}/lock", s.Guard("admin")(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Locked *bool `json:"locked"`
		}
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		if in.Locked == nil {
			return httpx.Invalid("Thiếu locked.")
		}
		u, err := s.Lock(r.Context(), httpx.Current(r), r.PathValue("id"), *in.Locked)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, u)
		return nil
	})))
}
