package venue

import (
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"sannha/be/internal/platform/database"
	"sannha/be/internal/platform/httpx"
)

//go:embed schema.postgres.sql
var schema string

type Service struct{ db *sql.DB }
type Court struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DayPrice     int64  `json:"dayPrice"`
	EveningPrice int64  `json:"eveningPrice"`
	Maintenance  bool   `json:"maintenance"`
}
type Venue struct {
	ID            string   `json:"id"`
	OwnerID       string   `json:"ownerId"`
	Name          string   `json:"name"`
	Sport         string   `json:"sport"`
	City          string   `json:"city"`
	Address       string   `json:"address"`
	Amenities     []string `json:"amenities"`
	Courts        []Court  `json:"courts"`
	Status        string   `json:"status"`
	ReviewComment string   `json:"reviewComment"`
	CreatedAt     string   `json:"createdAt"`
}
type CourtInput struct {
	Name         string `json:"name"`
	DayPrice     int64  `json:"dayPrice"`
	EveningPrice int64  `json:"eveningPrice"`
}
type CreateInput struct {
	Name      string       `json:"name"`
	Sport     string       `json:"sport"`
	City      string       `json:"city"`
	Address   string       `json:"address"`
	Amenities []string     `json:"amenities"`
	Courts    []CourtInput `json:"courts"`
}

// CourtInfo is the public contract consumed by the booking module, never its tables.
type CourtInfo struct {
	Court
	VenueID     string
	OwnerID     string
	VenueStatus string
}
type RatesInput struct {
	DayPrice     int64 `json:"dayPrice"`
	EveningPrice int64 `json:"eveningPrice"`
	Maintenance  *bool `json:"maintenance"`
}

func New(db *sql.DB) (*Service, error) {
	if err := database.Migrate(db, "venue_pg_v1", schema); err != nil {
		return nil, err
	}
	return &Service{db}, nil
}
func validText(value string, min, max int) bool {
	n := utf8.RuneCountInString(value)
	return n >= min && n <= max
}
func validPrice(value int64) bool { return value >= 1000 && value <= 10000000 }
func (s *Service) Create(ctx context.Context, a httpx.Actor, in CreateInput) (Venue, error) {
	if a.Role != "owner" {
		return Venue{}, httpx.Forbidden()
	}
	in.Name = strings.TrimSpace(in.Name)
	in.City = strings.TrimSpace(in.City)
	in.Address = strings.TrimSpace(in.Address)
	sportOK := false
	for _, sport := range []string{"Bóng đá", "Cầu lông", "Pickleball", "Tennis", "Bóng rổ"} {
		if in.Sport == sport {
			sportOK = true
		}
	}
	if !validText(in.Name, 2, 120) || !validText(in.City, 2, 80) || !validText(in.Address, 5, 300) || !sportOK || len(in.Courts) < 1 || len(in.Courts) > 20 || len(in.Amenities) > 12 {
		return Venue{}, httpx.Invalid("Thông tin sân, bộ môn hoặc số sân con không hợp lệ.")
	}
	for _, amenity := range in.Amenities {
		if !validText(amenity, 1, 80) {
			return Venue{}, httpx.Invalid("Tiện ích không hợp lệ.")
		}
	}
	names := map[string]bool{}
	for i := range in.Courts {
		c := &in.Courts[i]
		c.Name = strings.TrimSpace(c.Name)
		if !validText(c.Name, 1, 80) || !validPrice(c.DayPrice) || !validPrice(c.EveningPrice) || names[c.Name] {
			return Venue{}, httpx.Invalid("Tên sân con trùng, rỗng hoặc giá ngoài khoảng 1.000–10.000.000đ.")
		}
		names[c.Name] = true
	}
	if in.Amenities == nil {
		in.Amenities = []string{}
	}
	amenities, err := json.Marshal(in.Amenities)
	if err != nil {
		return Venue{}, err
	}
	id := rand.Text()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Venue{}, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO venue_venues(id,owner_id,name,sport,city,address,amenities,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, id, a.ID, in.Name, in.Sport, in.City, in.Address, string(amenities), time.Now().UTC().Format(time.RFC3339)); err != nil {
		return Venue{}, err
	}
	for _, c := range in.Courts {
		if _, err = tx.ExecContext(ctx, `INSERT INTO venue_courts(id,venue_id,name,day_price,evening_price) VALUES($1,$2,$3,$4,$5)`, rand.Text(), id, c.Name, c.DayPrice, c.EveningPrice); err != nil {
			return Venue{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Venue{}, err
	}
	return s.Get(ctx, id)
}

const columns = "id,owner_id,name,sport,city,address,amenities,status,review_comment,created_at"

type scanner interface{ Scan(...any) error }

func scan(row scanner) (Venue, error) {
	var v Venue
	var amenities string
	err := row.Scan(&v.ID, &v.OwnerID, &v.Name, &v.Sport, &v.City, &v.Address, &amenities, &v.Status, &v.ReviewComment, &v.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return v, httpx.NotFound()
	}
	if err != nil {
		return v, err
	}
	err = json.Unmarshal([]byte(amenities), &v.Amenities)
	return v, err
}
func (s *Service) Get(ctx context.Context, id string) (Venue, error) {
	v, err := scan(s.db.QueryRowContext(ctx, "SELECT "+columns+" FROM venue_venues WHERE id=$1", id))
	if err != nil {
		return v, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,day_price,evening_price,maintenance FROM venue_courts WHERE venue_id=$1 ORDER BY name`, id)
	if err != nil {
		return v, err
	}
	defer rows.Close()
	v.Courts = []Court{}
	for rows.Next() {
		var c Court
		if err = rows.Scan(&c.ID, &c.Name, &c.DayPrice, &c.EveningPrice, &c.Maintenance); err != nil {
			return v, err
		}
		v.Courts = append(v.Courts, c)
	}
	return v, rows.Err()
}
func (s *Service) Court(ctx context.Context, id string) (CourtInfo, error) {
	var c CourtInfo
	err := s.db.QueryRowContext(ctx, `SELECT c.id,c.name,c.day_price,c.evening_price,c.maintenance,v.id,v.owner_id,v.status FROM venue_courts c JOIN venue_venues v ON v.id=c.venue_id WHERE c.id=$1`, id).Scan(&c.ID, &c.Name, &c.DayPrice, &c.EveningPrice, &c.Maintenance, &c.VenueID, &c.OwnerID, &c.VenueStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return c, httpx.NotFound()
	}
	return c, err
}
func (s *Service) List(ctx context.Context, ownerID, status string) ([]Venue, error) {
	// shortcut: Lists are capped at 200; add cursor pagination before the catalog grows beyond this initial API.
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM venue_venues WHERE ($1='' OR owner_id=$1) AND ($2='' OR status=$2) ORDER BY created_at DESC LIMIT 200`, ownerID, status)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	result := []Venue{}
	for _, id := range ids {
		v, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}
func (s *Service) UpdateRates(ctx context.Context, a httpx.Actor, venueID, courtID string, in RatesInput) (Court, error) {
	c, err := s.Court(ctx, courtID)
	if err != nil {
		return Court{}, err
	}
	if a.Role != "owner" || a.ID != c.OwnerID {
		return Court{}, httpx.Forbidden()
	}
	if c.VenueID != venueID {
		return Court{}, httpx.NotFound()
	}
	if !validPrice(in.DayPrice) || !validPrice(in.EveningPrice) || in.Maintenance == nil {
		return Court{}, httpx.Invalid("Cần giá hợp lệ và trạng thái maintenance.")
	}
	_, err = s.db.ExecContext(ctx, `UPDATE venue_courts SET day_price=$1,evening_price=$2,maintenance=$3 WHERE id=$4`, in.DayPrice, in.EveningPrice, *in.Maintenance, courtID)
	if err != nil {
		return Court{}, err
	}
	c.DayPrice = in.DayPrice
	c.EveningPrice = in.EveningPrice
	c.Maintenance = *in.Maintenance
	return c.Court, nil
}
func (s *Service) Review(ctx context.Context, a httpx.Actor, id, status, comment string) (Venue, error) {
	if a.Role != "admin" {
		return Venue{}, httpx.Forbidden()
	}
	comment = strings.TrimSpace(comment)
	if status != "approved" && status != "rejected" || len(comment) > 1000 || status == "rejected" && !validText(comment, 10, 1000) {
		return Venue{}, httpx.Invalid("Trạng thái không hợp lệ; từ chối cần lý do từ 10 ký tự.")
	}
	res, err := s.db.ExecContext(ctx, `UPDATE venue_venues SET status=$1,review_comment=$2 WHERE id=$3`, status, comment, id)
	if err != nil {
		return Venue{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Venue{}, err
	}
	if n == 0 {
		return Venue{}, httpx.NotFound()
	}
	return s.Get(ctx, id)
}
func (s *Service) Favorite(ctx context.Context, userID, venueID string, add bool) error {
	if !add {
		_, err := s.db.ExecContext(ctx, `DELETE FROM venue_favorites WHERE user_id=$1 AND venue_id=$2`, userID, venueID)
		return err
	}
	v, err := s.Get(ctx, venueID)
	if err != nil {
		return err
	}
	if v.Status != "approved" {
		return httpx.NotFound()
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO venue_favorites(user_id,venue_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, userID, venueID)
	return err
}
func (s *Service) Favorites(ctx context.Context, userID string) ([]Venue, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT v.id FROM venue_venues v JOIN venue_favorites f ON f.venue_id=v.id WHERE f.user_id=$1 AND v.status='approved' ORDER BY v.name LIMIT 200`, userID)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	out := []Venue{}
	for _, id := range ids {
		v, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}
