package venue

import (
	"net/http"
	"strings"

	"sannha/be/internal/platform/httpx"
)

func (s *Service) Routes(mux *http.ServeMux, auth, owner, admin httpx.Guard, optional func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/v1/venues", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		list, err := s.List(r.Context(), "", "approved")
		if err != nil {
			return err
		}
		out := []Venue{}
		q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		for _, v := range list {
			if sport := r.URL.Query().Get("sport"); sport != "" && v.Sport != sport {
				continue
			}
			if city := r.URL.Query().Get("city"); city != "" && v.City != city {
				continue
			}
			if !strings.Contains(strings.ToLower(v.Name+" "+v.Address), q) {
				continue
			}
			out = append(out, v)
		}
		httpx.JSON(w, 200, out)
		return nil
	}))
	mux.HandleFunc("GET /api/v1/venues/{id}", optional(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		v, err := s.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			return err
		}
		a := httpx.Current(r)
		if v.Status != "approved" && a.ID != v.OwnerID && a.Role != "admin" {
			return httpx.NotFound()
		}
		httpx.JSON(w, 200, v)
		return nil
	})))
	mux.HandleFunc("GET /api/v1/owner/venues", owner(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		list, err := s.List(r.Context(), httpx.Current(r).ID, "")
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, list)
		return nil
	})))
	mux.HandleFunc("POST /api/v1/owner/venues", owner(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in CreateInput
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		v, err := s.Create(r.Context(), httpx.Current(r), in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, v)
		return nil
	})))
	mux.HandleFunc("PATCH /api/v1/owner/venues/{id}/courts/{courtID}", owner(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in RatesInput
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		c, err := s.UpdateRates(r.Context(), httpx.Current(r), r.PathValue("id"), r.PathValue("courtID"), in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, c)
		return nil
	})))
	mux.HandleFunc("GET /api/v1/admin/venues", admin(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		status := r.URL.Query().Get("status")
		if status != "" && status != "pending" && status != "approved" && status != "rejected" {
			return httpx.Invalid("Trạng thái không hợp lệ.")
		}
		list, err := s.List(r.Context(), "", status)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, list)
		return nil
	})))
	mux.HandleFunc("PATCH /api/v1/admin/venues/{id}/review", admin(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Status  string `json:"status"`
			Comment string `json:"comment"`
		}
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		v, err := s.Review(r.Context(), httpx.Current(r), r.PathValue("id"), in.Status, in.Comment)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, v)
		return nil
	})))
	mux.HandleFunc("GET /api/v1/me/favorites", auth(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		list, err := s.Favorites(r.Context(), httpx.Current(r).ID)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, list)
		return nil
	})))
	for _, method := range []string{"PUT", "DELETE"} {
		mux.HandleFunc(method+" /api/v1/me/favorites/{id}", auth(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
			if err := s.Favorite(r.Context(), httpx.Current(r).ID, r.PathValue("id"), r.Method == "PUT"); err != nil {
				return err
			}
			w.WriteHeader(204)
			return nil
		})))
	}
}
