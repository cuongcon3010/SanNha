package identity

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"sannha/be/internal/platform/database"
	"sannha/be/internal/platform/httpx"
)

//go:embed schema.postgres.sql
var schema string

type Service struct{ db *sql.DB }
type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Area      string `json:"area"`
	Role      string `json:"role"`
	Locked    bool   `json:"locked"`
	CreatedAt string `json:"createdAt"`
}
type Registration struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
type Profile struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Area  string `json:"area"`
}

var phonePattern = regexp.MustCompile(`^(0[0-9]{9}|\+84[0-9]{9})$`)

const userColumns = "id,name,email,phone,area,role,locked,created_at"
const sessionLifetime = 7 * 24 * time.Hour

func New(db *sql.DB) (*Service, error) {
	if err := database.Migrate(db, "identity_pg_v1", schema); err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}
func normalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value || len(value) > 254 || !strings.Contains(value, ".") {
		return "", httpx.Invalid("Email không hợp lệ.")
	}
	return value, nil
}
func passwordHash(password string) (string, error) {
	if utf8.RuneCountInString(password) < 12 || len(password) > 128 {
		return "", httpx.Invalid("Mật khẩu cần từ 12 ký tự đến 128 byte.")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, 600000, 32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(key), nil
}
func verifyPassword(password, stored string) bool {
	parts := strings.Split(stored, ":")
	if len(parts) != 2 || len(password) > 128 {
		return false
	}
	salt, e1 := hex.DecodeString(parts[0])
	want, e2 := hex.DecodeString(parts[1])
	if e1 != nil || e2 != nil || len(salt) != 16 || len(want) != 32 {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, 600000, 32)
	return err == nil && subtle.ConstantTimeCompare(got, want) == 1
}
func (s *Service) Register(ctx context.Context, in Registration) (User, error) {
	if in.Role == "" {
		in.Role = "user"
	}
	if in.Role != "user" && in.Role != "owner" {
		return User{}, httpx.Invalid("Chỉ có thể đăng ký người chơi hoặc chủ sân.")
	}
	return s.create(ctx, in)
}
func (s *Service) create(ctx context.Context, in Registration) (User, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Phone = strings.TrimSpace(in.Phone)
	if utf8.RuneCountInString(in.Name) < 2 || utf8.RuneCountInString(in.Name) > 80 {
		return User{}, httpx.Invalid("Họ tên phải có 2–80 ký tự.")
	}
	if in.Phone != "" && !phonePattern.MatchString(in.Phone) {
		return User{}, httpx.Invalid("Số điện thoại không hợp lệ.")
	}
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return User{}, err
	}
	hash, err := passwordHash(in.Password)
	if err != nil {
		return User{}, err
	}
	u := User{ID: rand.Text(), Name: in.Name, Email: email, Phone: in.Phone, Role: in.Role, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	_, err = s.db.ExecContext(ctx, `INSERT INTO identity_users(id,name,email,phone,password_hash,role,created_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, u.ID, u.Name, u.Email, u.Phone, hash, u.Role, u.CreatedAt)
	if database.IsUniqueViolation(err, "identity_users_email_key") {
		return User{}, httpx.Conflict("Email đã được đăng ký.")
	}
	return u, err
}
func (s *Service) BootstrapAdmin(ctx context.Context, email, password string) (User, error) {
	return s.create(ctx, Registration{Name: "Quản trị viên", Email: email, Password: password, Role: "admin"})
}

type scanner interface{ Scan(...any) error }

func scanUser(row scanner) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Area, &u.Role, &u.Locked, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, httpx.NotFound()
	}
	return u, err
}
func (s *Service) Get(ctx context.Context, id string) (User, error) {
	return scanUser(s.db.QueryRowContext(ctx, "SELECT "+userColumns+" FROM identity_users WHERE id=$1", id))
}
func tokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var id, hash string
	var locked bool
	err := s.db.QueryRowContext(ctx, `SELECT id,password_hash,locked FROM identity_users WHERE lower(email)=$1`, email).Scan(&id, &hash, &locked)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return User{}, "", err
	}
	if errors.Is(err, sql.ErrNoRows) {
		hash = strings.Repeat("0", 32) + ":" + strings.Repeat("0", 64)
	}
	valid := verifyPassword(password, hash)
	if err != nil || !valid || locked {
		return User{}, "", &httpx.Error{Status: 401, Message: "Email hoặc mật khẩu không đúng, hoặc tài khoản bị khóa."}
	}
	token := rand.Text() + rand.Text()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, "", err
	}
	defer tx.Rollback()
	// Serialize session creation with account locking and its session revocation.
	if err = tx.QueryRowContext(ctx, `SELECT locked FROM identity_users WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
		return User{}, "", err
	}
	if locked {
		return User{}, "", httpx.Unauthorized()
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM identity_sessions WHERE expires_at<=$1`, time.Now().Unix()); err != nil {
		return User{}, "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO identity_sessions(token_hash,user_id,expires_at) VALUES($1,$2,$3)`, tokenHash(token), id, time.Now().Add(sessionLifetime).Unix())
	if err != nil {
		return User{}, "", err
	}
	if err = tx.Commit(); err != nil {
		return User{}, "", err
	}
	u, err := s.Get(ctx, id)
	return u, token, err
}
func (s *Service) Authenticate(ctx context.Context, token string) (httpx.Actor, error) {
	if token == "" || len(token) > 128 {
		return httpx.Actor{}, httpx.Unauthorized()
	}
	var a httpx.Actor
	err := s.db.QueryRowContext(ctx, `SELECT u.id,u.role FROM identity_sessions s JOIN identity_users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.expires_at>$2 AND u.locked=FALSE`, tokenHash(token), time.Now().Unix()).Scan(&a.ID, &a.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return a, httpx.Unauthorized()
	}
	return a, err
}
func (s *Service) Logout(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM identity_sessions WHERE token_hash=$1`, tokenHash(token))
	return err
}
func (s *Service) UpdateProfile(ctx context.Context, id string, in Profile) (User, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Area = strings.TrimSpace(in.Area)
	if utf8.RuneCountInString(in.Name) < 2 || utf8.RuneCountInString(in.Name) > 80 || len(in.Area) > 300 || !phonePattern.MatchString(in.Phone) {
		return User{}, httpx.Invalid("Tên, điện thoại hoặc khu vực không hợp lệ.")
	}
	_, err := s.db.ExecContext(ctx, `UPDATE identity_users SET name=$1,phone=$2,area=$3 WHERE id=$4`, in.Name, in.Phone, in.Area, id)
	if err != nil {
		return User{}, err
	}
	return s.Get(ctx, id)
}
func (s *Service) Users(ctx context.Context) ([]User, error) {
	// shortcut: Admin lists return at most 200 users; add pagination when the account catalog grows.
	rows, err := s.db.QueryContext(ctx, "SELECT "+userColumns+" FROM identity_users ORDER BY created_at DESC LIMIT 200")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, rows.Err()
}
func (s *Service) Lock(ctx context.Context, actor httpx.Actor, id string, locked bool) (User, error) {
	if actor.Role != "admin" {
		return User{}, httpx.Forbidden()
	}
	u, err := s.Get(ctx, id)
	if err != nil {
		return User{}, err
	}
	if u.Role == "admin" {
		return User{}, httpx.Invalid("Không được khóa quản trị viên qua API này.")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `UPDATE identity_users SET locked=$1 WHERE id=$2`, locked, id); err != nil {
		return User{}, err
	}
	if locked {
		if _, err = tx.ExecContext(ctx, `DELETE FROM identity_sessions WHERE user_id=$1`, id); err != nil {
			return User{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return User{}, err
	}
	u.Locked = locked
	return u, nil
}
