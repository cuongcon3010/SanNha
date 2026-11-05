package booking

import (
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"sannha/be/internal/modules/venue"
	"sannha/be/internal/platform/database"
	"sannha/be/internal/platform/httpx"
)

//go:embed schema.postgres.sql
var schema string

// Catalog is the only way this module reads venue data; it never queries venue_* tables.
type Catalog interface {
	Court(context.Context, string) (venue.CourtInfo, error)
	Get(context.Context, string) (venue.Venue, error)
}
type Service struct {
	db            *sql.DB
	catalog       Catalog
	paymentConfig PaymentConfig
	vietqr        *vietQRClient
}
type Booking struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	VenueID     string `json:"venueId"`
	CourtID     string `json:"courtId"`
	Date        string `json:"date"`
	Hour        int    `json:"hour"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Total       int64  `json:"total"`
	Deposit     int64  `json:"deposit"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	DepositPaid bool   `json:"depositPaid"`
	Subtotal    int64  `json:"subtotal"`
	Discount    int64  `json:"discount"`
	VoucherCode string `json:"voucherCode"`
}
type CreateInput struct {
	CourtID     string `json:"courtId"`
	Date        string `json:"date"`
	Hour        int    `json:"hour"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	VoucherCode string `json:"voucherCode"`
}
type Slot struct {
	CourtID   string `json:"courtId"`
	Date      string `json:"date"`
	Hour      int    `json:"hour"`
	Price     int64  `json:"price"`
	Available bool   `json:"available"`
}

var vietnam = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)
var phonePattern = regexp.MustCompile(`^(0[0-9]{9}|\+84[0-9]{9})$`)

const columns = "id,user_id,venue_id,court_id,date,hour,name,phone,total,deposit,status,created_at,subtotal,discount,voucher_code"

func New(db *sql.DB, catalog Catalog) (*Service, error) {
	if err := database.Migrate(db, "booking_pg_v1", schema); err != nil {
		return nil, err
	}
	if err := database.Migrate(db, "booking_payment_pg_v1", paymentSchema); err != nil {
		return nil, err
	}
	if err := database.Migrate(db, "booking_voucher_pg_v1", voucherSchema); err != nil {
		return nil, err
	}
	return &Service{db: db, catalog: catalog}, nil
}
func parseDate(value string) (time.Time, error) {
	d, err := time.ParseInLocation("2006-01-02", value, vietnam)
	if err != nil || d.Format("2006-01-02") != value {
		return time.Time{}, httpx.Invalid("Ngày phải hợp lệ, định dạng YYYY-MM-DD.")
	}
	return d, nil
}
func validateSlot(date string, hour int) error {
	d, err := parseDate(date)
	if err != nil {
		return err
	}
	if hour < 16 || hour > 21 {
		return httpx.Invalid("Khung giờ hỗ trợ: 16:00–22:00, mỗi lịch 1 giờ.")
	}
	if !d.Add(time.Duration(hour) * time.Hour).After(time.Now()) {
		return httpx.Invalid("Không thể tạo lịch ở thời điểm đã qua.")
	}
	return nil
}
func Price(c venue.Court, hour int) int64 {
	if hour < 17 {
		return c.DayPrice
	}
	return c.EveningPrice
}
func Deposit(total int64) int64 { return (total*30 + 50) / 100 }
func (s *Service) Create(ctx context.Context, a httpx.Actor, in CreateInput, block bool) (Booking, error) {
	if a.ID == "" {
		return Booking{}, httpx.Unauthorized()
	}
	if err := validateSlot(in.Date, in.Hour); err != nil {
		return Booking{}, err
	}
	c, err := s.catalog.Court(ctx, in.CourtID)
	if err != nil {
		return Booking{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Phone = strings.TrimSpace(in.Phone)
	if utf8.RuneCountInString(in.Name) < 2 || utf8.RuneCountInString(in.Name) > 80 {
		return Booking{}, httpx.Invalid("Tên người đặt hoặc lý do khóa cần 2–80 ký tự.")
	}
	status := "pending"
	total := Price(c.Court, in.Hour)
	if block {
		if normalizeVoucher(in.VoucherCode) != "" {
			return Booking{}, httpx.Invalid("Không áp dụng voucher cho khóa giờ.")
		}
		if a.Role != "owner" || a.ID != c.OwnerID {
			return Booking{}, httpx.Forbidden()
		}
		status = "blocked"
		total = 0
		in.Phone = ""
	} else {
		if c.VenueStatus != "approved" {
			return Booking{}, httpx.NotFound()
		}
		if c.Maintenance {
			return Booking{}, httpx.Conflict("Sân đang bảo trì.")
		}
		if !phonePattern.MatchString(in.Phone) {
			return Booking{}, httpx.Invalid("Số điện thoại không hợp lệ.")
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Booking{}, err
	}
	defer tx.Rollback()
	q, err := voucherQuote(ctx, tx, a.ID, in.VoucherCode, total)
	if err != nil {
		return Booking{}, err
	}
	b := Booking{ID: rand.Text(), UserID: a.ID, VenueID: c.VenueID, CourtID: c.ID, Date: in.Date, Hour: in.Hour, Name: in.Name, Phone: in.Phone, Total: q.Total, Deposit: q.Deposit, Status: status, CreatedAt: time.Now().UTC().Format(time.RFC3339), Subtotal: q.Subtotal, Discount: q.Discount, VoucherCode: q.VoucherCode}
	_, err = tx.ExecContext(ctx, `INSERT INTO booking_bookings(`+columns+`) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`, b.ID, b.UserID, b.VenueID, b.CourtID, b.Date, b.Hour, b.Name, b.Phone, b.Total, b.Deposit, b.Status, b.CreatedAt, b.Subtotal, b.Discount, b.VoucherCode)
	if database.IsUniqueViolation(err, "booking_active_slot") {
		return Booking{}, httpx.Conflict("Khung giờ đã được đặt hoặc bị khóa.")
	}
	if err != nil {
		return Booking{}, err
	}
	if b.VoucherCode != "" {
		if _, err = tx.ExecContext(ctx, `INSERT INTO booking_voucher_uses(booking_id,voucher_code,user_id) VALUES($1,$2,$3)`, b.ID, b.VoucherCode, a.ID); err != nil {
			return Booking{}, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE booking_vouchers SET used_count=used_count+1 WHERE code=$1`, b.VoucherCode); err != nil {
			return Booking{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Booking{}, err
	}
	return b, nil
}

type scanner interface{ Scan(...any) error }

func scan(row scanner) (Booking, error) {
	var b Booking
	err := row.Scan(&b.ID, &b.UserID, &b.VenueID, &b.CourtID, &b.Date, &b.Hour, &b.Name, &b.Phone, &b.Total, &b.Deposit, &b.Status, &b.CreatedAt, &b.Subtotal, &b.Discount, &b.VoucherCode, &b.DepositPaid)
	if errors.Is(err, sql.ErrNoRows) {
		return b, httpx.NotFound()
	}
	return b, err
}
func (s *Service) list(ctx context.Context, where string, args ...any) ([]Booking, error) {
	// shortcut: Lists return at most 500 records; add cursor pagination before production-sized histories.
	rows, err := s.db.QueryContext(ctx, "SELECT "+columns+",deposit_paid FROM booking_bookings WHERE "+where+" ORDER BY date,hour LIMIT 500", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Booking{}
	for rows.Next() {
		b, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}
func (s *Service) Mine(ctx context.Context, a httpx.Actor) ([]Booking, error) {
	if a.ID == "" {
		return nil, httpx.Unauthorized()
	}
	return s.list(ctx, "user_id=$1 AND status<>'blocked'", a.ID)
}
func (s *Service) ForOwner(ctx context.Context, a httpx.Actor, venueID, date string) ([]Booking, error) {
	v, err := s.catalog.Get(ctx, venueID)
	if err != nil {
		return nil, err
	}
	if a.Role != "admin" && (a.Role != "owner" || a.ID != v.OwnerID) {
		return nil, httpx.Forbidden()
	}
	if date != "" {
		if _, err = parseDate(date); err != nil {
			return nil, err
		}
		return s.list(ctx, "venue_id=$1 AND date=$2", venueID, date)
	}
	return s.list(ctx, "venue_id=$1", venueID)
}
func (s *Service) Availability(ctx context.Context, venueID, date string) ([]Slot, error) {
	d, err := parseDate(date)
	if err != nil {
		return nil, err
	}
	v, err := s.catalog.Get(ctx, venueID)
	if err != nil {
		return nil, err
	}
	if v.Status != "approved" {
		return nil, httpx.NotFound()
	}
	rows, err := s.list(ctx, "venue_id=$1 AND date=$2 AND status<>'cancelled'", venueID, date)
	if err != nil {
		return nil, err
	}
	taken := map[string]map[int]bool{}
	for _, b := range rows {
		if taken[b.CourtID] == nil {
			taken[b.CourtID] = map[int]bool{}
		}
		taken[b.CourtID][b.Hour] = true
	}
	result := []Slot{}
	now := time.Now()
	for _, c := range v.Courts {
		for h := 16; h <= 21; h++ {
			result = append(result, Slot{CourtID: c.ID, Date: date, Hour: h, Price: Price(c, h), Available: !c.Maintenance && !taken[c.ID][h] && d.Add(time.Duration(h)*time.Hour).After(now)})
		}
	}
	return result, nil
}
func (s *Service) SetStatus(ctx context.Context, a httpx.Actor, id, status string) (Booking, error) {
	b, err := s.getBooking(ctx, id)
	if err != nil {
		return Booking{}, err
	}
	c, err := s.catalog.Court(ctx, b.CourtID)
	if err != nil {
		return Booking{}, err
	}
	manager := a.Role == "admin" || a.Role == "owner" && a.ID == c.OwnerID
	if !manager && (a.ID != b.UserID || b.Status == "blocked") {
		return Booking{}, httpx.Forbidden()
	}
	if !manager && status != "cancelled" {
		return Booking{}, httpx.Forbidden()
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Booking{}, err
	}
	defer tx.Rollback()
	if status == "cancelled" && b.VoucherCode != "" {
		if _, err = lockVoucher(ctx, tx, b.VoucherCode); err != nil {
			return Booking{}, err
		}
	}
	b, err = scan(tx.QueryRowContext(ctx, "SELECT "+columns+",deposit_paid FROM booking_bookings WHERE id=$1 FOR UPDATE", id))
	if err != nil {
		return Booking{}, err
	}
	allowed := status == "cancelled" && (b.Status == "pending" || b.Status == "confirmed" || b.Status == "blocked") || manager && status == "confirmed" && b.Status == "pending" || manager && status == "completed" && b.Status == "confirmed"
	if !allowed {
		return Booking{}, httpx.Conflict("Không thể chuyển trạng thái lịch theo yêu cầu.")
	}
	if s.paymentConfig.Enabled && status == "confirmed" && !b.DepositPaid {
		return Booking{}, httpx.Conflict("Lịch dùng VietQR chỉ được tự xác nhận sau callback thanh toán hợp lệ.")
	}
	res, err := tx.ExecContext(ctx, `UPDATE booking_bookings SET status=$1 WHERE id=$2 AND status=$3`, status, id, b.Status)
	if err != nil {
		return Booking{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Booking{}, err
	}
	if n == 0 {
		return Booking{}, httpx.Conflict("Lịch vừa thay đổi, vui lòng tải lại.")
	}
	if status == "cancelled" {
		if err = releaseVoucher(ctx, tx, b); err != nil {
			return Booking{}, err
		}
	}
	b.Status = status
	if err = tx.Commit(); err != nil {
		return Booking{}, err
	}
	return b, nil
}

type Report struct {
	Month          string `json:"month"`
	Count          int    `json:"count"`
	Total          int64  `json:"total"`
	Completed      int64  `json:"completed"`
	PendingDeposit int64  `json:"pendingDeposit"`
}

func (s *Service) Revenue(ctx context.Context, a httpx.Actor, venueID, month string) (Report, error) {
	if _, err := time.Parse("2006-01", month); err != nil {
		return Report{}, httpx.Invalid("Tháng phải có định dạng YYYY-MM.")
	}
	v, err := s.catalog.Get(ctx, venueID)
	if err != nil {
		return Report{}, err
	}
	if a.Role != "admin" && (a.Role != "owner" || a.ID != v.OwnerID) {
		return Report{}, httpx.Forbidden()
	}
	var report Report
	report.Month = month
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(total),0),COALESCE(SUM(CASE WHEN status='completed' THEN total ELSE 0 END),0),COALESCE(SUM(CASE WHEN status='pending' THEN deposit ELSE 0 END),0) FROM booking_bookings WHERE venue_id=$1 AND substr(date,1,7)=$2 AND status NOT IN ('cancelled','blocked')`, venueID, month).Scan(&report.Count, &report.Total, &report.Completed, &report.PendingDeposit)
	return report, err
}
