package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"sannha/be/internal/modules/booking"
	"sannha/be/internal/modules/identity"
	"sannha/be/internal/modules/venue"
	"sannha/be/internal/platform/database"
	"sannha/be/internal/platform/httpx"
	"sannha/be/internal/testutil"
)

type client struct {
	handler http.Handler
	cookie  *http.Cookie
}

func (c client) request(t *testing.T, method, path string, body any, status int) *httptest.ResponseRecorder {
	t.Helper()
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", "http://localhost:3000")
	if c.cookie != nil {
		r.AddCookie(c.cookie)
	}
	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("%s %s: expected %d got %d: %s", method, path, status, w.Code, w.Body.String())
	}
	return w
}
func decode[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(w.Body.Bytes(), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

type fixture struct {
	db                                *sql.DB
	databaseURL                       string
	public, admin, owner, user, other client
	ownerUser, userUser               identity.User
	venue                             venue.Venue
}

func setup(t *testing.T) fixture {
	t.Helper()
	db, dsn := testutil.Postgres(t)
	h, err := New(db, Config{FrontendOrigin: "http://localhost:3000", CookieSecure: true})
	if err != nil {
		t.Fatal(err)
	}
	f := fixture{db: db, databaseURL: dsn, public: client{handler: h}}
	password := rand.Text() + "-test-only"
	accounts, err := identity.New(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = accounts.BootstrapAdmin(context.Background(), "admin@example.test", password); err != nil {
		t.Fatal(err)
	}
	login := func(email string) client {
		w := f.public.request(t, "POST", "/api/v1/auth/login", map[string]string{"email": email, "password": password}, 200)
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteLaxMode {
			t.Fatal("insecure session cookie")
		}
		if strings.Contains(w.Body.String(), password) || strings.Contains(w.Body.String(), cookies[0].Value) {
			t.Fatal("credentials leaked in response")
		}
		return client{handler: h, cookie: cookies[0]}
	}
	register := func(email, role string) (client, identity.User) {
		w := f.public.request(t, "POST", "/api/v1/auth/register", identity.Registration{Name: "Tài khoản kiểm thử", Email: email, Phone: "0901234567", Password: password, Role: role}, 201)
		return login(email), decode[identity.User](t, w)
	}
	f.admin = login("admin@example.test")
	f.owner, f.ownerUser = register("owner@example.test", "owner")
	f.user, f.userUser = register("player@example.test", "user")
	f.other, _ = register("other-owner@example.test", "owner")
	w := f.owner.request(t, "POST", "/api/v1/owner/venues", venue.CreateInput{Name: "Sân kiểm thử", Sport: "Bóng đá", City: "Hà Nội", Address: "Cầu Giấy, Hà Nội", Amenities: []string{"Đèn LED"}, Courts: []venue.CourtInput{{Name: "Sân A", DayPrice: 180000, EveningPrice: 220000}}}, 201)
	f.venue = decode[venue.Venue](t, w)
	return f
}
func (f fixture) approve(t *testing.T) {
	t.Helper()
	f.admin.request(t, "PATCH", "/api/v1/admin/venues/"+f.venue.ID+"/review", map[string]string{"status": "approved"}, 200)
}
func futureDate() string {
	return time.Now().Add(48 * time.Hour).In(time.FixedZone("VN", 7*3600)).Format("2006-01-02")
}
func (f fixture) input() booking.CreateInput {
	return booking.CreateInput{CourtID: f.venue.Courts[0].ID, Date: futureDate(), Hour: 19, Name: "Người chơi kiểm thử", Phone: "0901234567"}
}

func TestAuthOwnershipAndRevocation(t *testing.T) {
	f := setup(t)
	f.public.request(t, "GET", "/api/v1/me", nil, 401)
	f.public.request(t, "POST", "/api/v1/auth/register", identity.Registration{Name: "Bad Admin", Email: "bad@example.test", Password: rand.Text(), Role: "admin"}, 400)
	f.public.request(t, "POST", "/api/v1/auth/register", identity.Registration{Name: "Duplicate", Email: "OWNER@EXAMPLE.TEST", Password: rand.Text(), Role: "user"}, 409)
	f.public.request(t, "POST", "/api/v1/auth/login", map[string]string{"email": "player@example.test", "password": "incorrect-password"}, 401)
	f.user.request(t, "GET", "/api/v1/admin/users", nil, 403)
	f.public.request(t, "GET", "/api/v1/admin/vouchers", nil, 401)
	f.user.request(t, "POST", "/api/v1/admin/vouchers", map[string]any{}, 403)
	f.public.request(t, "POST", "/api/v1/bookings/quote", map[string]any{}, 401)
	f.public.request(t, "GET", "/api/v1/bookings/unknown/payment", nil, 401)
	f.user.request(t, "GET", "/api/v1/admin/payments/review", nil, 403)
	f.user.request(t, "GET", "/api/v1/bookings/unknown/payment", nil, 503)
	f.user.request(t, "POST", "/api/v1/owner/venues", venue.CreateInput{}, 403)
	f.other.request(t, "PATCH", "/api/v1/owner/venues/"+f.venue.ID+"/courts/"+f.venue.Courts[0].ID, map[string]any{"dayPrice": 100000, "eveningPrice": 200000, "maintenance": false}, 403)
	f.user.request(t, "PATCH", "/api/v1/me", map[string]any{"name": "Player", "phone": "0901234567", "area": "Hà Nội", "role": "admin"}, 400)
	w := f.user.request(t, "PATCH", "/api/v1/me", identity.Profile{Name: "Tên mới", Phone: "0901111111", Area: "Hà Nội"}, 200)
	if decode[identity.User](t, w).Role != "user" {
		t.Fatal("profile changed role")
	}
	f.admin.request(t, "PATCH", "/api/v1/admin/users/"+f.userUser.ID+"/lock", map[string]bool{"locked": true}, 200)
	f.user.request(t, "GET", "/api/v1/me", nil, 401)
	f.admin.request(t, "PATCH", "/api/v1/admin/users/"+f.userUser.ID+"/lock", map[string]bool{"locked": false}, 200)
	f.user.request(t, "GET", "/api/v1/me", nil, 401)
	f.owner.request(t, "POST", "/api/v1/auth/logout", struct{}{}, 204)
	f.owner.request(t, "GET", "/api/v1/me", nil, 401)
}

func TestVoucherAdminRoutesAndPriceQuote(t *testing.T) {
	f := setup(t)
	f.approve(t)
	in := booking.VoucherInput{Code: "api50", Kind: "fixed", Value: 50000, MinTotal: 100000, StartsAt: time.Now().Add(-time.Minute).Unix(), EndsAt: time.Now().Add(time.Hour).Unix(), MaxUses: 3}
	v := decode[booking.Voucher](t, f.admin.request(t, "POST", "/api/v1/admin/vouchers", in, 201))
	if v.Code != "API50" {
		t.Fatal("voucher code not normalized")
	}
	q := decode[booking.Quote](t, f.user.request(t, "POST", "/api/v1/bookings/quote", booking.QuoteInput{CourtID: f.input().CourtID, Date: futureDate(), Hour: 20, VoucherCode: "api50"}, 200))
	if q.Total != 170000 || q.Deposit != 51000 {
		t.Fatal("quote route returned wrong price")
	}
	f.user.request(t, "PATCH", "/api/v1/admin/vouchers/API50", map[string]bool{"active": false}, 403)
	f.admin.request(t, "PATCH", "/api/v1/admin/vouchers/API50", map[string]bool{"active": false}, 200)
	f.admin.request(t, "PATCH", "/api/v1/admin/vouchers/API50", map[string]any{}, 400)
	f.user.request(t, "POST", "/api/v1/bookings/quote", booking.QuoteInput{CourtID: f.input().CourtID, Date: futureDate(), Hour: 20, VoucherCode: "api50"}, 409)
	f.user.request(t, "POST", "/api/v1/bookings", map[string]any{"courtId": f.input().CourtID, "date": futureDate(), "hour": 20, "name": "Test Player", "phone": "0901234567", "voucherCode": "API50", "discount": 50000}, 400)
}

func TestVenueReviewBookingPricingAndPrivacy(t *testing.T) {
	f := setup(t)
	if len(decode[[]venue.Venue](t, f.public.request(t, "GET", "/api/v1/venues", nil, 200))) != 0 {
		t.Fatal("pending venue publicly visible")
	}
	f.public.request(t, "GET", "/api/v1/venues/"+f.venue.ID, nil, 404)
	f.user.request(t, "POST", "/api/v1/bookings", f.input(), 404)
	f.admin.request(t, "PATCH", "/api/v1/admin/venues/"+f.venue.ID+"/review", map[string]string{"status": "rejected", "comment": "short"}, 400)
	f.approve(t)
	f.user.request(t, "PUT", "/api/v1/me/favorites/"+f.venue.ID, nil, 204)
	if len(decode[[]venue.Venue](t, f.user.request(t, "GET", "/api/v1/me/favorites", nil, 200))) != 1 {
		t.Fatal("favorite not saved")
	}
	w := f.user.request(t, "POST", "/api/v1/bookings", f.input(), 201)
	b := decode[booking.Booking](t, w)
	if b.Total != 220000 || b.Deposit != 66000 || b.Status != "pending" || b.UserID != f.userUser.ID {
		t.Fatalf("incorrect booking: total=%d deposit=%d status=%s", b.Total, b.Deposit, b.Status)
	}
	f.user.request(t, "POST", "/api/v1/bookings", f.input(), 409)
	f.other.request(t, "PATCH", "/api/v1/bookings/"+b.ID+"/status", map[string]string{"status": "cancelled"}, 403)
	f.user.request(t, "PATCH", "/api/v1/bookings/"+b.ID+"/status", map[string]string{"status": "confirmed"}, 403)
	f.owner.request(t, "PATCH", "/api/v1/bookings/"+b.ID+"/status", map[string]string{"status": "confirmed"}, 200)
	f.owner.request(t, "PATCH", "/api/v1/bookings/"+b.ID+"/status", map[string]string{"status": "pending"}, 409)
	f.owner.request(t, "PATCH", "/api/v1/owner/venues/"+f.venue.ID+"/courts/"+b.CourtID, map[string]any{"dayPrice": 190000, "eveningPrice": 240000, "maintenance": false}, 200)
	mine := decode[[]booking.Booking](t, f.user.request(t, "GET", "/api/v1/me/bookings", nil, 200))
	if len(mine) != 1 || mine[0].Total != 220000 {
		t.Fatal("price change altered existing booking")
	}
	availability := f.public.request(t, "GET", "/api/v1/venues/"+f.venue.ID+"/availability?date="+b.Date, nil, 200)
	if strings.Contains(availability.Body.String(), b.Phone) || strings.Contains(availability.Body.String(), b.Name) || strings.Contains(availability.Body.String(), b.ID) {
		t.Fatal("public availability leaked private booking data")
	}
	for _, slot := range decode[[]booking.Slot](t, availability) {
		if slot.Hour == 19 && slot.Available {
			t.Fatal("booked slot advertised as free")
		}
	}
	f.user.request(t, "PATCH", "/api/v1/bookings/"+b.ID+"/status", map[string]string{"status": "cancelled"}, 200)
	b = decode[booking.Booking](t, f.user.request(t, "POST", "/api/v1/bookings", f.input(), 201))
	if b.Total != 240000 {
		t.Fatal("new booking does not use new price")
	}
	f.owner.request(t, "PATCH", "/api/v1/owner/venues/"+f.venue.ID+"/courts/"+b.CourtID, map[string]any{"dayPrice": 190000, "eveningPrice": 240000, "maintenance": true}, 200)
	in := f.input()
	in.Hour = 20
	f.user.request(t, "POST", "/api/v1/bookings", in, 409)
	report := decode[booking.Report](t, f.owner.request(t, "GET", "/api/v1/owner/revenue?venueId="+f.venue.ID+"&month="+b.Date[:7], nil, 200))
	if report.Total != 240000 || report.Count != 1 {
		t.Fatalf("incorrect report %+v", report)
	}
	f.other.request(t, "GET", "/api/v1/owner/bookings?venueId="+f.venue.ID, nil, 403)
	f.user.request(t, "DELETE", "/api/v1/me/favorites/"+f.venue.ID, nil, 204)
}

func TestConcurrentBookingAndBlockRelease(t *testing.T) {
	f := setup(t)
	f.approve(t)
	catalog, err := venue.New(f.db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := booking.New(f.db, catalog)
	if err != nil {
		t.Fatal(err)
	}
	// A separate connection pool represents another API instance using the same DB.
	otherDB, err := database.Open(f.databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer otherDB.Close()
	otherCatalog, err := venue.New(otherDB)
	if err != nil {
		t.Fatal(err)
	}
	otherService, err := booking.New(otherDB, otherCatalog)
	if err != nil {
		t.Fatal(err)
	}
	actor := httpx.Actor{ID: f.userUser.ID, Role: "user"}
	results := make(chan error, 12)
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		instance := service
		if i%2 == 1 {
			instance = otherService
		}
		go func() {
			defer wg.Done()
			_, err := instance.Create(context.Background(), actor, f.input(), false)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success, conflicts := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if e, ok := err.(*httpx.Error); ok && e.Status == 409 {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflicts != 11 {
		t.Fatalf("success=%d conflicts=%d", success, conflicts)
	}
	in := f.input()
	in.Hour = 20
	in.Name = "Bảo trì"
	block := decode[booking.Booking](t, f.owner.request(t, "POST", "/api/v1/owner/blocks", in, 201))
	if block.Total != 0 || block.Deposit != 0 {
		t.Fatal("block charged money")
	}
	in.Name = "Người chơi"
	f.user.request(t, "POST", "/api/v1/bookings", in, 409)
	f.owner.request(t, "PATCH", "/api/v1/bookings/"+block.ID+"/status", map[string]string{"status": "cancelled"}, 200)
	f.user.request(t, "POST", "/api/v1/bookings", in, 201)
	for _, bad := range []booking.CreateInput{{CourtID: in.CourtID, Date: "2026-02-30", Hour: 20, Name: in.Name, Phone: in.Phone}, {CourtID: in.CourtID, Date: "2000-01-01", Hour: 20, Name: in.Name, Phone: in.Phone}, {CourtID: in.CourtID, Date: in.Date, Hour: 22, Name: in.Name, Phone: in.Phone}} {
		f.user.request(t, "POST", "/api/v1/bookings", bad, 400)
	}
}

func TestHTTPBoundariesAndPersistence(t *testing.T) {
	f := setup(t)
	f.approve(t)
	r := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{}`))
	r.Header.Set("Origin", "https://evil.example")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.public.handler.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("unexpected origin accepted")
	}
	r = httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(`{}`))
	r.Header.Set("Content-Type", "text/plain")
	w = httptest.NewRecorder()
	f.public.handler.ServeHTTP(w, r)
	if w.Code != 415 {
		t.Fatal("non-JSON body accepted")
	}
	f.user.request(t, "POST", "/api/v1/bookings", map[string]any{"courtId": f.input().CourtID, "date": futureDate(), "hour": 20, "name": "Test player", "phone": "0901234567", "total": 1}, 400)
	r = httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"email":"a@example.test","password":"`+strings.Repeat("x", 70<<10)+`"}`))
	r.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	f.public.handler.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal("oversize body accepted")
	}
	w = f.public.request(t, "GET", "/api/v1/health", nil, 200)
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" || w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("missing credential CORS headers")
	}
	for i := 0; i < 31; i++ {
		r = httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{`))
		r.Header.Set("Content-Type", "application/json")
		r.RemoteAddr = "192.0.2.99:1234"
		w = httptest.NewRecorder()
		f.public.handler.ServeHTTP(w, r)
	}
	if w.Code != 429 {
		t.Fatal("auth rate limiter did not enforce limit")
	}
	created := decode[booking.Booking](t, f.user.request(t, "POST", "/api/v1/bookings", f.input(), 201))
	if err := f.db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := database.Open(f.databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	handler, err := New(db, Config{FrontendOrigin: "http://localhost:3000", CookieSecure: true})
	if err != nil {
		t.Fatal(err)
	}
	f.user.handler = handler
	got := decode[identity.User](t, f.user.request(t, "GET", "/api/v1/me", nil, 200))
	if got.ID != f.userUser.ID {
		t.Fatal("user/session not persisted after reopen")
	}
	bookings := decode[[]booking.Booking](t, f.user.request(t, "GET", "/api/v1/me/bookings", nil, 200))
	if len(bookings) != 1 || bookings[0].ID != created.ID {
		t.Fatal("booking not persisted after reopen")
	}
}

func TestConfigRejectsUnsafeOrigins(t *testing.T) {
	t.Setenv("HTTP_ADDR", "127.0.0.1:8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("COOKIE_SECURE", "false")
	t.Setenv("FRONTEND_ORIGIN", "http://localhost:3000")
	if _, err := ConfigFromEnv(); err != nil {
		t.Fatal(err)
	}
	for _, origin := range []string{"*", "https://example.com", "http://localhost:3000/path", "http://user:password@localhost:3000", "http://localhost:3000?query=1"} {
		t.Setenv("FRONTEND_ORIGIN", origin)
		if _, err := ConfigFromEnv(); err == nil {
			t.Fatalf("unsafe config accepted for %s", origin)
		}
	}
	t.Setenv("FRONTEND_ORIGIN", "https://example.com")
	t.Setenv("COOKIE_SECURE", "true")
	if _, err := ConfigFromEnv(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DATABASE_URL", "")
	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("missing DATABASE_URL accepted")
	}
}

func TestVietQRProductionRequiresTLS(t *testing.T) {
	for key, value := range map[string]string{"DATABASE_URL": "postgres://localhost/test", "HTTP_ADDR": "127.0.0.1:8080", "VIETQR_ENABLED": "true", "VIETQR_ENV": "production", "VIETQR_USERNAME": "test-api", "VIETQR_PASSWORD": rand.Text(), "VIETQR_CALLBACK_USERNAME": "test-callback", "VIETQR_CALLBACK_PASSWORD": rand.Text() + rand.Text(), "VIETQR_BANK_CODE": "MB", "VIETQR_BANK_ACCOUNT": "1234567890", "VIETQR_BANK_NAME": "TEST ACCOUNT", "FRONTEND_ORIGIN": "http://localhost:3000", "COOKIE_SECURE": "false"} {
		t.Setenv(key, value)
	}
	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("production allowed insecure session")
	}
	t.Setenv("COOKIE_SECURE", "true")
	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("production allowed HTTP frontend")
	}
	t.Setenv("FRONTEND_ORIGIN", "https://app.example.com")
	if _, err := ConfigFromEnv(); err != nil {
		t.Fatal(err)
	}
}
