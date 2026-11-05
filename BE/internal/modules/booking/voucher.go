package booking

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"sannha/be/internal/platform/database"
	"sannha/be/internal/platform/httpx"
)

//go:embed voucher.sql
var voucherSchema string
var voucherCodePattern = regexp.MustCompile(`^[A-Z0-9]{3,24}$`)

type VoucherInput struct {
	Code        string `json:"code"`
	Kind        string `json:"kind"`
	Value       int64  `json:"value"`
	MinTotal    int64  `json:"minTotal"`
	MaxDiscount int64  `json:"maxDiscount"`
	StartsAt    int64  `json:"startsAt"`
	EndsAt      int64  `json:"endsAt"`
	MaxUses     int    `json:"maxUses"`
}
type Voucher struct {
	VoucherInput
	UsedCount int   `json:"usedCount"`
	Active    bool  `json:"active"`
	CreatedAt int64 `json:"createdAt"`
}

const voucherColumns = "code,kind,value,min_total,max_discount,starts_at,ends_at,max_uses,used_count,active,created_at"

type QuoteInput struct {
	CourtID     string `json:"courtId"`
	Date        string `json:"date"`
	Hour        int    `json:"hour"`
	VoucherCode string `json:"voucherCode"`
}
type Quote struct {
	Subtotal    int64  `json:"subtotal"`
	Discount    int64  `json:"discount"`
	Total       int64  `json:"total"`
	Deposit     int64  `json:"deposit"`
	VoucherCode string `json:"voucherCode"`
}

func normalizeVoucher(code string) string { return strings.ToUpper(strings.TrimSpace(code)) }
func scanVoucher(row scanner) (Voucher, error) {
	var v Voucher
	err := row.Scan(&v.Code, &v.Kind, &v.Value, &v.MinTotal, &v.MaxDiscount, &v.StartsAt, &v.EndsAt, &v.MaxUses, &v.UsedCount, &v.Active, &v.CreatedAt)
	return v, err
}
func (v VoucherInput) validate() error {
	if !voucherCodePattern.MatchString(v.Code) || (v.Kind != "fixed" && v.Kind != "percent") || v.Value < 1 || v.Kind == "percent" && v.Value > 99 || v.Value > 10000000 || v.MinTotal < 0 || v.MinTotal > 10000000 || v.MaxDiscount < 0 || v.MaxDiscount > 10000000 || v.MaxUses < 1 || v.MaxUses > 1000000 || v.StartsAt < 0 || v.EndsAt <= v.StartsAt || v.EndsAt <= time.Now().Unix() {
		return httpx.Invalid("Voucher không hợp lệ: mã 3–24 chữ/số, giảm cố định hoặc 1–99%, thời gian và số lượt phải hợp lệ.")
	}
	return nil
}
func Discount(v VoucherInput, subtotal int64) int64 {
	amount := v.Value
	if v.Kind == "percent" {
		amount = subtotal * v.Value / 100
	}
	if v.MaxDiscount > 0 {
		amount = min(amount, v.MaxDiscount)
	}
	// Keep a payable minimum for VietQR; this flow does not support free bookings.
	return max(int64(0), min(amount, subtotal-1000))
}
func (s *Service) CreateVoucher(ctx context.Context, a httpx.Actor, in VoucherInput) (Voucher, error) {
	if a.Role != "admin" {
		return Voucher{}, httpx.Forbidden()
	}
	in.Code = normalizeVoucher(in.Code)
	if err := in.validate(); err != nil {
		return Voucher{}, err
	}
	v := Voucher{VoucherInput: in, Active: true, CreatedAt: time.Now().Unix()}
	_, err := s.db.ExecContext(ctx, `INSERT INTO booking_vouchers(code,kind,value,min_total,max_discount,starts_at,ends_at,max_uses,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, in.Code, in.Kind, in.Value, in.MinTotal, in.MaxDiscount, in.StartsAt, in.EndsAt, in.MaxUses, v.CreatedAt)
	if database.IsUniqueViolation(err, "booking_vouchers_pkey") {
		return Voucher{}, httpx.Conflict("Mã voucher đã tồn tại; không tái sử dụng mã cũ.")
	}
	return v, err
}
func (s *Service) Vouchers(ctx context.Context, a httpx.Actor) ([]Voucher, error) {
	if a.Role != "admin" {
		return nil, httpx.Forbidden()
	}
	// shortcut: Admin lists at most 200 vouchers; add pagination before a larger campaign catalog.
	rows, err := s.db.QueryContext(ctx, "SELECT "+voucherColumns+" FROM booking_vouchers ORDER BY created_at DESC,code LIMIT 200")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Voucher{}
	for rows.Next() {
		v, err := scanVoucher(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Service) SetVoucherActive(ctx context.Context, a httpx.Actor, code string, active bool) (Voucher, error) {
	if a.Role != "admin" {
		return Voucher{}, httpx.Forbidden()
	}
	v, err := scanVoucher(s.db.QueryRowContext(ctx, `UPDATE booking_vouchers SET active=$1 WHERE code=$2 RETURNING `+voucherColumns, active, normalizeVoucher(code)))
	if errors.Is(err, sql.ErrNoRows) {
		return Voucher{}, httpx.NotFound()
	}
	return v, err
}

// Voucher is locked until booking commit; capacity and one-use-per-user are reserved atomically.
func lockVoucher(ctx context.Context, tx *sql.Tx, code string) (Voucher, error) {
	return scanVoucher(tx.QueryRowContext(ctx, "SELECT "+voucherColumns+" FROM booking_vouchers WHERE code=$1 FOR UPDATE", code))
}
func voucherQuote(ctx context.Context, tx *sql.Tx, userID, code string, subtotal int64) (Quote, error) {
	q := Quote{Subtotal: subtotal, Total: subtotal, Deposit: Deposit(subtotal), VoucherCode: normalizeVoucher(code)}
	if q.VoucherCode == "" {
		return q, nil
	}
	if !voucherCodePattern.MatchString(q.VoucherCode) {
		return Quote{}, httpx.Invalid("Mã voucher không hợp lệ.")
	}
	v, err := lockVoucher(ctx, tx, q.VoucherCode)
	if errors.Is(err, sql.ErrNoRows) {
		return Quote{}, httpx.Invalid("Voucher không tồn tại.")
	}
	if err != nil {
		return Quote{}, err
	}
	now := time.Now().Unix()
	if !v.Active || now < v.StartsAt || now >= v.EndsAt {
		return Quote{}, httpx.Conflict("Voucher chưa bắt đầu, hết hạn hoặc đã tắt.")
	}
	if v.UsedCount >= v.MaxUses {
		return Quote{}, httpx.Conflict("Voucher đã hết lượt.")
	}
	if subtotal < v.MinTotal {
		return Quote{}, httpx.Invalid("Giá tiền sân chưa đạt mức tối thiểu của voucher.")
	}
	var used bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM booking_voucher_uses WHERE voucher_code=$1 AND user_id=$2 AND NOT released)`, v.Code, userID).Scan(&used); err != nil {
		return Quote{}, err
	}
	if used {
		return Quote{}, httpx.Conflict("Tài khoản đã dùng voucher này.")
	}
	q.Discount = Discount(v.VoucherInput, subtotal)
	if q.Discount <= 0 {
		return Quote{}, httpx.Invalid("Voucher không giảm được giá cho lịch này.")
	}
	q.Total = subtotal - q.Discount
	q.Deposit = Deposit(q.Total)
	return q, nil
}
func (s *Service) Quote(ctx context.Context, a httpx.Actor, in QuoteInput) (Quote, error) {
	if a.ID == "" {
		return Quote{}, httpx.Unauthorized()
	}
	if err := validateSlot(in.Date, in.Hour); err != nil {
		return Quote{}, err
	}
	c, err := s.catalog.Court(ctx, in.CourtID)
	if err != nil {
		return Quote{}, err
	}
	if c.VenueStatus != "approved" {
		return Quote{}, httpx.NotFound()
	}
	if c.Maintenance {
		return Quote{}, httpx.Conflict("Sân đang bảo trì.")
	}
	var taken bool
	if err = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM booking_bookings WHERE court_id=$1 AND date=$2 AND hour=$3 AND status<>'cancelled')`, in.CourtID, in.Date, in.Hour).Scan(&taken); err != nil {
		return Quote{}, err
	}
	if taken {
		return Quote{}, httpx.Conflict("Khung giờ đã có lịch.")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Quote{}, err
	}
	defer tx.Rollback()
	return voucherQuote(ctx, tx, a.ID, in.VoucherCode, Price(c.Court, in.Hour))
}
func releaseVoucher(ctx context.Context, tx *sql.Tx, b Booking) error {
	if b.VoucherCode == "" || b.DepositPaid {
		return nil
	}
	// Caller locks voucher before booking, matching creation and avoiding slot/cancellation deadlocks.
	result, err := tx.ExecContext(ctx, `UPDATE booking_voucher_uses SET released=TRUE WHERE booking_id=$1 AND NOT released`, b.ID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count > 0 {
		_, err = tx.ExecContext(ctx, `UPDATE booking_vouchers SET used_count=used_count-1 WHERE code=$1`, b.VoucherCode)
	}
	return err
}
func (s *Service) VoucherRoutes(mux *http.ServeMux, auth, admin httpx.Guard) {
	mux.HandleFunc("POST /api/v1/bookings/quote", auth(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in QuoteInput
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		q, err := s.Quote(r.Context(), httpx.Current(r), in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, q)
		return nil
	})))
	mux.HandleFunc("GET /api/v1/admin/vouchers", admin(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		list, err := s.Vouchers(r.Context(), httpx.Current(r))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, list)
		return nil
	})))
	mux.HandleFunc("POST /api/v1/admin/vouchers", admin(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in VoucherInput
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		v, err := s.CreateVoucher(r.Context(), httpx.Current(r), in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, v)
		return nil
	})))
	mux.HandleFunc("PATCH /api/v1/admin/vouchers/{code}", admin(httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Active *bool `json:"active"`
		}
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		if in.Active == nil {
			return httpx.Invalid("Thiếu active.")
		}
		v, err := s.SetVoucherActive(r.Context(), httpx.Current(r), r.PathValue("code"), *in.Active)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, v)
		return nil
	})))
}
