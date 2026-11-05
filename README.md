# Sân Nhà

Voucher: quản trị thật `/quan-tri/vouchers`, nhập mã tại `/dat-san`, cọc VietQR tính trên giá sau giảm. Hướng dẫn và giới hạn: [BE/VOUCHERS.md](BE/VOUCHERS.md).

Đăng nhập Google/GitHub: xem [BE/OAUTH.md](BE/OAUTH.md). Luồng OAuth, `/tai-khoan`, `/dat-san` và `/thanh-toan/[id]` dùng session backend; các dashboard demo còn lại chưa tích hợp.

VietQR/Tingo Pay: xem [BE/VIETQR.md](BE/VIETQR.md). `/dat-san` và `/thanh-toan/[id]` dùng API thật, tự xác nhận cọc từ callback đã đối soát. Mặc định tắt/sandbox, cần đăng ký merchant và cấu hình callback HTTPS; không thu tiền cho booking localStorage demo.

```text
FE/   Next.js 16 + React + TypeScript — giao diện từ Penpot
BE/   Go — backend Modulith, một tiến trình và một PostgreSQL database
```

## Chạy frontend

Yêu cầu Node.js 24+, npm 11+.

```bash
cd FE
npm ci
npm run dev
```

Mở http://localhost:3000. Chi tiết trang, tài nguyên và kiểm tra: [FE/README.md](FE/README.md).

## Chạy backend

Yêu cầu Go 1.25+ và PostgreSQL 17+. Không cần CGO. Có `BE/compose.yaml` để chạy database local bằng Docker Compose.

```bash
cd BE
cp .env.example .env
# Sửa mật khẩu riêng trong POSTGRES_PASSWORD và DATABASE_URL của .env trước khi chạy.
docker compose up -d --wait postgres
set -a
. ./.env
set +a
go mod download
go run ./cmd/api
```

API mặc định: http://127.0.0.1:8080/api/v1/health. `DATABASE_URL` bắt buộc; có thể dùng database PostgreSQL sẵn có thay Docker. Database phải tồn tại trước khi chạy; API tự chạy migration bảng. Không có tài khoản hay mật khẩu mặc định. Dữ liệu SQLite cũ được giữ nguyên, không tự chuyển sang PostgreSQL.

Tạo admin bằng lệnh CLI và biến môi trường; xem [BE/README.md](BE/README.md). Hợp đồng API và ví dụ request: [BE/API.md](BE/API.md).

## Kiến trúc backend

- **identity**: đăng ký, đăng nhập, session cookie, hồ sơ, phân quyền, khóa tài khoản và thu hồi phiên.
- **venue**: sân, sân con, bảng giá, bảo trì, yêu thích và kiểm duyệt sân.
- **booking**: lịch đặt, kiểm tra giờ trống, chống trùng bằng unique index, khóa/mở giờ và báo cáo theo tháng.
- **app**: composition root, khởi tạo module và ghép HTTP routes. `booking` gọi `venue` qua interface `Catalog`; không đọc bảng của module khác.
- Mỗi module có migration riêng; password PBKDF2-SHA256, token phiên chỉ lưu hash; tiền là số nguyên VND và giữ nguyên giá ở thời điểm đặt.

## Kiểm tra

Trong `FE/`:

```bash
npm test
npm run lint
npm run typecheck
npm run build
```

Trong `BE/`:

```bash
make test
go vet ./...
go build ./...
```

`make test` yêu cầu `TEST_DATABASE_URL` trỏ database test PostgreSQL riêng; xem cách tạo trong `BE/README.md`. `make test-unit` chạy không cần DB và skip integration. Test backend bao gồm HTTP authorization, quyền sở hữu sân, thu hồi session, validation, kiểm duyệt, đổi giá, yêu thích, đặt trùng đồng thời qua hai connection pool, khóa/mở giờ, persistence, migration đồng thời/rollback và ranh giới module.

## Trạng thái tích hợp

**FE có luồng API thật riêng, không tự chuyển dữ liệu demo.** Google/GitHub OAuth, `/tai-khoan`, `/dat-san` và `/thanh-toan/[id]` dùng BE/PostgreSQL. Các trang sân fixture và dashboard `/user`, `/chu-san`, `/admin` vẫn dùng localStorage. Database ban đầu trống; cần có sân được duyệt và merchant sandbox để thử thanh toán.

Đã có tích hợp cọc VietQR/Tingo Pay nhưng chưa nghiệm thu gateway thật. Chưa có hoàn/chia tiền, gửi email/OTP, reset mật khẩu thực, upload ảnh hoặc triển khai production. Xem giới hạn bảo mật, cấu hình TLS và vận hành trong tài liệu BE trước khi đưa lên mạng.
