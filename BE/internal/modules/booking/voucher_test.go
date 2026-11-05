package booking

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"sannha/be/internal/platform/database"
	"sannha/be/internal/platform/httpx"
	"sannha/be/internal/testutil"
)

func voucherInput(code string) VoucherInput {
	return VoucherInput{Code: code, Kind: "fixed", Value: 50000, MinTotal: 100000, MaxUses: 10, StartsAt: time.Now().Add(-time.Minute).Unix(), EndsAt: time.Now().Add(time.Hour).Unix()}
}

var voucherAdmin = httpx.Actor{ID: "admin-test", Role: "admin"}

func TestVoucherDiscountUsesIntegerMoney(t *testing.T) {
	for _, tc := range []struct {
		name           string
		in             VoucherInput
		subtotal, want int64
	}{
		{"fixed", VoucherInput{Kind: "fixed", Value: 50000}, 220000, 50000},
		{"percent round down", VoucherInput{Kind: "percent", Value: 15}, 220001, 33000},
		{"percent capped", VoucherInput{Kind: "percent", Value: 50, MaxDiscount: 40000}, 220000, 40000},
		{"minimum payable", VoucherInput{Kind: "fixed", Value: 500000}, 220000, 219000},
		{"minimum price", VoucherInput{Kind: "fixed", Value: 1}, 1000, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Discount(tc.in, tc.subtotal); got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
	if normalizeVoucher(" promo50 ") != "PROMO50" {
		t.Fatal("voucher normalization failed")
	}
	for _, change := range []func(*VoucherInput){func(v *VoucherInput) { v.Code = "bad-code" }, func(v *VoucherInput) { v.Kind = "other" }, func(v *VoucherInput) { v.Value = 0 }, func(v *VoucherInput) { v.Kind = "percent"; v.Value = 100 }, func(v *VoucherInput) { v.MaxUses = 0 }, func(v *VoucherInput) { v.EndsAt = v.StartsAt }, func(v *VoucherInput) { v.MinTotal = -1 }} {
		v := voucherInput("VALID")
		change(&v)
		if v.validate() == nil {
			t.Fatal("invalid voucher input accepted")
		}
	}
}

func TestVoucherQuoteReservationAndDiscountedQR(t *testing.T) {
	s, a, original, _ := paymentFixture(t)
	ctx := context.Background()
	if _, err := s.CreateVoucher(ctx, a, voucherInput("PROMO50")); err == nil {
		t.Fatal("user created voucher")
	}
	v, err := s.CreateVoucher(ctx, voucherAdmin, voucherInput(" promo50 "))
	if err != nil || v.Code != "PROMO50" {
		t.Fatal("cannot create voucher", err)
	}
	if _, err = s.CreateVoucher(ctx, voucherAdmin, voucherInput("promo50")); err == nil {
		t.Fatal("duplicate code accepted")
	}
	qi := QuoteInput{CourtID: original.CourtID, Date: original.Date, Hour: 20, VoucherCode: "promo50"}
	for i := 0; i < 2; i++ {
		q, err := s.Quote(ctx, a, qi)
		if err != nil || q.Subtotal != 220000 || q.Discount != 50000 || q.Total != 170000 || q.Deposit != 51000 {
			t.Fatalf("wrong quote %+v %v", q, err)
		}
	}
	list, err := s.Vouchers(ctx, voucherAdmin)
	if err != nil || list[0].UsedCount != 0 {
		t.Fatal("preview consumed voucher", err)
	}
	in := CreateInput{CourtID: qi.CourtID, Date: qi.Date, Hour: qi.Hour, Name: "Voucher Player", Phone: "0901234567", VoucherCode: " PROMO50 "}
	b, err := s.Create(ctx, a, in, false)
	if err != nil {
		t.Fatal(err)
	}
	if b.Total != 170000 || b.Deposit != 51000 || b.Discount != 50000 || b.Subtotal != 220000 || b.VoucherCode != "PROMO50" {
		t.Fatalf("wrong booking %+v", b)
	}
	if _, err = s.SetVoucherActive(ctx, voucherAdmin, v.Code, false); err != nil {
		t.Fatal(err)
	}
	p, err := s.GetPayment(ctx, a, b.ID, true)
	if err != nil || p.Amount != 51000 || p.Total != 170000 || p.VoucherCode != "PROMO50" {
		t.Fatal("QR did not use discounted snapshot", err)
	}
	if _, err = s.syncPayment(ctx, transfer(p)); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SetStatus(ctx, a, b.ID, "cancelled"); err != nil {
		t.Fatal(err)
	}
	list, err = s.Vouchers(ctx, voucherAdmin)
	if err != nil || list[0].UsedCount != 1 {
		t.Fatal("paid cancellation returned voucher use", err)
	}
	if _, err = s.SetVoucherActive(ctx, voucherAdmin, v.Code, true); err != nil {
		t.Fatal(err)
	}
	in.Hour = 21
	if _, err = s.Create(ctx, a, in, false); err == nil {
		t.Fatal("paid voucher was reused by same user")
	}
}

func TestVoucherRollbackAndUnpaidCancellation(t *testing.T) {
	s, a, original, _ := paymentFixture(t)
	ctx := context.Background()
	v, err := s.CreateVoucher(ctx, voucherAdmin, voucherInput("RETURN50"))
	if err != nil {
		t.Fatal(err)
	}
	in := CreateInput{CourtID: original.CourtID, Date: original.Date, Hour: 19, Name: "Voucher Player", Phone: "0901234567", VoucherCode: v.Code}
	if _, err = s.Create(ctx, a, in, false); err == nil {
		t.Fatal("occupied slot accepted")
	}
	list, err := s.Vouchers(ctx, voucherAdmin)
	if err != nil || list[0].UsedCount != 0 {
		t.Fatal("failed booking consumed quota", err)
	}
	in.Hour = 20
	b, err := s.Create(ctx, a, in, false)
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.GetPayment(ctx, a, b.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Quote(ctx, a, QuoteInput{CourtID: in.CourtID, Date: in.Date, Hour: 21, VoucherCode: v.Code}); err == nil {
		t.Fatal("reserved voucher previewed twice for same user")
	}
	if _, err = s.SetStatus(ctx, a, b.ID, "cancelled"); err != nil {
		t.Fatal(err)
	}
	list, err = s.Vouchers(ctx, voucherAdmin)
	if err != nil || list[0].UsedCount != 0 {
		t.Fatal("unpaid cancellation did not release quota", err)
	}
	in.Hour = 21
	replacement, err := s.Create(ctx, a, in, false)
	if err != nil {
		t.Fatal("released voucher cannot be reused", err)
	}
	if _, err = s.syncPayment(ctx, transfer(p)); err != nil {
		t.Fatal(err)
	}
	old, err := s.getBooking(ctx, b.ID)
	if err != nil || old.Status != "cancelled" || old.DepositPaid {
		t.Fatal("late payment revived old voucher booking", err)
	}
	newBooking, err := s.getBooking(ctx, replacement.ID)
	if err != nil || newBooking.Discount != 50000 || newBooking.DepositPaid {
		t.Fatal("late old callback affected replacement booking", err)
	}
	list, err = s.Vouchers(ctx, voucherAdmin)
	if err != nil || list[0].UsedCount != 1 {
		t.Fatal("incorrect quota after late callback", err)
	}
}

func TestVoucherGlobalQuotaConcurrent(t *testing.T) {
	s, _, original, _ := paymentFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	v := voucherInput("LASTONE")
	v.MaxUses = 1
	if _, err := s.CreateVoucher(ctx, voucherAdmin, v); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 12)
	date, _ := parseDate(original.Date)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Create(ctx, httpx.Actor{ID: fmt.Sprintf("user-%d", i), Role: "user"}, CreateInput{CourtID: original.CourtID, Date: date.AddDate(0, 0, i+1).Format("2006-01-02"), Hour: 20, Name: "Concurrent Player", Phone: "0901234567", VoucherCode: v.Code}, false)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if public, ok := err.(*httpx.Error); !ok || public.Status != 409 {
			t.Fatal("unexpected booking failure", err)
		}
	}
	if success != 1 {
		t.Fatalf("last use granted %d times", success)
	}
	list, err := s.Vouchers(ctx, voucherAdmin)
	if err != nil || list[0].UsedCount != 1 {
		t.Fatal("quota overrun", err)
	}
}

func TestVoucherValidityAndConditions(t *testing.T) {
	s, a, original, _ := paymentFixture(t)
	ctx := context.Background()
	for _, name := range []string{"DISABLED", "FUTURE", "EXPIRED", "MINIMUM"} {
		v := voucherInput(name)
		if name == "FUTURE" {
			v.StartsAt = time.Now().Add(time.Hour).Unix()
			v.EndsAt = time.Now().Add(2 * time.Hour).Unix()
		}
		if name == "MINIMUM" {
			v.MinTotal = 500000
		}
		if _, err := s.CreateVoucher(ctx, voucherAdmin, v); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.SetVoucherActive(ctx, voucherAdmin, "DISABLED", false); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`UPDATE booking_vouchers SET starts_at=$1,ends_at=$2 WHERE code='EXPIRED'`, time.Now().Add(-time.Hour).Unix(), time.Now().Add(-time.Second).Unix()); err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"DISABLED", "FUTURE", "EXPIRED", "MINIMUM", "UNKNOWN", "BAD-CODE"} {
		if _, err := s.Quote(ctx, a, QuoteInput{CourtID: original.CourtID, Date: original.Date, Hour: 20, VoucherCode: code}); err == nil {
			t.Fatalf("invalid voucher %s accepted", code)
		}
	}
}

func TestVoucherOneUsePerUserConcurrent(t *testing.T) {
	s, a, original, _ := paymentFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	v := voucherInput("ONEPERUSER")
	if _, err := s.CreateVoucher(ctx, voucherAdmin, v); err != nil {
		t.Fatal(err)
	}
	date, _ := parseDate(original.Date)
	var wg sync.WaitGroup
	results := make(chan error, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Create(ctx, a, CreateInput{CourtID: original.CourtID, Date: date.AddDate(0, 0, i+1).Format("2006-01-02"), Hour: 20, Name: "Concurrent Player", Phone: "0901234567", VoucherCode: v.Code}, false)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if public, ok := err.(*httpx.Error); !ok || public.Status != 409 {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatalf("same user received %d uses", success)
	}
}

func TestVoucherMigrationPreservesExistingPrices(t *testing.T) {
	db, _ := testutil.Postgres(t)
	if err := database.Migrate(db, "booking_pg_v1", schema); err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, "booking_payment_pg_v1", paymentSchema); err != nil {
		t.Fatal(err)
	}
	_, err := db.Exec(`INSERT INTO booking_bookings(id,user_id,venue_id,court_id,date,hour,name,phone,total,deposit,status,created_at) VALUES('legacy','user','venue','court','2030-01-01',19,'Legacy Player','0901234567',220000,66000,'confirmed','2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(db, nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.getBooking(context.Background(), "legacy")
	if err != nil || b.Subtotal != 220000 || b.Total != 220000 || b.Deposit != 66000 || b.Discount != 0 || b.VoucherCode != "" {
		t.Fatal("existing booking price changed", err)
	}
}
