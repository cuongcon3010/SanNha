# Google + GitHub OAuth

OAuth chạy trong module `identity`, lưu tài khoản liên kết và state một lần trong PostgreSQL. Dùng authorization code + PKCE S256, cookie HttpOnly để ràng buộc state với trình duyệt, thời hạn 10 phút. Token Google/GitHub không lưu vào DB hoặc gửi cho FE. Sau đăng nhập, BE tạo session cookie hiện có, chuyển tới FE `/tai-khoan`; trang này gọi `/me` và hỗ trợ logout thật.

## Cấu hình

Trong BE environment (xem `.env.example`):

```dotenv
OAUTH_API_ORIGIN=http://localhost:8080
FRONTEND_ORIGIN=http://localhost:3000
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
```

Trong `FE/.env.local`:

```dotenv
NEXT_PUBLIC_API_ORIGIN=http://localhost:8080
```

Không đưa Client Secret vào biến `NEXT_PUBLIC_*`, code hoặc chat. Khởi động lại BE/FE sau khi thay environment. Provider thiếu cả ID và secret sẽ bị tắt, không hiển thị nút; chỉ thiếu một trường thì BE từ chối khởi động.

### Google

Tạo OAuth Client loại **Web application** trong Google Cloud Console, cấu hình consent screen và test users nếu app đang Testing. Authorized redirect URI phải chính xác:

`http://localhost:8080/api/v1/auth/oauth/google/callback`

Scopes: `openid email profile`. Yêu cầu email đã xác minh. Danh tính lấy từ Google userinfo qua access token server-side, dùng `sub` làm ID ổn định; không giải mã hoặc tin ID token chưa xác minh.

### GitHub

Tạo OAuth App trong GitHub Settings → Developer settings. Homepage URL là FE, callback URL:

`http://localhost:8080/api/v1/auth/oauth/github/callback`

Scopes: `read:user user:email`. Dùng GitHub numeric ID, yêu cầu primary email đã verified (hỗ trợ email ẩn qua `/user/emails`).

## Quy tắc tài khoản

- Lần đầu tạo user role `user`, không tự tạo owner/admin. Tài khoản OAuth có password ngẫu nhiên bị bỏ đi, không có password mặc định để đăng nhập.
- Lần sau tra `(provider, subject)`, không dùng email làm khóa danh tính.
- Email đã có ở tài khoản khác: từ chối với `account_exists`, yêu cầu đăng nhập bằng phương thức ban đầu. **Chưa có account linking, unlinking hoặc chuyển sang mật khẩu**; không tự ghép vì email trùng.
- User bị khóa không được đăng nhập bằng OAuth. Session dùng cơ chế bảo vệ và logout hiện có.
- State dùng một lần, lưu hash state/browser, lưu PKCE verifier ngắn hạn. Callback thiếu cookie, sai provider, hết hạn hoặc replay bị từ chối. URL redirect về FE là cố định, không nhận `returnTo` từ client.
- Production cần HTTPS, `COOKIE_SECURE=true`, FE/API cùng site; thay origin và redirect URI ở cả app và console từng provider. Không dùng URL nội bộ container làm callback public.

## Routes

- `GET /api/v1/auth/oauth/providers`: danh sách provider đã bật.
- `GET /api/v1/auth/oauth/{google|github}/start`: điều hướng trình duyệt tới provider.
- `GET /api/v1/auth/oauth/{google|github}/callback`: callback server-side.

OAuth, `/tai-khoan` và luồng VietQR mới `/dat-san`, `/thanh-toan/[id]` dùng xác thực backend thật. Form mật khẩu, chi tiết sân fixture và dashboard FE vẫn là demo; chọn role demo không cấp quyền thật. Backend password login vẫn có ở API nhưng chưa nối với form mật khẩu FE. Xem [VIETQR.md](VIETQR.md) để cấu hình thanh toán.

## Test

`go test ./...` kiểm tra cấu hình provider và parsing hồ sơ qua HTTP server giả. Với `TEST_DATABASE_URL` riêng, `make test` kiểm tra state/browser binding, PKCE, từ chối callback replay/hết hạn, tạo session, trở lại cùng tài khoản, không ghép email và khóa user.

Chưa kiểm chứng đăng nhập end-to-end với Google/GitHub thật khi chưa có Client ID/Secret. Không dùng credentials giả để khẳng định provider thật hoạt động.
