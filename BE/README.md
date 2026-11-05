# Sân Nhà backend — Go Modulith

Voucher giảm tiền sân và cọc VietQR sau giảm: [VOUCHERS.md](VOUCHERS.md). Admin tạo/tắt mã, server giữ quota và booking atomically.

Google/GitHub OAuth: cấu hình, callback URLs và giới hạn tại [OAUTH.md](OAUTH.md).

Thanh toán cọc VietQR/Tingo Pay tự đối soát: xem [VIETQR.md](VIETQR.md). Mặc định tắt, dùng một tài khoản merchant chung; sandbox trước, chưa nghiệm thu giao dịch thật.

Một HTTP API chạy trong **một tiến trình**, deploy một executable, dùng một PostgreSQL database. Các module được phân tách theo nghiệp vụ, không phải microservice và không chia package chỉ theo controller/repository.

## Cấu trúc và ranh giới

```text
cmd/api/                     # HTTP server, graceful shutdown
cmd/admin/                   # Tạo quản trị viên ngoài HTTP
internal/app/                # Composition root, config, integration/architecture tests
internal/modules/
  identity/                  # users, sessions, auth, profile, role, account lock
  venue/                     # venues, courts, prices, maintenance, favorites, review
  booking/                   # bookings, availability, blocks, revenue
internal/platform/
  database/                  # PostgreSQL connection pool + migration runner
  httpx/                     # JSON, errors, principal context, HTTP policy
internal/testutil/           # PostgreSQL schema riêng cho mỗi integration test
```

```text
cmd/api → app → identity
             → venue
             → booking ──Catalog interface──→ venue public contracts
```

- Module có `schema.postgres.sql`, service chứa quy tắc nghiệp vụ + SQL, HTTP adapter và tests. Không tạo repository interface nếu chưa có nhu cầu thay implementation.
- `identity_*`, `venue_*`, `booking_*` là bảng do từng module sở hữu. Chỉ module sở hữu được đọc/ghi trực tiếp. Không có cross-module SQL join hoặc foreign key; reference được xác minh qua API module. Sessions và courts/favorites có foreign key **bên trong** module.
- `booking.Catalog` là consumer-owned interface; `venue.Service` đáp ứng bằng `Court`/`Get`. `booking` chỉ import các public DTO của `venue`.
- Architecture test chặn dependency không được phép và việc tham chiếu bảng của module khác. Đây là bảo vệ ở mức package/source, không thay thế database permission nếu sau này tách service.
- Migration PostgreSQL (`identity_pg_v1`, `venue_pg_v1`, `booking_pg_v1`) chạy trong transaction, ghi tên vào `schema_migrations`, dùng transaction advisory lock để tránh hai instance chạy migration đồng thời. Khi nâng schema, thêm migration có tên mới, không sửa nội dung V1 đã áp dụng.

## Cài đặt & chạy

Go **1.25+**, PostgreSQL **17+**. Driver `github.com/jackc/pgx/v5` qua `database/sql`, không cần CGO/C compiler để build API. Docker Compose chỉ cần nếu muốn chạy PostgreSQL local bằng cấu hình có sẵn; cũng có thể dùng PostgreSQL đã cài hoặc dịch vụ managed.

```bash
cd BE
cp .env.example .env
# Sửa .env: thay CHANGE_ME trong POSTGRES_PASSWORD và DATABASE_URL bằng mật khẩu riêng.
docker compose up -d --wait postgres
set -a
. ./.env
set +a
go mod download
go run ./cmd/api
```

Hoặc `make db-up`, export environment rồi `make run`. API listen `127.0.0.1:8080`. `DATABASE_URL` bắt buộc; chương trình không tự tạo database, chỉ migrate bảng sau khi kết nối. Không có tài khoản hay mật khẩu mặc định. `compose.yaml` dùng PostgreSQL 17, publish cổng **chỉ ở 127.0.0.1:5432**, lưu dữ liệu trong named volume `postgres_data`. `docker compose down` giữ volume; không dùng `down -v` nếu cần giữ dữ liệu. Đổi mật khẩu trong `.env` không tự đổi mật khẩu của database đã khởi tạo.

Kiểm tra:

```bash
curl http://127.0.0.1:8080/api/v1/health
```

### Biến môi trường

`.env.example` chỉ là mẫu; chương trình đọc **environment**, không tự load `.env`.

| Biến                            | Mặc định                       | Ý nghĩa                                                                                                |
| ------------------------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------ |
| `HTTP_ADDR`                     | `127.0.0.1:8080`               | Địa chỉ listen                                                                                         |
| `DATABASE_URL`                  | Không có, bắt buộc             | URL PostgreSQL; credentials không được commit/log; production dùng `sslmode=verify-full` và CA phù hợp |
| `POSTGRES_DB` / `POSTGRES_USER` | `sannha`                       | Chỉ dùng cho Docker Compose local                                                                      |
| `POSTGRES_PASSWORD`             | Không có, bắt buộc với Compose | Mật khẩu riêng; `CHANGE_ME` chỉ là placeholder                                                         |
| `TEST_DATABASE_URL`             | Không có                       | Database test riêng; role cần quyền CREATE SCHEMA, không trỏ production                                |
| `FRONTEND_ORIGIN`               | `http://localhost:3000`        | Một origin CORS chính xác, không có dấu `/` cuối                                                       |
| `COOKIE_SECURE`                 | `false`                        | Local HTTP mới được `false`; HTTPS frontend bắt buộc `true`                                            |
| `ADMIN_EMAIL`                   | Không có                       | Chỉ dùng trong lệnh tạo admin                                                                          |
| `ADMIN_PASSWORD`                | Không có                       | 12 ký tự trở lên, tối đa 128 byte; không hard-code                                                     |

Chương trình không tự đọc `.env`; sau khi điền giá trị, export bằng shell trong terminal riêng. URL-encode ký tự đặc biệt trong phần mật khẩu của `DATABASE_URL` và đặt giá trị trong dấu nháy khi source bằng shell. `sslmode=disable` chỉ dành cho PostgreSQL local:

```bash
set -a
. ./.env
set +a
go run ./cmd/api
```

Không đặt mật khẩu thật trong shell history, code, log hoặc terminal chia sẻ. Lệnh tạo admin có thể đọc mật khẩu tương tác:

```bash
read -r -p 'Admin email: ' ADMIN_EMAIL
read -r -s -p 'Admin password: ' ADMIN_PASSWORD
export ADMIN_EMAIL ADMIN_PASSWORD
go run ./cmd/admin
unset ADMIN_EMAIL ADMIN_PASSWORD
```

Không có endpoint tự đăng ký admin. CLI chỉ tạo tài khoản mới, không tự nâng quyền hay đổi mật khẩu tài khoản đã tồn tại. API database mới trống; đăng ký owner, tạo sân và duyệt sân bằng admin trước khi đặt.

### Build / test / format

Tạo database test riêng (một lần với Compose):

```bash
docker compose exec postgres sh -c 'createdb -U "$POSTGRES_USER" sannha_test'
```

Điền `TEST_DATABASE_URL` vào `.env` với database `sannha_test`, export lại environment như trên, sau đó:

```bash
go build -o bin/api ./cmd/api
go build -o bin/admin ./cmd/admin
make test
go test ./internal/app -run TestConcurrentBookingAndBlockRelease -v
go vet ./...
gofmt -w cmd internal
```

`make test` yêu cầu `TEST_DATABASE_URL` rồi chạy `go test -race ./...`. Mỗi integration test tạo schema ngẫu nhiên, cấu hình `search_path` cho toàn bộ connection pool rồi xóa đúng schema đó sau test, không truncate bảng dùng chung. Không dùng database production cho test. Tests kiểm tra migration đồng thời/rollback, persistence khi mở lại kết nối, và 12 yêu cầu đặt cùng giờ từ hai pool chỉ có một yêu cầu thành công.

Chạy unit/architecture tests không cần DB: `make test-unit` (integration tests sẽ skip rõ ràng). Chạy `go test ./...` trực tiếp mà chưa đặt URL cũng skip integration, không được coi là đã kiểm thử database.

Makefile còn cung cấp `run`, `build`, `lint`, `fmt`, `db-up`, `db-down`. Có graceful shutdown 10 giây, timeouts HTTP, giới hạn header 16 KB/body JSON 64 KB.

## Quy tắc nghiệp vụ

- Đăng ký role `user` hoặc `owner`; admin chỉ tạo bằng CLI. Profile không thể thay role/email/locked.
- Session cookie HttpOnly, SameSite=Lax, hạn 7 ngày; token ngẫu nhiên chỉ lưu SHA-256 hash trong DB. Password dùng PBKDF2-HMAC-SHA256, 600.000 vòng, salt riêng. Mọi request bảo vệ đều kiểm tra user chưa bị khóa; khóa user xóa session; mở khóa không khôi phục phiên cũ.
- Tạo session khóa hàng user bằng `FOR UPDATE` để đồng bộ với thao tác khóa tài khoản/thu hồi phiên khi có nhiều connection hoặc instance.
- Chủ sân chỉ tạo/cập nhật sân của mình. Sân mới `pending`, chỉ công khai và nhận đặt sau khi admin duyệt `approved`. `rejected` cần lý do từ 10 ký tự.
- Lịch dài **1 giờ**, giờ bắt đầu từ **16 đến 21** theo FE hiện tại. Ngày `YYYY-MM-DD`, giờ Việt Nam UTC+7. Từ chối giờ đã qua.
- Giá 16:00 dùng `dayPrice`, 17:00–22:00 dùng `eveningPrice`. Backend không nhận `total` từ client. `total` và cọc 30% được snapshot; sửa giá không đổi lịch cũ.
- Lịch mới `pending`; owner/admin có thể xác nhận thủ công, không đồng nghĩa đã nhận tiền. Unique partial index bảo vệ `(court_id,date,hour)` cho mọi lịch chưa hủy, kể cả yêu cầu đồng thời. Hủy giải phóng giờ; hoàn thành không tự giải phóng slot.
- User chỉ xem/hủy lịch mình. Owner chỉ xem/quản lý lịch sân mình. Admin có quyền quản lý lịch qua API theo cùng state machine. Owner có thể tạo `blocked` với giá/cọc 0 và mở bằng `cancelled`.
- Public availability chỉ trả giờ/giá/trống, không trả tên/số điện thoại/mã lịch. Revenue là **giá trị lịch**, không phải tiền đã thu; callback cọc có event riêng trong module booking, chưa có ledger quyết toán/hoàn tiền.
- Catalog/favorites/users giới hạn 200 bản ghi, danh sách lịch 500; chưa có pagination. Availability tối đa 120 slot/ngày theo giới hạn 20 sân con. Revenue tổng hợp SQL toàn tháng, không bị cắt theo giới hạn list.

## Bảo mật và giới hạn vận hành

- CORS chỉ cho `FRONTEND_ORIGIN`, credentials bật, Origin khác bị từ chối. API dùng JSON và kiểm tra Origin/Fetch Metadata để giảm CSRF. Cookie Lax yêu cầu FE/API **cùng site** khi tích hợp trên trình duyệt (local nên dùng cùng hostname, ví dụ đều `localhost`).
- Login/register có giới hạn mỗi IP: 30 yêu cầu/15 phút mỗi endpoint, bộ nhớ bounded. Không tin `X-Forwarded-For`; nếu đặt sau proxy, phải bổ sung chính sách trusted proxy/rate limit ở edge để tránh toàn bộ khách dùng chung IP.
- **Không coi đây là hệ thống sẵn sàng production**: cần HTTPS/reverse proxy, backup DB/WAL, monitoring, audit log bền vững, chống abuse phân tán, chính sách hủy/hoàn cọc và cơ chế phục hồi tài khoản. VietQR cọc đã có tích hợp nhưng chưa nghiệm thu thật; chưa có gửi email, OTP, reset mật khẩu, upload file, hoàn/chia tiền hay Idempotency-Key cho toàn bộ API.
- Pool PostgreSQL mặc định tối đa 10 connections, 5 idle, lifetime 30 phút, connect timeout 5 giây. Cần tính tổng connections của mọi instance và giới hạn database khi triển khai. Giá/maintenance được đọc qua contract module tại lúc tạo yêu cầu; chưa có transaction chung cho thay đổi đồng thời giữa các module.
- Không tự seed tài khoản/credentials hoặc nhập fixture FE. Tests dùng dữ liệu ngẫu nhiên và schema riêng.

## Dữ liệu SQLite cũ

Ứng dụng chỉ chạy PostgreSQL, không còn dùng `DB_PATH` hoặc driver SQLite. Các file `schema.sqlite.sql` được giữ nguyên làm lịch sử V1, không được embed/chạy. File `data/sannha.db` cũ (nếu có) được giữ nguyên, không xóa và **không tự nhập vào PostgreSQL**. PostgreSQL mới bắt đầu trống. Nếu cần giữ dữ liệu cũ, phải thực hiện một đợt export/import có kiểm tra riêng, bao gồm email trùng không phân biệt hoa thường, boolean và unique slot; chưa có công cụ chuyển dữ liệu tự động.

## FE

FE trong `../FE/` dùng API thật cho OAuth, `/tai-khoan`, `/dat-san` và `/thanh-toan/[id]`, gửi `credentials: "include"` và kiểm tra `/me`. Form mật khẩu cùng dashboard demo vẫn dùng localStorage. Không giữ quyền hạn đáng tin trong localStorage; API kiểm tra owner/admin ở server, không chỉ ẩn menu trên FE.

Xem [API.md](API.md) để có request/response và danh sách endpoints.
