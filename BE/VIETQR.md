# VietQR / Tingo Pay — QR động và tự đối soát cọc

Nguồn hợp đồng API: https://api.vietqr.vn/vi/api-vietqr-callback/ (đọc ngày triển khai). Đây là **VietQR/Tingo Pay callback API**, không phải API `api.vietqr.io` hoặc dịch vụ chỉ sinh ảnh QR.

## Phạm vi

- Một tài khoản merchant/ngân hàng **chung của Sân Nhà**, không tự chia/chuyển tiền về từng chủ sân.
- Thu **30% tiền sân** từ booking đã lưu trên backend. Server xác định số tiền; không nhận số tiền từ FE hoặc localStorage.
- QR động `qrType=0`, `transType=C`, `orderId` dài 13 ký tự, nội dung không dấu/không ký tự đặc biệt. Retry giữ nguyên orderId, không tạo đơn thanh toán mới.
- VietQR cấp token cho BE gọi QR API; BE cấp **token callback riêng**, thời hạn 300 giây, cho VietQR gửi Transaction Sync. Token callback ngẫu nhiên, chỉ lưu hash trong PostgreSQL; không yêu cầu token phải là JWT theo hợp đồng.
- Chỉ callback được xác thực, đúng orderId, tiền cọc, tài khoản, chiều ghi có và thời gian hợp lệ mới đặt `depositPaid=true` và chuyển booking `pending → confirmed`. Hai thay đổi cùng một PostgreSQL transaction.
- Khi bật VietQR, API xác nhận thủ công không được thay thế bằng chứng đã nhận cọc. Không có nút “Tôi đã chuyển tiền” để tự ghi nhận thanh toán.
- Callback trùng cùng transactionid/nội dung được trả thành công, không thu/ghi sổ hai lần. Transactionid bị dùng với nội dung khác hoặc reference ngân hàng đã dùng bị từ chối.
- Sai tiền, sai tài khoản, chiều ghi nợ, khác môi trường, quá hạn hoặc callback đến sau khi hủy/đổi trạng thái: lưu event `review`, không xác nhận sân. Admin xem danh sách đối soát qua API. Không cộng dồn nhiều giao dịch thiếu tiền hoặc chấp nhận thừa tiền tự động.
- Lịch đã nhận cọc bị hủy vẫn giữ `depositPaid=true`; event xuất hiện trong đối soát với lý do `refund_required`. **Không hoàn tiền tự động.**

Đối soát nằm trong module `booking` (`payment*.go`, `vietqr.go`) vì thanh toán cọc và trạng thái lịch phải commit cùng nhau. Không truy vấn bảng `identity` hoặc `venue` từ phần thanh toán. Migration riêng `booking_payment_pg_v1`, không sửa schema cũ đã áp dụng.

## Đăng ký merchant trước khi thử thật

1. Mở tài khoản VietQR, liên kết tài khoản ngân hàng nhận cọc, khai báo merchant connection tại https://vietqr.vn/merchant/request.
2. Nhận **username/password sandbox** từ VietQR; không sử dụng credentials mẫu trong tài liệu công khai.
3. Tạo callback username/password của riêng Sân Nhà, đăng ký các callback URL HTTPS công khai với VietQR. Không chia sẻ secret trong chat hoặc commit `.env`.
4. Test Get Token, Generate QR và Test Callback theo hướng dẫn VietQR; hoàn tất nghiệm thu rồi mới đổi sang production và credentials production.

## Environment BE

```dotenv
VIETQR_ENABLED=false
VIETQR_ENV=sandbox
VIETQR_USERNAME=
VIETQR_PASSWORD=
VIETQR_CALLBACK_USERNAME=
VIETQR_CALLBACK_PASSWORD=
VIETQR_BANK_CODE=
VIETQR_BANK_ACCOUNT=
VIETQR_BANK_NAME=
```

Chỉ bật `VIETQR_ENABLED=true` sau khi điền đủ thông tin. Callback password riêng, ít nhất 32 byte ngẫu nhiên. `VIETQR_BANK_CODE` là mã dạng `MB`/`BIDV` do VietQR hỗ trợ; `VIETQR_BANK_ACCOUNT` là tài khoản merchant đã liên kết; `VIETQR_BANK_NAME` là tên chủ tài khoản IN HOA, không dấu. Không có secret hoặc tài khoản nhận mặc định.

BE gọi:

| Môi trường | Base URL                         |
| ---------- | -------------------------------- |
| sandbox    | `https://dev.vietqr.org/vqr/api` |
| production | `https://api.vietqr.org/vqr/api` |

Endpoints: `POST /token_generate` dùng Basic Auth; `POST /qr/generate-customer` dùng Bearer token. Outbound HTTP timeout 10 giây, không theo redirect để tránh lộ credentials. Token cache theo process; credential không gửi FE. QR trả về được kiểm tra receiver/amount/order rồi render PNG cục bộ, không dùng dịch vụ ảnh bên thứ ba.

`sign` trong callback là optional và tài liệu công khai không quy định thuật toán/canonical string thống nhất; bản này xác thực bằng token callback được cấp sau Basic Auth, **không giả định một chữ ký HMAC tự chế**. Nếu hợp đồng merchant yêu cầu ký `sign`, cần bổ sung đúng specification được VietQR cấp trước GoLive.

## Đăng ký callback HTTPS

Giả sử public API origin `https://api.example.com`:

- Token: `POST https://api.example.com/api/v1/payments/vietqr/api/token_generate`
- Transaction Sync: `POST https://api.example.com/api/v1/payments/vietqr/bank/api/transaction-sync`

Token endpoint dùng Basic Auth của `VIETQR_CALLBACK_USERNAME/PASSWORD`, trả `access_token`, `token_type=Bearer`, `expires_in=300`. Token này chỉ cấp quyền gọi callback, không thể dùng làm session user/admin.

Transaction Sync yêu cầu `Authorization: Bearer <callback-token>`, `Content-Type: application/json`:

```json
{
  "bankaccount": "<merchant-account>",
  "amount": 66000,
  "transType": "C",
  "content": "SNABCDEFGHIJK",
  "transactionid": "<provider-transaction-id>",
  "transactiontime": 1794000000000,
  "referencenumber": "<bank-reference>",
  "orderId": "SNABCDEFGHIJK"
}
```

Các giá trị trên chỉ minh họa, không dùng để xác nhận giao dịch thật. `transactiontime` là timestamp milliseconds. Giao dịch hợp lệ không được trước lúc tạo order quá 60 giây hoặc ở tương lai quá 5 phút. Việc ghép đơn dựa vào `orderId` VietQR trả, không tìm chuỗi gần đúng trong nội dung chuyển khoản.

Response đúng hợp đồng VietQR: `error`, `errorReason`, `toastMessage`, `object.reftransactionid`. ACK chỉ sau khi event được commit; lỗi database trả 500 để VietQR retry. Callback sai số tiền nhưng thuộc order đã biết được ACK sau khi lưu đối soát, không đánh dấu đã trả tiền. Order không biết hoặc transactionid bị tái sử dụng trả lỗi.

Các callback là server-to-server, không gửi session cookie của người chơi, không dùng CORS từ browser. Public reverse proxy phải cho qua Authorization và POST body, bật HTTPS, không ghi token/credentials/body có thông tin ngân hàng vào access log; bổ sung rate limit tại edge, bảo vệ Basic token endpoint. Không mở callback nếu merchant chưa cấu hình.

## API cho FE

| Method | Path (base `/api/v1`)    | Quyền                                                      |
| ------ | ------------------------ | ---------------------------------------------------------- |
| GET    | `/payments/config`       | Public, chỉ bật/tắt và môi trường                          |
| POST   | `/bookings/{id}/payment` | Người đặt sở hữu booking, body `{}`; tạo/lấy QR cùng order |
| GET    | `/bookings/{id}/payment` | Người đặt sở hữu booking; chỉ đọc, không tạo QR            |
| GET    | `/admin/payments/review` | Admin, tối đa 200 event cần đối soát/hoàn cọc              |

Payment response có `bookingId`, `orderId`, `amount`, `bankCode`, `bankAccount`, `bankName`, `content`, `status` (`pending/paid/review`), `bookingStatus`, `environment`, `expiresAt` (Unix seconds), `expired` do server tính và `qrImage` PNG data URL khi còn chờ hợp lệ. Không trả access token, callback token hoặc secret.

FE:

- `/tai-khoan`: session thật Google/GitHub và danh sách lịch backend.
- `/dat-san`: chọn cơ sở/sân con/ngày/giờ từ API, tạo booking thật; chỉ cho tiếp tục khi VietQR đã bật. DB mới phải có sân được admin duyệt trước.
- `/thanh-toan/[id]`: người đặt bấm tạo QR; kiểm tra trạng thái qua GET mỗi 5 giây tới khi trả tiền/cần đối soát/quá hạn. Nhãn sandbox không được thay bằng “đã thanh toán”. Không thu tiền cho mã lịch demo.
- `/san/*`, `/user`, `/chu-san`, `/admin` cũ vẫn là giao diện demo; không chuyển localStorage booking sang payment thật. Có link từ chi tiết sân demo đến `/dat-san`.

## Giới hạn phải biết trước vận hành

- **QR hết hạn cục bộ sau 15 phút**: FE ngừng hiển thị và callback giao dịch sau hạn vào `review`. Đây không phải lệnh hủy QR phía ngân hàng/VietQR; QR đã lưu/chụp có thể vẫn chuyển tiền được. Không chuyển khoản sau hạn.
- **Chưa có tự hết hạn giữ sân**: booking pending vẫn giữ slot đến khi user/owner/admin hủy. Chưa tự giải phóng sau 15 phút vì cần chính sách xử lý thanh toán trễ/hoàn cọc. Không tạo lịch hàng loạt để thử tải trên production.
- Giao dịch đến sau khi lịch hủy/đã bắt đầu/đã xác nhận bằng phương thức khác không phục hồi hay đặt trùng sân; cần xử lý đối soát thủ công. Event được giữ để không mất dấu tiền nhận.
- Chưa hoàn tiền, chia tiền chủ sân, quyết toán, thu phần còn lại, retry job đối soát ngân hàng hoặc dashboard xử lý refund. Endpoint review chỉ đọc, chưa đánh dấu đã giải quyết.
- Không đổi bank/environment của order cũ: snapshot tài khoản và môi trường được giữ nguyên, tạo QR lại bị chặn nếu cấu hình merchant thay đổi. Sandbox và production phải dùng credentials/callback credentials riêng và tốt nhất database/deployment riêng.
- Chưa kiểm thử gateway thật hoặc nghiệm thu merchant khi chưa có credentials và public callback. Unit/integration tests dùng mock HTTP và PostgreSQL riêng; không thực hiện chuyển tiền thật.

## Kiểm tra

```bash
# Trong BE/, với TEST_DATABASE_URL riêng
go test -race ./...
go test -race ./internal/modules/booking -run 'TestPayment|TestCallback|TestVietQR' -v
go vet ./...
```

Test bao gồm provider auth/response mismatch, giá cọc, authorization, cùng order khi retry, callback lặp đồng thời, token giả/hết hạn, sửa nội dung transactionid, sai receiver/tiền/chiều, callback quá hạn và lịch hủy. Chạy sandbox Test Callback của VietQR là bước nghiệm thu bổ sung, không thay thế bằng gọi callback giả trên production.
