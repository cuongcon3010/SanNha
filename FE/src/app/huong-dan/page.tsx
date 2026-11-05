import Link from "next/link";
import { Header } from "@/components/header";
import { DemoNotice, Footer } from "@/components/ui";

export default function GuidePage() {
  return (
    <>
      <Header />
      <main id="main-content" className="container section guide">
        <span className="eyebrow">SÂN NHÀ / BẢN THỬ NGHIỆM</span>
        <h1>Trước khi bắt đầu cuộc chơi.</h1>
        <DemoNotice />
        <section className="panel">
          <h2>Đặt sân trong bản demo</h2>
          <p>
            Chọn sân, ngày, sân con và một khung giờ. Dùng tên và số điện thoại
            giả hợp lệ để tạo lịch minh họa. Lịch được lưu trong trình duyệt
            này, không gửi đến chủ sân.
          </p>
          <h2>Tài khoản & quyền truy cập demo</h2>
          <p>
            Form mật khẩu đăng nhập, đăng ký và khôi phục mật khẩu chỉ minh họa
            giao diện. Không tạo tài khoản, không gửi email. Các trang người
            chơi, chủ sân và admin đều công khai để duyệt thiết kế; đây không
            phải cơ chế phân quyền.
          </p>
          <h2>Thanh toán, hủy lịch & dữ liệu demo</h2>
          <p>
            Không thu tiền hay tạo giao dịch. Hủy lịch chỉ cập nhật dữ liệu
            demo. Không có chính sách hoàn cọc thực tế. Không nhập mật khẩu hay
            thông tin cá nhân thật. Xóa dữ liệu website trong trình duyệt để đặt
            lại lịch và hồ sơ mẫu.
          </p>
          <h2>Đặt sân và thanh toán qua API thật</h2>
          <p>
            Đăng nhập Google/GitHub dùng session backend. Trang /tai-khoan hiển
            thị lịch backend; /dat-san tạo lịch thật và /thanh-toan/[id] nhận QR
            cọc VietQR/Tingo Pay. Các luồng này tách biệt với demo: không thanh
            toán cho lịch localStorage. Kiểm tra rõ nhãn SANDBOX hay PRODUCTION
            trước khi thao tác.
          </p>
          <p>
            Sandbox chỉ dùng giao dịch giả lập, không chuyển tiền thật. Khi
            production được cấu hình, tiền cọc vào tài khoản merchant chung Sân
            Nhà. Chỉ callback được xác thực/đối soát mới xác nhận cọc; không có
            nút tự đánh dấu đã chuyển tiền. QR quá hạn hiển thị sau 15 phút
            không phải lệnh hủy chuyển khoản ngân hàng. Không chuyển tiền sau
            hạn.
          </p>
          <p>
            Chưa có hoàn cọc tự động hoặc tự giải phóng lịch chờ. Nếu giao dịch
            cần đối soát, bị sai tiền hoặc lịch đã hủy, không chuyển thêm tiền;
            liên hệ đơn vị vận hành Sân Nhà. Tích hợp chưa nghiệm thu merchant
            thật, không dùng production trước khi hoàn tất cấu hình và nghiệm
            thu.
          </p>
          <p>
            Ảnh từ Unsplash chỉ dùng để minh họa thể thao, không đại diện cho cơ
            sở thực tế.
          </p>
          <Link href="/" className="button">
            Bắt đầu khám phá ↗
          </Link>
        </section>
      </main>
      <Footer />
    </>
  );
}
