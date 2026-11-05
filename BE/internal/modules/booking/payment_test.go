package booking

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"sannha/be/internal/modules/venue"
	"sannha/be/internal/platform/httpx"
	"sannha/be/internal/testutil"
)

func testPaymentConfig() PaymentConfig {
	return PaymentConfig{Enabled: true, Environment: "sandbox", Username: "merchant-test", Password: rand.Text(), CallbackUsername: "callback-test", CallbackPassword: rand.Text() + rand.Text(), BankCode: "MB", BankAccount: "1234567890", BankName: "TEST ACCOUNT"}
}
func fakeVietQR(t *testing.T, cfg PaymentConfig) (*vietQRClient, *atomic.Int32) {
	t.Helper()
	count := new(atomic.Int32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/token_generate":
			u, p, ok := r.BasicAuth()
			if !ok || u != cfg.Username || p != cfg.Password {
				t.Error("missing API basic auth")
				w.WriteHeader(401)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"access_token": "test-access-token", "token_type": "Bearer", "expires_in": 300})
		case "/qr/generate-customer":
			if r.Header.Get("Authorization") != "Bearer test-access-token" {
				t.Error("missing bearer auth")
				w.WriteHeader(401)
				return
			}
			var in map[string]any
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			if in["qrType"] != float64(0) || in["transType"] != "C" {
				t.Error("not a credit dynamic QR")
			}
			count.Add(1)
			json.NewEncoder(w).Encode(map[string]any{"bankCode": in["bankCode"], "bankAccount": in["bankAccount"], "amount": fmt.Sprint(in["amount"]), "orderId": in["orderId"], "content": in["content"], "qrCode": "00020101021238570010A00000072701270006970422TESTQR"})
		default:
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(server.Close)
	client := newVietQRClient(cfg)
	client.baseURL = server.URL
	client.http = server.Client()
	return client, count
}
func TestVietQRConfigAndTransactionValidation(t *testing.T) {
	c := testPaymentConfig()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := c
	bad.Environment = "dev"
	if bad.Validate() == nil {
		t.Fatal("invalid environment accepted")
	}
	bad = c
	bad.CallbackPassword = "short"
	if bad.Validate() == nil {
		t.Fatal("weak callback password accepted")
	}
	bad = c
	bad.BankName = "Tên có dấu"
	if bad.Validate() == nil {
		t.Fatal("accented bank name accepted")
	}
	if equalSecret("a", "b") || !equalSecret("same", "same") {
		t.Fatal("credential comparison broken")
	}
	if err := (Transaction{Amount: 1, TransType: "C", TransactionID: "tx", Reference: "ref", OrderID: "SN123", BankAccount: "1234", Content: "SN123", TransactionTime: time.Now().UnixMilli()}).validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Transaction{Amount: -1}).validate(); err == nil {
		t.Fatal("invalid transfer accepted")
	}
}
func TestVietQRHTTPClient(t *testing.T) {
	c := testPaymentConfig()
	client, calls := fakeVietQR(t, c)
	p := Payment{OrderID: "SN12345678901", Amount: 66000, BankCode: c.BankCode, BankAccount: c.BankAccount, BankName: c.BankName, Content: "SN12345678901"}
	for i := 0; i < 2; i++ {
		if qr, err := client.Generate(context.Background(), p); err != nil || qr == "" {
			t.Fatal("generate QR failed", err)
		}
	}
	if calls.Load() != 2 {
		t.Fatal("QR request not performed")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"bankCode": "MB", "bankAccount": "WRONG", "amount": "66000", "orderId": p.OrderID, "content": p.Content, "qrCode": strings.Repeat("0", 60)})
	}))
	defer server.Close()
	client.baseURL = server.URL
	if _, err := client.Generate(context.Background(), p); err == nil {
		t.Fatal("mismatching QR receiver accepted")
	}
}

func paymentFixture(t *testing.T) (*Service, httpx.Actor, Booking, *atomic.Int32) {
	t.Helper()
	db, _ := testutil.Postgres(t)
	catalog, err := venue.New(db)
	if err != nil {
		t.Fatal(err)
	}
	owner := httpx.Actor{ID: "owner-test", Role: "owner"}
	v, err := catalog.Create(context.Background(), owner, venue.CreateInput{Name: "Payment Test", Sport: "Bóng đá", City: "Hà Nội", Address: "Address for payment test", Courts: []venue.CourtInput{{Name: "A", DayPrice: 180000, EveningPrice: 220000}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = catalog.Review(context.Background(), httpx.Actor{ID: "admin", Role: "admin"}, v.ID, "approved", ""); err != nil {
		t.Fatal(err)
	}
	s, err := New(db, catalog)
	if err != nil {
		t.Fatal(err)
	}
	cfg := testPaymentConfig()
	if err = s.ConfigurePayments(cfg); err != nil {
		t.Fatal(err)
	}
	client, calls := fakeVietQR(t, cfg)
	s.vietqr = client
	actor := httpx.Actor{ID: "player-test", Role: "user"}
	date := time.Now().Add(48 * time.Hour).In(vietnam).Format("2006-01-02")
	b, err := s.Create(context.Background(), actor, CreateInput{CourtID: v.Courts[0].ID, Date: date, Hour: 19, Name: "Test Player", Phone: "0901234567"}, false)
	if err != nil {
		t.Fatal(err)
	}
	return s, actor, b, calls
}
func transfer(p Payment) Transaction {
	return Transaction{BankAccount: p.BankAccount, Amount: p.Amount, TransType: "C", Content: p.Content, TransactionID: rand.Text(), TransactionTime: time.Now().UnixMilli(), Reference: rand.Text(), OrderID: p.OrderID}
}
func TestPaymentIsAtomicIdempotentAndAuthorized(t *testing.T) {
	s, a, b, calls := paymentFixture(t)
	ctx := context.Background()
	if _, err := s.GetPayment(ctx, httpx.Actor{ID: "other", Role: "user"}, b.ID, true); err == nil {
		t.Fatal("other user generated QR")
	}
	p, err := s.GetPayment(ctx, a, b.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if p.Amount != 66000 || len(p.OrderID) > 13 || !strings.HasPrefix(p.QRImage, "data:image/png;base64,") {
		t.Fatal("invalid deposit QR")
	}
	p2, err := s.GetPayment(ctx, a, b.ID, true)
	if err != nil || p2.OrderID != p.OrderID || calls.Load() != 1 {
		t.Fatal("QR retries changed order or recalled provider", err)
	}
	if _, err = s.SetStatus(ctx, httpx.Actor{ID: "owner-test", Role: "owner"}, b.ID, "confirmed"); err == nil {
		t.Fatal("manual confirmation bypassed deposit")
	}
	e := transfer(p)
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := s.syncPayment(ctx, e); results <- err }()
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal("duplicate callback failed", err)
		}
	}
	got, err := s.getBooking(ctx, b.ID)
	if err != nil || !got.DepositPaid || got.Status != "confirmed" {
		t.Fatal("deposit and booking not updated together", err)
	}
	var count int
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM booking_payment_events`).Scan(&count); err != nil || count != 1 {
		t.Fatal("callback duplicated", err)
	}
	e.Amount++
	if _, err = s.syncPayment(ctx, e); err == nil {
		t.Fatal("same transaction with changed data accepted")
	}
	if _, err = s.SetStatus(ctx, a, b.ID, "cancelled"); err != nil {
		t.Fatal(err)
	}
	got, _ = s.getBooking(ctx, b.ID)
	if !got.DepositPaid {
		t.Fatal("cancellation erased received money")
	}
	if _, err = s.syncPayment(ctx, transfer(p)); err != nil {
		t.Fatal(err)
	}
	got, _ = s.getBooking(ctx, b.ID)
	if got.Status != "cancelled" {
		t.Fatal("callback revived cancelled booking")
	}
	paid, err := s.GetPayment(ctx, a, b.ID, false)
	if err != nil || paid.QRImage != "" || paid.Status != "paid" || paid.BookingStatus != "cancelled" {
		t.Fatal("cancelled paid booking exposed payable QR or lost payment", err)
	}
}
func TestPaymentMismatchesRequireReview(t *testing.T) {
	for _, kind := range []string{"wrong_amount", "wrong_receiver", "debit", "expired", "cancelled"} {
		t.Run(kind, func(t *testing.T) {
			s, a, b, _ := paymentFixture(t)
			ctx := context.Background()
			p, err := s.GetPayment(ctx, a, b.ID, true)
			if err != nil {
				t.Fatal(err)
			}
			e := transfer(p)
			switch kind {
			case "wrong_amount":
				e.Amount--
			case "wrong_receiver":
				e.BankAccount = "9876543210"
			case "debit":
				e.TransType = "D"
			case "expired":
				if _, err = s.db.Exec(`UPDATE booking_payments SET created_at=$1 WHERE booking_id=$2`, time.Now().Add(-20*time.Minute).Unix(), b.ID); err != nil {
					t.Fatal(err)
				}
			case "cancelled":
				if _, err = s.SetStatus(ctx, a, b.ID, "cancelled"); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = s.syncPayment(ctx, e); err != nil {
				t.Fatal(err)
			}
			got, err := s.getBooking(ctx, b.ID)
			if err != nil || got.DepositPaid || got.Status == "confirmed" {
				t.Fatal("invalid payment confirmed booking", err)
			}
			saved, err := s.GetPayment(ctx, a, b.ID, false)
			if err != nil || saved.Status != "review" || saved.QRImage != "" {
				t.Fatal("mismatch not flagged for review", err)
			}
		})
	}
}
func TestCallbackAuthentication(t *testing.T) {
	s, a, b, _ := paymentFixture(t)
	p, err := s.GetPayment(context.Background(), a, b.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	guard := func(h http.HandlerFunc) http.HandlerFunc { return h }
	s.PaymentRoutes(mux, guard, guard)
	r := httptest.NewRequest("POST", "/api/v1/payments/vietqr/api/token_generate", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal("unauthenticated token issued")
	}
	r = httptest.NewRequest("POST", "/api/v1/payments/vietqr/api/token_generate", nil)
	r.SetBasicAuth(s.paymentConfig.CallbackUsername, s.paymentConfig.CallbackPassword)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var token struct {
		Access string `json:"access_token"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &token); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(transfer(p))
	for _, authorization := range []string{"", "Bearer fake-token-that-is-not-issued-123456789", "Bearer " + token.Access} {
		r = httptest.NewRequest("POST", "/api/v1/payments/vietqr/bank/api/transaction-sync", strings.NewReader(string(data)))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", authorization)
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		want := 401
		if authorization == "Bearer "+token.Access {
			want = 200
		}
		if w.Code != want {
			t.Fatalf("expected %d got %d", want, w.Code)
		}
	}
	if _, err = s.db.Exec(`UPDATE booking_callback_tokens SET expires_at=0`); err != nil {
		t.Fatal(err)
	}
	r = httptest.NewRequest("POST", "/api/v1/payments/vietqr/bank/api/transaction-sync", strings.NewReader(string(data)))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+token.Access)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal("expired callback token accepted")
	}
}
