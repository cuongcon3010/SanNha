package booking

import (
	"net/http"
	"sannha/be/internal/platform/httpx"
)

func (s *Service) Routes(mux *http.ServeMux, auth, owner httpx.Guard) {
	for _, route := range []struct {
		path  string
		block bool
		guard httpx.Guard
	}{{"POST /api/v1/bookings", false, auth}, {"POST /api/v1/owner/blocks", true, owner}} {
		mux.HandleFunc(route.path, route.guard(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
			var in CreateInput
			if err := httpx.Decode(w, r, &in); err != nil {
				return err
			}
			b, err := s.Create(r.Context(), httpx.Current(r), in, route.block)
			if err != nil {
				return err
			}
			httpx.JSON(w, 201, b)
			return nil
		})))
	}
	mux.HandleFunc("GET /api/v1/me/bookings", auth(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		list, err := s.Mine(r.Context(), httpx.Current(r))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, list)
		return nil
	})))
	mux.HandleFunc("GET /api/v1/owner/bookings", owner(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		list, err := s.ForOwner(r.Context(), httpx.Current(r), r.URL.Query().Get("venueId"), r.URL.Query().Get("date"))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, list)
		return nil
	})))
	mux.HandleFunc("GET /api/v1/venues/{id}/availability", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		slots, err := s.Availability(r.Context(), r.PathValue("id"), r.URL.Query().Get("date"))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, slots)
		return nil
	}))
	mux.HandleFunc("PATCH /api/v1/bookings/{id}/status", auth(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Status string `json:"status"`
		}
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		b, err := s.SetStatus(r.Context(), httpx.Current(r), r.PathValue("id"), in.Status)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, b)
		return nil
	})))
	mux.HandleFunc("GET /api/v1/owner/revenue", owner(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		report, err := s.Revenue(r.Context(), httpx.Current(r), r.URL.Query().Get("venueId"), r.URL.Query().Get("month"))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, report)
		return nil
	})))
}
