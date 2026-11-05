package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config credentials are server-only. Callback credentials are independent from VietQR API credentials.
type PaymentConfig struct {
	Enabled          bool
	Environment      string
	Username         string
	Password         string
	CallbackUsername string
	CallbackPassword string
	BankCode         string
	BankAccount      string
	BankName         string
}

func (c PaymentConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Environment != "sandbox" && c.Environment != "production" {
		return fmt.Errorf("VIETQR_ENV must be sandbox or production")
	}
	if c.Username == "" || c.Password == "" || c.CallbackUsername == "" || len(c.CallbackPassword) < 32 {
		return fmt.Errorf("VietQR requires API credentials and separate callback credentials (callback password at least 32 bytes)")
	}
	if c.Username == c.CallbackUsername && c.Password == c.CallbackPassword {
		return fmt.Errorf("VietQR callback credentials must differ from API credentials")
	}
	if !regexp.MustCompile(`^[A-Z0-9]{2,12}$`).MatchString(c.BankCode) || !regexp.MustCompile(`^[0-9]{4,32}$`).MatchString(c.BankAccount) || !regexp.MustCompile(`^[A-Z0-9 ]{2,100}$`).MatchString(c.BankName) {
		return fmt.Errorf("VietQR receiving bank code, account or unaccented uppercase account name is invalid")
	}
	return nil
}

type vietQRClient struct {
	config  PaymentConfig
	baseURL string
	http    *http.Client
	mu      sync.Mutex
	token   string
	expires time.Time
}

func newVietQRClient(c PaymentConfig) *vietQRClient {
	base := "https://dev.vietqr.org/vqr/api"
	if c.Environment == "production" {
		base = "https://api.vietqr.org/vqr/api"
	}
	return &vietQRClient{config: c, baseURL: base, http: &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (v *vietQRClient) accessToken(ctx context.Context) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.token != "" && time.Now().Before(v.expires) {
		return v.token, nil
	}
	r, err := http.NewRequestWithContext(ctx, "POST", v.baseURL+"/token_generate", nil)
	if err != nil {
		return "", err
	}
	r.SetBasicAuth(v.config.Username, v.config.Password)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	response, err := v.http.Do(r)
	if err != nil {
		return "", errors.New("VietQR token service unavailable")
	}
	defer response.Body.Close()
	var out struct {
		Token   string `json:"access_token"`
		Type    string `json:"token_type"`
		Expires int    `json:"expires_in"`
	}
	if response.StatusCode != 200 || json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&out) != nil || out.Token == "" || !strings.EqualFold(out.Type, "Bearer") || out.Expires < 30 || out.Expires > 86400 {
		return "", errors.New("invalid VietQR token response")
	}
	v.token = out.Token
	v.expires = time.Now().Add(time.Duration(out.Expires-15) * time.Second)
	return v.token, nil
}
func (v *vietQRClient) Generate(ctx context.Context, p Payment) (string, error) {
	token, err := v.accessToken(ctx)
	if err != nil {
		return "", err
	}
	payload := struct {
		BankCode    string `json:"bankCode"`
		BankAccount string `json:"bankAccount"`
		BankName    string `json:"userBankName"`
		Content     string `json:"content"`
		QRType      int    `json:"qrType"`
		Amount      int64  `json:"amount"`
		OrderID     string `json:"orderId"`
		TransType   string `json:"transType"`
	}{p.BankCode, p.BankAccount, p.BankName, p.Content, 0, p.Amount, p.OrderID, "C"}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	r, err := http.NewRequestWithContext(ctx, "POST", v.baseURL+"/qr/generate-customer", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	response, err := v.http.Do(r)
	if err != nil {
		return "", errors.New("VietQR QR service unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode == 401 {
		v.mu.Lock()
		v.token = ""
		v.mu.Unlock()
	}
	var out struct {
		BankCode    string      `json:"bankCode"`
		BankAccount string      `json:"bankAccount"`
		Amount      json.Number `json:"amount"`
		OrderID     string      `json:"orderId"`
		Content     string      `json:"content"`
		QR          string      `json:"qrCode"`
	}
	if response.StatusCode != 200 || json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&out) != nil {
		return "", errors.New("invalid VietQR QR response")
	}
	amount, err := strconv.ParseInt(string(out.Amount), 10, 64)
	if err != nil || amount != p.Amount || out.BankAccount != p.BankAccount || out.BankCode != p.BankCode || out.OrderID != p.OrderID || out.Content == "" || len(out.Content) > 100 || len(out.QR) < 30 || len(out.QR) > 4000 {
		return "", errors.New("VietQR QR details do not match payment")
	}
	// content may be normalized by VietQR; orderId and receiver/amount remain authoritative.
	return out.QR, nil
}
