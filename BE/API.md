# HTTP API v1

Voucher: `POST /bookings/quote`, `GET/POST /admin/vouchers`, `PATCH /admin/vouchers/{code}`; tạo booking thêm `voucherCode`. Booking/payment có `subtotal`, `discount`, `voucherCode` và tổng/cọc sau giảm. Chi tiết: [VOUCHERS.md](VOUCHERS.md).

VietQR/Tingo Pay bổ sung payment config, QR cọc theo booking, callback Get Token/Transaction Sync và danh sách admin đối soát. Hợp đồng, xác thực và giới hạn tại [VIETQR.md](VIETQR.md). Booking responses thêm `depositPaid`; khi bật VietQR, chỉ callback hợp lệ được xác nhận cọc tự động.

Google/GitHub OAuth bổ sung: `GET /auth/oauth/providers`, `GET /auth/oauth/{provider}/start`, `GET /auth/oauth/{provider}/callback`. Callback dùng session cookie hiện có, không trả token provider cho FE. Hướng dẫn cấu hình tại [OAUTH.md](OAUTH.md).

Base URL local: `http://localhost:8080/api/v1`. Listen mặc định `127.0.0.1:8080`; dùng hostname `localhost` khi gọi từ FE `http://localhost:3000` để cookie cùng site.

Body POST/PATCH dùng `Content-Type: application/json`. JSON thừa trường, sai kiểu hoặc lớn hơn 64 KB bị từ chối. PUT/DELETE favorites không cần body. Errors có dạng `{"error":"Thông báo"}`; không trả SQL/stacktrace/password/token trong response JSON.

| HTTP            | Ý nghĩa                                                   |
| --------------- | --------------------------------------------------------- |
| 200 / 201 / 204 | Đọc/cập nhật, tạo, thao tác không có body response        |
| 400 / 415       | Dữ liệu/JSON không hợp lệ, sai Content-Type               |
| 401 / 403       | Chưa đăng nhập / không có quyền                           |
| 404             | Không có dữ liệu hoặc sân chưa được công khai             |
| 409             | Email/khung giờ trùng, bảo trì hoặc chuyển trạng thái sai |
| 429             | Quá giới hạn login/register                               |
| 500             | Lỗi nội bộ; chi tiết chỉ ghi log server                   |

## Tài khoản

| Method | Path                     | Quyền                                    |
| ------ | ------------------------ | ---------------------------------------- |
| GET    | `/health`                | Public                                   |
| POST   | `/auth/register`         | Public; role chỉ user/owner              |
| POST   | `/auth/login`            | Public                                   |
| POST   | `/auth/logout`           | Đăng nhập; body `{}`                     |
| GET    | `/me`                    | Đăng nhập                                |
| PATCH  | `/me`                    | Đăng nhập; chỉ name/phone/area           |
| GET    | `/admin/users`           | Admin; tối đa 200                        |
| PATCH  | `/admin/users/{id}/lock` | Admin; body `{"locked":true}` hoặc false |

Đăng ký:

```json
{
  "name": "Tên người dùng",
  "email": "player@example.com",
  "phone": "0901234567",
  "password": "<mật khẩu riêng từ 12 ký tự>",
  "role": "user"
}
```

Login: `{"email":"player@example.com","password":"<mật khẩu>"}`. Trả User và cookie session; không trả bearer/JWT. Đăng ký không tự đăng nhập. Email normalize lowercase; phone có thể bỏ qua khi đăng ký, nhưng cần số hợp lệ khi đặt sân/cập nhật hồ sơ. User response:

```json
{
  "id": "<id>",
  "name": "Tên người dùng",
  "email": "player@example.com",
  "phone": "0901234567",
  "area": "",
  "role": "user",
  "locked": false,
  "createdAt": "2026-11-06T08:00:00Z"
}
```

PATCH `/me`: `{"name":"Tên mới","phone":"0901234567","area":"Cầu Giấy, Hà Nội"}`. Không cho đổi role/email bằng profile. Admin không thể khóa admin qua endpoint `/lock`.

## Sân

| Method       | Path                                        | Quyền / query                                                             |
| ------------ | ------------------------------------------- | ------------------------------------------------------------------------- |
| GET          | `/venues`                                   | Public; `sport`, `city`, `q` tùy chọn; chỉ approved, tối đa 200 trước lọc |
| GET          | `/venues/{id}`                              | Public approved; owner của sân/admin có thể xem pending/rejected          |
| GET          | `/venues/{id}/availability?date=YYYY-MM-DD` | Public approved; không trả thông tin người đặt                            |
| GET          | `/owner/venues`                             | Owner; chỉ sân của mình                                                   |
| POST         | `/owner/venues`                             | Owner; sân mới pending                                                    |
| PATCH        | `/owner/venues/{id}/courts/{courtID}`       | Owner của sân                                                             |
| GET          | `/admin/venues?status=pending`              | Admin; status có thể pending/approved/rejected hoặc bỏ qua                |
| PATCH        | `/admin/venues/{id}/review`                 | Admin                                                                     |
| GET          | `/me/favorites`                             | Đăng nhập; trả sân approved đã lưu                                        |
| PUT / DELETE | `/me/favorites/{id}`                        | Đăng nhập; thêm/bỏ, idempotent, trả 204                                   |

Tạo sân (1–20 sân con; giá nguyên VND trong khoảng 1.000–10.000.000):

```json
{
  "name": "Sân bóng Thành Đồng",
  "sport": "Bóng đá",
  "city": "Hà Nội",
  "address": "Nam Trung Yên, Cầu Giấy, Hà Nội",
  "amenities": ["Đèn LED", "Bãi đỗ xe"],
  "courts": [{ "name": "Sân A", "dayPrice": 180000, "eveningPrice": 220000 }]
}
```

Venue response có `id`, `ownerId`, `name`, `sport`, `city`, `address`, `amenities`, `status`, `reviewComment`, `createdAt`, `courts`. Mỗi court có `id`, `name`, `dayPrice`, `eveningPrice`, `maintenance`.

Sửa sân con: `{"dayPrice":180000,"eveningPrice":240000,"maintenance":false}`; cả 3 trường bắt buộc. Không đổi giá của booking cũ. Chưa có endpoint đổi mô tả/cơ sở hoặc thêm sân con sau khi tạo.

Duyệt: `{"status":"approved","comment":"Đã kiểm tra hồ sơ"}`. Từ chối: `{"status":"rejected","comment":"Cần bổ sung hồ sơ sở hữu"}`. Lý do từ chối ít nhất 10 ký tự. FE có thể hiển thị rejected thành “Cần bổ sung”.

Availability trả array:

```json
[
  {
    "courtId": "<court-id>",
    "date": "2026-11-08",
    "hour": 19,
    "price": 220000,
    "available": true
  }
]
```

## Đặt lịch

| Method | Path                                          | Quyền                                                    |
| ------ | --------------------------------------------- | -------------------------------------------------------- |
| POST   | `/bookings`                                   | Đăng nhập; tạo lịch cho mình                             |
| GET    | `/me/bookings`                                | Đăng nhập; chỉ lịch của mình, không chứa block           |
| GET    | `/owner/bookings?venueId=...&date=YYYY-MM-DD` | Owner của sân hoặc admin; date có thể bỏ qua             |
| POST   | `/owner/blocks`                               | Owner của sân                                            |
| PATCH  | `/bookings/{id}/status`                       | User của lịch / owner của sân / admin theo state machine |
| GET    | `/owner/revenue?venueId=...&month=YYYY-MM`    | Owner của sân hoặc admin                                 |

Tạo lịch:

```json
{
  "courtId": "<court-id>",
  "date": "2026-11-08",
  "hour": 19,
  "name": "Tên người đặt",
  "phone": "0901234567"
}
```

Ngày ví dụ cần đổi thành ngày/giờ trong tương lai. Không truyền `userId`, `total`, `deposit` hay `status` — server tự xác định, trường thừa bị từ chối. Một lịch dài 60 phút. `hour` là giờ bắt đầu (16–21). Phone: 10 chữ số bắt đầu 0 hoặc `+84` + 9 chữ số, không có khoảng trắng.

Response:

```json
{
  "id": "<booking-id>",
  "userId": "<user-id>",
  "venueId": "<venue-id>",
  "courtId": "<court-id>",
  "date": "2026-11-08",
  "hour": 19,
  "name": "Tên người đặt",
  "phone": "0901234567",
  "total": 220000,
  "deposit": 66000,
  "status": "pending",
  "createdAt": "2026-11-06T08:00:00Z"
}
```

Block dùng cùng body nhưng `name` là lý do khóa; phone có thể bỏ qua, server luôn đặt total/deposit = 0, status = blocked.

PATCH status body: `{"status":"cancelled"}` (hoặc confirmed/completed).

```text
pending   → confirmed  (owner của sân / admin)
confirmed → completed  (owner của sân / admin)
pending / confirmed → cancelled  (người đặt / owner của sân / admin)
blocked → cancelled  (owner của sân / admin, mở khóa giờ)
```

Không có chuyển trạng thái ngược hoặc sửa lịch đã completed/cancelled. Xác nhận chỉ là thao tác vận hành; **không xác nhận tiền cọc thực**. Revenue trả `month`, `count`, `total`, `completed`, `pendingDeposit` — giá trị lịch, không phải ledger tài chính.

## Gọi thử từ trình duyệt

FE demo chưa gọi API. Khi tích hợp, phải gửi cookie:

```js
const response = await fetch("http://localhost:8080/api/v1/me", {
  credentials: "include",
});
if (!response.ok) throw new Error(`HTTP ${response.status}`);
const user = await response.json();
```

POST/PATCH thêm `headers: {"Content-Type":"application/json"}` và `body: JSON.stringify(data)`. Để dùng curl, dùng cookie jar riêng (`-c /path/cookies -b /path/cookies`), giữ file ngoài repository và không chia sẻ vì chứa session thật.
