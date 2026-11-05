# Sân Nhà — Next.js frontend

Voucher API thật: Admin tạo/tắt mã ở `/quan-tri/vouchers`; người chơi áp dụng mã ở `/dat-san`, xem giảm giá/cọc ở `/thanh-toan/[id]`. [Quy tắc voucher](../BE/VOUCHERS.md). Dashboard `/admin` cũ vẫn là demo.

Google/GitHub OAuth đã có nút theo cấu hình provider của BE và trang `/tai-khoan` kiểm tra session/logout thật. Xem [BE/OAUTH.md](../BE/OAUTH.md). Các form mật khẩu và dashboard/booking dưới đây vẫn dùng demo, chưa tích hợp API.

Ngoại lệ tích hợp thật: `/dat-san` lấy sân/khung giờ từ API và tạo booking; `/thanh-toan/[id]` tạo QR cọc VietQR/Tingo Pay và theo dõi callback; `/tai-khoan` có danh sách lịch backend. Xem [BE/VIETQR.md](../BE/VIETQR.md). Không tự chuyển lịch demo thành lịch thu tiền. Merchant chưa cấu hình thì payment bị vô hiệu hóa.

Ứng dụng hiện nằm trong `FE/`. Chạy các lệnh dưới đây từ thư mục này. Backend Go nằm tại [`../BE/`](../BE/README.md); dữ liệu backend và fixture/localStorage demo tách biệt.

Giao diện tiếng Việt từ thiết kế Penpot Sân Nhà: vàng `#F5D547`, đen `#111210`, trắng ngà `#F6F5EF`, đỏ `#E84938`, font Inter và nền chấm tròn lớn nhỏ. Responsive desktop/mobile.

## Chạy dự án

Yêu cầu **Node.js 24+** (bao gồm chạy test TypeScript bằng Node), npm 11+.

```bash
npm ci
npm run dev
```

Mở http://localhost:3000. Không cần biến môi trường hoặc khóa API.

```bash
npm run build
npm start
npm run lint
npm run typecheck
npm test
node --experimental-strip-types --test tests/booking.test.ts
npm run format
npm run format:check
```

`npm start` yêu cầu đã chạy `npm run build`. File `package-lock.json` cố định phiên bản dependency.

## Các trang

| Đường dẫn                                          | Nội dung                                                       |
| -------------------------------------------------- | -------------------------------------------------------------- |
| `/`                                                | Trang chủ, lọc môn/khu vực/tên sân, yêu thích                  |
| `/san/thanh-dong`, `/san/the-yard`, `/san/ha-dong` | Chi tiết, chọn sân con/ngày/giờ, đặt thử nghiệm                |
| `/xac-nhan?id=...`                                 | Xác nhận lịch mẫu vừa tạo                                      |
| `/dang-nhap`, `/dang-ky`, `/dang-ky?role=owner`    | Form đăng nhập/đăng ký demo                                    |
| `/quen-mat-khau`                                   | Form khôi phục mật khẩu demo                                   |
| `/user`, `/user/ho-so`, `/user/yeu-thich`          | Lịch đặt, hủy lịch mẫu, hồ sơ, yêu thích                       |
| `/chu-san`                                         | Timeline, xác nhận/hoàn thành/hủy lịch mẫu, khóa/mở giờ        |
| `/chu-san/san`                                     | Cập nhật giá sân con, trạng thái bảo trì                       |
| `/chu-san/doanh-thu`                               | Tổng giá trị lịch demo, lọc theo tháng                         |
| `/admin`, `/admin/nguoi-dung`, `/admin/kiem-duyet` | Tổng quan, lọc/khóa tài khoản mẫu, duyệt/yêu cầu bổ sung hồ sơ |
| `/huong-dan`                                       | Giới hạn, hướng dẫn và chính sách dữ liệu demo                 |

## Phạm vi demo, không phải hệ thống production

- **Các dashboard demo** `/user`, `/chu-san`, `/admin` công khai để duyệt thiết kế, chọn role demo **không cấp quyền**. Ngược lại, luồng OAuth, `/tai-khoan`, `/dat-san` và `/thanh-toan/[id]` dùng Go API/PostgreSQL/session thật; API kiểm tra quyền server-side. Không cho tự đăng ký admin.
- Không nhập mật khẩu hoặc thông tin cá nhân thật. Mật khẩu không lưu và không gửi. Đăng ký chỉ kiểm tra form, chưa tạo tài khoản; khôi phục không gửi email.
- Lịch mẫu, yêu thích, hồ sơ, bảng giá và thao tác admin lưu ở `localStorage` với khóa `san-nha-demo-v3`. Xóa dữ liệu website để reset. Dữ liệu dùng chung cho các vai trò demo trong **một tab trình duyệt**, chưa đồng bộ nhiều tab, thiết bị hoặc người dùng.
- Kiểm tra trùng lịch chỉ thực hiện trên dữ liệu trình duyệt, **không bảo đảm chống trùng trên hệ thống thực tế**. Cần API/transaction phía server khi tích hợp backend.
- Chủ sân thay đổi giá chỉ ảnh hưởng các lượt đặt sau đó. Lịch cũ giữ nguyên tổng tiền. Chưa có hoàn cọc hoặc giao dịch thật.
- Timeline và fixture mẫu bắt đầu ngày **06/11/2026** theo Penpot. Trang đặt chỉ cho tạo lịch từ ngày hiện tại; nếu quá ngày fixture, chọn ngày mới.
- Biểu đồ 7 ngày của admin và đánh giá sân là minh họa. Báo cáo chủ sân tính từ lịch demo. Ảnh thể thao minh họa không đại diện cơ sở thực tế.
- Trước khi triển khai thật: nghiệm thu merchant/callback VietQR, kiểm tra triển khai HTTPS/session, chính sách đổi/hủy/hoàn cọc và bảo vệ thông tin cá nhân. Backend đã có kiểm tra quyền/đặt trùng, nhưng các dashboard demo chưa chuyển sang API.

## Cấu trúc

```text
src/app/                   # Next.js App Router, metadata, global CSS
src/components/            # Trang chủ, auth, booking và thành phần chung
src/components/dashboard/  # Shell + nội dung user / owner / admin
src/lib/                   # Fixtures, quy tắc đặt sân, demo store, menu
tests/                     # Unit tests tìm sân, giá/cọc, validation, trùng lịch
public/assets/             # Font Inter và ảnh minh họa lưu cục bộ
```

Stack: Next.js App Router, React, TypeScript, CSS thuần, Lucide icons; ESLint và Prettier. Không dùng UI framework hoặc chart library. Font/ảnh cục bộ nên build không cần kết nối dịch vụ font/ảnh.

## Tài nguyên

Inter: Rasmus Andersson, SIL Open Font License 1.1, xem `public/assets/Inter-LICENSE.txt`. Ảnh minh họa từ Unsplash (URL nguồn trong `public/assets/SOURCES.md`). Lucide: ISC license.

Kiểm tra dependency: `npm audit --omit=dev` không có cảnh báo tại thời điểm dựng. `npm audit` còn 8 cảnh báo mức moderate trong chuỗi công cụ lint (`minimatch`/`typescript-eslint`); không dùng `audit fix --force` vì đề xuất hạ Next ESLint config khác major. Cần theo dõi bản vá tooling khi cập nhật dependency.
