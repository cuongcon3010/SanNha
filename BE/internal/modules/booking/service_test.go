package booking

import (
	"sannha/be/internal/modules/venue"
	"testing"
)

func TestMoneyAndCalendar(t *testing.T) {
	c := venue.Court{DayPrice: 180000, EveningPrice: 220000}
	if Price(c, 16) != 180000 || Price(c, 17) != 220000 || Deposit(220000) != 66000 || Deposit(150001) != 45000 {
		t.Fatal("incorrect integer price/deposit")
	}
	for _, value := range []string{"2026-02-30", "2026-02-29", "2026-13-01", "2026-1-1", "not-a-date"} {
		if _, err := parseDate(value); err == nil {
			t.Fatalf("invalid date accepted: %s", value)
		}
	}
	if _, err := parseDate("2028-02-29"); err != nil {
		t.Fatal(err)
	}
}
