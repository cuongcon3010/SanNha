package booking

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"sannha/be/internal/platform/database"
	"sannha/be/internal/platform/httpx"
)

//go:embed payment.sql
var paymentSchema string

type Payment struct {
	BookingID     string `json:"bookingId"`
	OrderID       string `json:"orderId"`
	Amount        int64  `json:"amount"`
	BankCode      string `json:"bankCode"`
	BankAccount   string `json:"bankAccount"`
	BankName      string `json:"bankName"`
	Content       string `json:"content"`
	QRCode        string `json:"-"`
	Status        string `json:"status"`
	CreatedAt     int64  `json:"createdAt"`
	QRImage       string `json:"qrImage,omitempty"`
	Environment   string `json:"environment"`
	BookingStatus string `json:"bookingStatus"`
	ExpiresAt     int64  `json:"expiresAt"`
	Expired       bool   `json:"expired"`
	Subtotal      int64  `json:"subtotal"`
	Discount      int64  `json:"discount"`
	Total         int64  `json:"total"`
	VoucherCode   string `json:"voucherCode"`
}

const paymentColumns = "booking_id,order_id,amount,bank_code,bank_account,bank_name,content,qr_code,status,created_at,environment"

func scanPayment(row scanner) (Payment, error) {
	var p Payment
	err := row.Scan(&p.BookingID, &p.OrderID, &p.Amount, &p.BankCode, &p.BankAccount, &p.BankName, &p.Content, &p.QRCode, &p.Status, &p.CreatedAt, &p.Environment)
	return p, err
}
func hashValue(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func (s *Service) ConfigurePayments(c PaymentConfig) error {
	if err := c.Validate(); err != nil {
		return err
	}
	s.paymentConfig = c
	s.vietqr = nil
	if c.Enabled {
		s.vietqr = newVietQRClient(c)
	}
	return nil
}
func (s *Service) getBooking(ctx context.Context, id string) (Booking, error) {
	return scan(s.db.QueryRowContext(ctx, "SELECT "+columns+",deposit_paid FROM booking_bookings WHERE id=$1", id))
}
func (s *Service) GetPayment(ctx context.Context, a httpx.Actor, id string, generate bool) (Payment, error) {
	if s.vietqr == nil {
		return Payment{}, &httpx.Error{Status: 503, Message: "VietQR chưa được cấu hình."}
	}
	b, err := s.getBooking(ctx, id)
	if err != nil {
		return Payment{}, err
	}
	if b.UserID != a.ID {
		return Payment{}, httpx.Forbidden()
	}
	if generate && b.Status != "pending" {
		return Payment{}, httpx.Conflict("Chỉ tạo QR cho lịch đang chờ cọc.")
	}
	if !generate {
		var p Payment
		var date string
		var hour int
		err := s.db.QueryRowContext(ctx, `SELECT p.booking_id,p.order_id,p.amount,p.bank_code,p.bank_account,p.bank_name,p.content,p.qr_code,p.status,p.created_at,p.environment,b.status,b.date,b.hour,b.subtotal,b.discount,b.total,b.voucher_code FROM booking_payments p JOIN booking_bookings b ON b.id=p.booking_id WHERE p.booking_id=$1`, id).Scan(&p.BookingID, &p.OrderID, &p.Amount, &p.BankCode, &p.BankAccount, &p.BankName, &p.Content, &p.QRCode, &p.Status, &p.CreatedAt, &p.Environment, &p.BookingStatus, &date, &hour, &p.Subtotal, &p.Discount, &p.Total, &p.VoucherCode)
		if errors.Is(err, sql.ErrNoRows) {
			return Payment{}, httpx.NotFound()
		}
		if err != nil {
			return Payment{}, err
		}
		if bookingStarted(Booking{Date: date, Hour: hour}) {
			p.QRCode = ""
		}
		return s.publicPayment(p)
	}
	// Store a stable order before calling the provider; retries must reuse this order.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Payment{}, err
	}
	defer tx.Rollback()
	b, err = scan(tx.QueryRowContext(ctx, "SELECT "+columns+",deposit_paid FROM booking_bookings WHERE id=$1 FOR UPDATE", id))
	if err != nil {
		return Payment{}, err
	}
	if b.Status != "pending" || b.DepositPaid {
		return Payment{}, httpx.Conflict("Lịch không còn chờ thanh toán.")
	}
	if err = validateSlot(b.Date, b.Hour); err != nil {
		return Payment{}, err
	}
	order := "SN" + strings.ToUpper(rand.Text()[:11])
	cfg := s.paymentConfig
	_, err = tx.ExecContext(ctx, `INSERT INTO booking_payments(booking_id,order_id,amount,bank_code,bank_account,bank_name,content,created_at,environment) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(booking_id) DO NOTHING`, id, order, b.Deposit, cfg.BankCode, cfg.BankAccount, cfg.BankName, order, time.Now().Unix(), cfg.Environment)
	if err != nil {
		return Payment{}, err
	}
	if err = tx.Commit(); err != nil {
		return Payment{}, err
	}
	// Lock in the same order as callback/cancellation: booking first, then payment.
	tx, err = s.db.BeginTx(ctx, nil)
	if err != nil {
		return Payment{}, err
	}
	defer tx.Rollback()
	b, err = scan(tx.QueryRowContext(ctx, "SELECT "+columns+",deposit_paid FROM booking_bookings WHERE id=$1 FOR UPDATE", id))
	if err != nil {
		return Payment{}, err
	}
	p, err := scanPayment(tx.QueryRowContext(ctx, "SELECT "+paymentColumns+" FROM booking_payments WHERE booking_id=$1 FOR UPDATE", id))
	if err != nil {
		return Payment{}, err
	}
	if b.Status != "pending" || p.Status != "pending" {
		return Payment{}, httpx.Conflict("Lịch không còn chờ tạo QR; tải lại trạng thái thanh toán.")
	}
	if err = validateSlot(b.Date, b.Hour); err != nil {
		return Payment{}, err
	}
	if p.Environment != s.paymentConfig.Environment || p.BankAccount != s.paymentConfig.BankAccount || p.BankCode != s.paymentConfig.BankCode {
		return Payment{}, httpx.Conflict("Cấu hình merchant đã đổi. Liên hệ Sân Nhà để đối soát, không tạo QR mới cho đơn cũ.")
	}
	if time.Now().Unix() > p.CreatedAt+900 {
		return Payment{}, httpx.Conflict("QR đã hết hạn 15 phút. Không chuyển tiền; liên hệ Sân Nhà để hủy/đối soát lịch.")
	}
	if p.QRCode == "" {
		qr, err := s.vietqr.Generate(ctx, p)
		if err != nil {
			return Payment{}, &httpx.Error{Status: 502, Message: "Chưa lấy được QR từ VietQR. Thử lại với cùng lịch, không chuyển tiền thủ công."}
		}
		p.QRCode = qr
		if _, err = tx.ExecContext(ctx, `UPDATE booking_payments SET qr_code=$1 WHERE booking_id=$2`, qr, id); err != nil {
			return Payment{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Payment{}, err
	}
	p.BookingStatus = b.Status
	p.Subtotal = b.Subtotal
	p.Discount = b.Discount
	p.Total = b.Total
	p.VoucherCode = b.VoucherCode
	return s.publicPayment(p)
}
func (s *Service) publicPayment(p Payment) (Payment, error) {
	// shortcut: QR expires locally after 15 minutes; pending slots require explicit cancellation until a hold-expiry policy is implemented.
	p.ExpiresAt = p.CreatedAt + 900
	p.Expired = time.Now().Unix() >= p.ExpiresAt
	if p.QRCode != "" && p.Status == "pending" && p.BookingStatus == "pending" && !p.Expired {
		png, err := qrcode.Encode(p.QRCode, qrcode.Medium, 320)
		if err != nil {
			return Payment{}, err
		}
		p.QRImage = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	}
	return p, nil
}

type Transaction struct {
	BankAccount     string `json:"bankaccount"`
	Amount          int64  `json:"amount"`
	TransType       string `json:"transType"`
	Content         string `json:"content"`
	TransactionID   string `json:"transactionid"`
	TransactionTime int64  `json:"transactiontime"`
	Reference       string `json:"referencenumber"`
	OrderID         string `json:"orderId"`
	TerminalCode    string `json:"terminalCode"`
	SubTerminalCode string `json:"subTerminalCode"`
	ServiceCode     string `json:"serviceCode"`
	URLLink         string `json:"urlLink"`
	Sign            string `json:"sign"`
}

func (t Transaction) validate() error {
	if len(t.OrderID) < 1 || len(t.OrderID) > 13 || len(t.TransactionID) < 1 || len(t.TransactionID) > 200 || len(t.Reference) < 1 || len(t.Reference) > 200 || len(t.BankAccount) < 1 || len(t.BankAccount) > 32 || t.Amount <= 0 || t.TransactionTime <= 0 || len(t.Content) < 1 || len(t.Content) > 1000 || (t.TransType != "C" && t.TransType != "D") {
		return httpx.Invalid("Dữ liệu giao dịch không hợp lệ.")
	}
	return nil
}
func (t Transaction) fingerprint() string {
	data, _ := json.Marshal(struct {
		Account   string
		Amount    int64
		Type      string
		Order     string
		ID        string
		Reference string
		Time      int64
		Content   string
	}{t.BankAccount, t.Amount, t.TransType, t.OrderID, t.TransactionID, t.Reference, t.TransactionTime, t.Content})
	return hashValue(string(data))
}
func (s *Service) syncPayment(ctx context.Context, event Transaction) (string, error) {
	if err := event.validate(); err != nil {
		return "", err
	}
	var bookingID string
	err := s.db.QueryRowContext(ctx, `SELECT booking_id FROM booking_payments WHERE order_id=$1`, event.OrderID).Scan(&bookingID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", httpx.NotFound()
	}
	if err != nil {
		return "", err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	b, err := scan(tx.QueryRowContext(ctx, "SELECT "+columns+",deposit_paid FROM booking_bookings WHERE id=$1 FOR UPDATE", bookingID))
	if err != nil {
		return "", err
	}
	p, err := scanPayment(tx.QueryRowContext(ctx, "SELECT "+paymentColumns+" FROM booking_payments WHERE booking_id=$1 FOR UPDATE", bookingID))
	if err != nil {
		return "", err
	}
	var existing string
	err = tx.QueryRowContext(ctx, `SELECT fingerprint FROM booking_payment_events WHERE transaction_id=$1`, event.TransactionID).Scan(&existing)
	if err == nil {
		if existing != event.fingerprint() {
			return "", httpx.Conflict("Giao dịch trùng có nội dung khác.")
		}
		return event.TransactionID, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	reason := ""
	switch {
	case event.BankAccount != p.BankAccount:
		reason = "receiver_mismatch"
	case event.TransType != "C":
		reason = "not_credit"
	case event.Amount != p.Amount:
		reason = "amount_mismatch"
	case p.Environment != s.paymentConfig.Environment:
		reason = "environment_mismatch"
	case event.TransactionTime > (p.CreatedAt+900)*1000:
		reason = "payment_expired"
	case event.TransactionTime < p.CreatedAt*1000-60000 || event.TransactionTime > time.Now().Add(5*time.Minute).UnixMilli():
		reason = "transaction_time_mismatch"
	case b.Status != "pending":
		reason = "booking_not_pending"
	case bookingStarted(b):
		reason = "booking_started"
	case b.DepositPaid || p.Status == "paid":
		reason = "additional_transfer"
	}
	result := "paid"
	if reason != "" {
		result = "review"
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO booking_payment_events(transaction_id,bank_account,reference_number,booking_id,fingerprint,result,reason,received_at,amount,trans_type,transaction_time) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, event.TransactionID, event.BankAccount, event.Reference, bookingID, event.fingerprint(), result, reason, time.Now().Unix(), event.Amount, event.TransType, event.TransactionTime)
	if database.IsUniqueViolation(err, "booking_payment_events_pkey") || database.IsUniqueViolation(err, "booking_payment_events_bank_account_reference_number_key") {
		return "", httpx.Conflict("Mã giao dịch đã được sử dụng.")
	}
	if err != nil {
		return "", err
	}
	if result == "paid" {
		if _, err = tx.ExecContext(ctx, `UPDATE booking_bookings SET deposit_paid=TRUE,status='confirmed' WHERE id=$1`, bookingID); err != nil {
			return "", err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE booking_payments SET status='paid' WHERE booking_id=$1`, bookingID); err != nil {
			return "", err
		}
	} else if p.Status != "paid" {
		if _, err = tx.ExecContext(ctx, `UPDATE booking_payments SET status='review' WHERE booking_id=$1`, bookingID); err != nil {
			return "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return event.TransactionID, nil
}

func bookingStarted(b Booking) bool {
	d, err := parseDate(b.Date)
	return err != nil || !d.Add(time.Duration(b.Hour)*time.Hour).After(time.Now())
}
