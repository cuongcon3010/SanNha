# Voucher Sân Nhà

Voucher giảm tiền sân trong **luồng API thật** `/dat-san` → `/thanh-toan/[id]`. Quản trị thật tại `/quan-tri/vouchers`. Các trang `/user`, `/admin` và sân fixture cũ vẫn dùng demo; không áp dụng voucher thật vào localStorage hoặc mã lịch demo.

## Quy tắc bản đầu

- Admin tạo voucher toàn nền tảng (không giới hạn riêng chủ sân/cơ sở), loại `fixed` (số tiền VND) hoặc `percent` (1–99%). Không cộng dồn nhiều voucher trong một booking.
- Mã gồm **3–24 chữ cái Latin/chữ số**, không dấu, không khoảng trắng ở giữa; server trim và chuyển IN HOA. Mã duy nhất, không tái sử dụng tên mã đã tắt/hết hạn.
- Có ngày bắt đầu/kết thúc, mức giá sân tối thiểu, số tiền giảm tối đa và tổng số lượt. `maxDiscount=0` là không có trần riêng, nhưng tổng sau giảm luôn **tối thiểu 1.000đ** để VietQR có tiền cọc dương. Không hỗ trợ booking miễn phí.
- Giảm phần trăm làm tròn xuống theo VND nguyên. Cọc = 30% **tổng sau giảm**, làm tròn nguyên theo quy tắc backend hiện có. Không nhận `discount`, `total` hoặc `deposit` từ FE.
- Mỗi tài khoản được giữ/dùng **một lượt/mã**. Lượt được giữ ngay khi booking tạo thành công, không chờ tới lúc thanh toán; tránh cấp vượt ngân sách voucher khi nhiều người đặt cùng lúc.
- Báo giá chỉ kiểm tra, **không giữ lượt/giờ/giá**. Server kiểm tra lại mã, hạn, lượt và giá lúc tạo booking. Nếu mã vừa hết lượt/hết hạn/được tắt, booking bị từ chối; không âm thầm bỏ mã rồi tính giá cao hơn.
- Hủy lịch khi `depositPaid=false`: trả lượt và cho tài khoản dùng lại. Hủy lịch đã nhận cọc không tự trả lượt; không có hoàn tiền hoặc hoàn voucher thủ công qua API trong bản đầu. Callback đến sau khi đã hủy vẫn vào đối soát, không phục hồi lịch/voucher cũ.
- Giá gốc, giảm giá, mã, tổng và tiền cọc là **snapshot** trong booking. Sửa giá sân hoặc tắt voucher không đổi giá/QR của lịch đã tạo. Không thêm/đổi voucher sau khi đặt hoặc tạo QR.
- Admin chỉ tạo, xem, bật/tắt voucher; chưa sửa số tiền, thời hạn hoặc tổng lượt của mã đã tạo để tránh thay đổi điều kiện của những lịch đã giữ lượt. Bật lại mã không reset lịch sử dùng.

## Modulith & transaction

Tất cả nằm trong module `booking`: `booking_vouchers`, `booking_voucher_uses`, và snapshot `subtotal/discount/voucher_code` của booking. Không đọc trực tiếp bảng module khác.

Migration forward-only `booking_voucher_pg_v1` backfill `subtotal=total` cho lịch có sẵn, giữ nguyên tiền sân/cọc. Tạo booking, giữ lượt người dùng và tăng `used_count` commit trong cùng transaction; lỗi trùng sân hoặc lỗi ghi DB rollback tất cả. Unique index bảo vệ một lượt/tài khoản, khóa hàng voucher bảo vệ tổng quota giữa các process.

Khóa voucher **trước** hàng booking khi hủy lịch để khớp thứ tự tạo lịch và tránh deadlock (hai request chờ khóa lẫn nhau). Callback cọc cùng dùng khóa booking nên không thể vừa ghi nhận cọc vừa trả lượt sai thứ tự.

**Nâng cấp:** dừng API phiên bản cũ trước khi chạy binary mới/migration này. Binary cũ không ghi các cột snapshot mới và không biết quy tắc voucher; không chạy đồng thời binary cũ/mới trong đợt nâng cấp schema này. Sao lưu DB trước nâng cấp.

## API (base `/api/v1`)

| Method | Path                     | Quyền                                          |
| ------ | ------------------------ | ---------------------------------------------- |
| POST   | `/bookings/quote`        | Đã đăng nhập; báo giá theo sân/ngày/giờ/mã     |
| POST   | `/admin/vouchers`        | Admin; tạo mã                                  |
| GET    | `/admin/vouchers`        | Admin; tối đa 200 mã, có lượt đang giữ/đã dùng |
| PATCH  | `/admin/vouchers/{code}` | Admin; chỉ `{"active":false}` hoặc true        |

Tạo voucher:

```json
{
  "code": "SANNHA50",
  "kind": "fixed",
  "value": 50000,
  "minTotal": 100000,
  "maxDiscount": 0,
  "startsAt": 1793872800,
  "endsAt": 1796464800,
  "maxUses": 100
}
```

Thời gian là Unix **seconds**; ví dụ cần thay bằng thời gian chiến dịch thực tế. Giờ nhập trong form admin là giờ thiết bị, được đổi sang Unix seconds. Tổng lượt 1–1.000.000, giá trị tiền tối đa 10.000.000đ. Với `kind=percent`, `value` là phần trăm, `maxDiscount` là VND.

Báo giá:

```json
{
  "courtId": "<court-id>",
  "date": "2030-01-01",
  "hour": 19,
  "voucherCode": "sannha50"
}
```

Response ví dụ cho sân 220.000đ:

```json
{
  "subtotal": 220000,
  "discount": 50000,
  "total": 170000,
  "deposit": 51000,
  "voucherCode": "SANNHA50"
}
```

`voucherCode` có thể bỏ qua/rỗng để báo giá thường. Tạo booking qua `POST /bookings` thêm trường `voucherCode` vào body hiện có; không gửi lại số tiền báo giá. Response booking, payment và danh sách lịch có snapshot voucher. Không áp dụng mã cho `POST /owner/blocks`.

Mã không hợp lệ/không tồn tại/không đủ giá tối thiểu trả 400; mã hết hạn, chưa bắt đầu, tắt, hết lượt hoặc tài khoản đã dùng trả 409. Thiếu đăng nhập trả 401; không phải admin cho endpoint quản trị trả 403.

## Giao diện

- `/dat-san`: chọn sân/ngày/giờ, nhập mã và bấm **Áp dụng mã**. Đổi điều kiện hoặc sửa mã phải áp dụng lại. Báo giá cũ đang tải về không được dùng cho lựa chọn mới. Chỉ tạo booking sau khi mã đã được kiểm tra, và BE vẫn tính lại độc lập.
- `/thanh-toan/[id]`: hiển thị giá gốc, mã, tiền giảm, tổng sau giảm và cọc đúng snapshot; không sửa mã ở bước chuyển tiền.
- `/tai-khoan`: lịch backend hiển thị voucher đã dùng; admin có link đến quản lý mã.
- `/quan-tri/vouchers`: yêu cầu session admin thật. Có form đăng nhập mật khẩu gọi BE, không phải demo và không lưu mật khẩu vào localStorage. Admin phải được tạo bằng CLI `BE/cmd/admin` theo README, không có đăng ký admin công khai.

## Kiểm tra

Trong `BE/`, đặt `TEST_DATABASE_URL` vào DB test riêng:

```bash
go test -race ./...
go test -race ./internal/modules/booking -run TestVoucher -v
```

Tests kiểm tra tiền nguyên, trần giảm, giá tối thiểu còn phải trả, thời hạn/điều kiện, quyền admin, normalize mã, quote không giữ lượt, rollback khi trùng sân, tranh lượt cuối, tranh cùng tài khoản, trả lượt khi hủy chưa trả tiền, không trả lượt đã nhận cọc, cọc QR sau giảm và dữ liệu cũ sau migration.

## Chưa làm

Chưa phân voucher riêng từng chủ sân/bộ môn, tự phát mã, phân khúc khách hàng, nhiều mã một đơn, free booking hoặc tự trả lượt khi hold hết hạn. Lịch `pending` hiện không tự giải phóng: voucher cũng tiếp tục giữ lượt đến khi lịch hủy; người vận hành cần theo dõi/hủy lịch bỏ dở. Cần quy trình chống lạm dụng trước khi phát chiến dịch lớn. VietQR thật vẫn cần cấu hình/merchant nghiệm thu, không tự sinh mã mẫu hoặc thực hiện chuyển tiền khi test.
