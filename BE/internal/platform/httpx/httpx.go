package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
)

type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string     { return e.Message }
func Invalid(message string) error { return &Error{http.StatusBadRequest, message} }
func Unauthorized() error          { return &Error{http.StatusUnauthorized, "Vui lòng đăng nhập."} }
func Forbidden() error {
	return &Error{http.StatusForbidden, "Bạn không có quyền thực hiện thao tác này."}
}
func NotFound() error               { return &Error{http.StatusNotFound, "Không tìm thấy dữ liệu."} }
func Conflict(message string) error { return &Error{http.StatusConflict, message} }

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status != http.StatusNoContent {
		_ = json.NewEncoder(w).Encode(value)
	}
}
func Fail(w http.ResponseWriter, err error) {
	var public *Error
	if errors.As(err, &public) {
		JSON(w, public.Status, map[string]string{"error": public.Message})
		return
	}
	slog.Error("request failed", "error", err)
	JSON(w, http.StatusInternalServerError, map[string]string{"error": "Có lỗi hệ thống. Vui lòng thử lại."})
}
func Decode(w http.ResponseWriter, r *http.Request, value any) error {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return &Error{http.StatusUnsupportedMediaType, "Yêu cầu Content-Type: application/json."}
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err = d.Decode(value); err != nil {
		return Invalid("JSON không hợp lệ, thừa trường hoặc vượt quá 64 KB.")
	}
	if err = d.Decode(new(any)); err != io.EOF {
		return Invalid("Chỉ chấp nhận một JSON object.")
	}
	return nil
}

type Actor struct {
	ID   string
	Role string
}
type actorKey struct{}

func WithActor(r *http.Request, actor Actor) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), actorKey{}, actor))
}
func Current(r *http.Request) Actor { a, _ := r.Context().Value(actorKey{}).(Actor); return a }

type Guard func(http.HandlerFunc) http.HandlerFunc
type Endpoint func(http.ResponseWriter, *http.Request) error

func Handle(f Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			Fail(w, err)
		}
	}
}

func Middleware(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Add("Vary", "Origin")
		if o := r.Header.Get("Origin"); o != "" {
			if o != origin {
				Fail(w, Forbidden())
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			Fail(w, Forbidden())
			return
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("request panic", "panic", recovered)
				Fail(w, errors.New("internal panic"))
			}
		}()
		if strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		Fail(w, NotFound())
	})
}
