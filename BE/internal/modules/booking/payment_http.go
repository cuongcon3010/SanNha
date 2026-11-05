package booking

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"sannha/be/internal/platform/httpx"
)

func equalSecret(a, b string) bool {
	x, y := sha256.Sum256([]byte(a)), sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(x[:], y[:]) == 1
}
func (s *Service) PaymentRoutes(mux *http.ServeMux, auth, admin httpx.Guard) {
	mux.HandleFunc("GET /api/v1/payments/config", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, 200, map[string]any{"enabled": s.paymentConfig.Enabled, "environment": s.paymentConfig.Environment})
	})
	for _, method := range []string{"GET", "POST"} {
		mux.HandleFunc(method+" /api/v1/bookings/{id}/payment", auth(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
			ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
			defer cancel()
			if r.Method == "POST" {
				var in struct{}
				if err := httpx.Decode(w, r, &in); err != nil {
					return err
				}
			}
			p, err := s.GetPayment(ctx, httpx.Current(r), r.PathValue("id"), r.Method == "POST")
			if err != nil {
				return err
			}
			httpx.JSON(w, 200, p)
			return nil
		})))
	}
	mux.HandleFunc("GET /api/v1/admin/payments/review", admin(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		rows, err := s.db.QueryContext(r.Context(), `SELECT e.transaction_id,e.booking_id,e.result,CASE WHEN e.result='paid' AND b.status='cancelled' THEN 'refund_required' ELSE e.reason END,e.received_at,e.amount,e.trans_type,e.transaction_time,e.reference_number FROM booking_payment_events e JOIN booking_bookings b ON b.id=e.booking_id WHERE e.result='review' OR (e.result='paid' AND b.status='cancelled') ORDER BY e.received_at DESC LIMIT 200`)
		if err != nil {
			return err
		}
		defer rows.Close()
		type review struct {
			TransactionID   string `json:"transactionId"`
			BookingID       string `json:"bookingId"`
			Result          string `json:"result"`
			Reason          string `json:"reason"`
			ReceivedAt      int64  `json:"receivedAt"`
			Amount          int64  `json:"amount"`
			TransType       string `json:"transType"`
			TransactionTime int64  `json:"transactionTime"`
			Reference       string `json:"reference"`
		}
		out := []review{}
		for rows.Next() {
			var item review
			if err = rows.Scan(&item.TransactionID, &item.BookingID, &item.Result, &item.Reason, &item.ReceivedAt, &item.Amount, &item.TransType, &item.TransactionTime, &item.Reference); err != nil {
				return err
			}
			out = append(out, item)
		}
		if err = rows.Err(); err != nil {
			return err
		}
		httpx.JSON(w, 200, out)
		return nil
	})))
	mux.HandleFunc("POST /api/v1/payments/vietqr/api/token_generate", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		if !s.paymentConfig.Enabled {
			return httpx.NotFound()
		}
		user, password, ok := r.BasicAuth()
		if !ok || !equalSecret(user, s.paymentConfig.CallbackUsername) || !equalSecret(password, s.paymentConfig.CallbackPassword) {
			return httpx.Unauthorized()
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		token := rand.Text() + rand.Text()
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if _, err = tx.ExecContext(ctx, `DELETE FROM booking_callback_tokens WHERE expires_at<=$1`, time.Now().Unix()); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO booking_callback_tokens(token_hash,expires_at) VALUES($1,$2)`, hashValue(token), time.Now().Add(5*time.Minute).Unix()); err != nil {
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"access_token": token, "token_type": "Bearer", "expires_in": 300})
		return nil
	}))
	mux.HandleFunc("POST /api/v1/payments/vietqr/bank/api/transaction-sync", func(w http.ResponseWriter, r *http.Request) {
		fail := func(status int, code string) {
			httpx.JSON(w, status, map[string]any{"error": true, "errorReason": code, "toastMessage": "Transaction not accepted", "object": nil})
		}
		if !s.paymentConfig.Enabled {
			fail(404, "DISABLED")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || len(token) < 32 || len(token) > 128 {
			fail(401, "INVALID_TOKEN")
			return
		}
		var valid bool
		err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM booking_callback_tokens WHERE token_hash=$1 AND expires_at>$2)`, hashValue(token), time.Now().Unix()).Scan(&valid)
		if err != nil {
			fail(500, "RETRY_LATER")
			return
		}
		if !valid {
			fail(401, "INVALID_TOKEN")
			return
		}
		var event Transaction
		if err = httpx.Decode(w, r, &event); err != nil {
			fail(400, "INVALID_BODY")
			return
		}
		ref, err := s.syncPayment(ctx, event)
		if err != nil {
			var public *httpx.Error
			if errors.As(err, &public) {
				fail(public.Status, "TRANSACTION_MISMATCH")
			} else {
				fail(500, "RETRY_LATER")
			}
			return
		}
		httpx.JSON(w, 200, map[string]any{"error": false, "errorReason": nil, "toastMessage": "Transaction recorded", "object": map[string]string{"reftransactionid": ref}})
	})
}
