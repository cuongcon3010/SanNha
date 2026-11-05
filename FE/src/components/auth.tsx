"use client";

import { OAuthButtons } from "./oauth";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  ArrowLeft,
  ArrowRight,
  Eye,
  EyeOff,
  UserRound,
  Building2,
  ShieldCheck,
} from "lucide-react";
import { Logo, DemoNotice, SuccessMessage } from "./ui";

export function Auth({
  mode,
  owner = false,
  oauthError = "",
}: {
  mode: "login" | "register" | "reset";
  owner?: boolean;
  oauthError?: string;
}) {
  const router = useRouter();
  const [role, setRole] = useState(owner ? "owner" : "user");
  const [visible, setVisible] = useState(false);
  const [error, setError] = useState("");
  const [done, setDone] = useState(false);
  const isRegister = mode === "register";
  const isReset = mode === "reset";
  function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");
    const f = new FormData(e.currentTarget);
    if (isRegister && String(f.get("name") || "").trim().length < 2) {
      setError(
        "Vui lòng nhập họ tên có ít nhất 2 ký tự, không chỉ khoảng trắng.",
      );
      return;
    }
    if (isRegister && f.get("password") !== f.get("confirm")) {
      setError("Mật khẩu xác nhận chưa trùng khớp.");
      return;
    }
    if (mode === "login") {
      const identifier = String(f.get("email") || "").trim();
      if (!/^(?:[^\s@]+@[^\s@]+\.[^\s@]+|0\d{9})$/.test(identifier)) {
        setError("Vui lòng nhập email hoặc số điện thoại 10 chữ số hợp lệ.");
        return;
      }
      router.push("/user");
      return;
    }
    setDone(true);
  }
  return (
    <main id="main-content" className="auth-layout">
      <aside className="auth-brand dotted-dark">
        <Logo light />
        <div className="auth-brand-copy">
          <span className="eyebrow">MỘT TÀI KHOẢN. NHIỀU CUỘC CHƠI.</span>
          <h2>
            {isRegister
              ? "Gia nhập đội hình."
              : isReset
                ? "Lấy lại quyền truy cập."
                : "Chào mừng"}
            <br />
            <span>
              {isRegister
                ? "Sẵn sàng ra sân."
                : isReset
                  ? "Để cuộc chơi tiếp nối."
                  : "trở lại sân nhà."}
            </span>
          </h2>
          <p>
            Tìm sân đúng gu, quản lý lịch dễ dàng
            <br />
            và không bỏ lỡ một kèo hay nào.
          </p>
        </div>
        <div className="auth-brand-bottom">
          <div className="auth-brand-card">
            <strong>
              01 / Tìm sân <span>02 / Chọn giờ</span> 03 / Ra sân
            </strong>
            <p>Cho người chơi. Đồng hành cùng chủ sân.</p>
          </div>
          <small>© 2026 Sân Nhà · Phiên bản demo</small>
        </div>
      </aside>
      <section className="auth-content">
        <Link href={isReset ? "/dang-nhap" : "/"} className="back-link">
          <ArrowLeft size={15} />
          {isReset ? "Quay lại đăng nhập" : "Về trang chủ"}
        </Link>
        <div className="auth-form-wrap">
          <span className="eyebrow auth-kicker">
            {isRegister
              ? "BẮT ĐẦU CUỘC CHƠI"
              : isReset
                ? "CHÚNG TÔI LUÔN Ở ĐÂY"
                : "SÂN NHÀ CHỜ BẠN"}
          </span>
          <h1>
            {isRegister
              ? role === "owner"
                ? "Đăng ký chủ sân"
                : "Tạo tài khoản"
              : isReset
                ? "Quên mật khẩu?"
                : "Đăng nhập"}
          </h1>
          <p className="auth-subtitle">
            {isRegister
              ? "Một bước nhỏ để bắt đầu cuộc chơi."
              : isReset
                ? "Nhập email đã dùng để tạo tài khoản."
                : "Đăng nhập để tiếp tục hành trình của bạn."}
          </p>
          {!isReset && <OAuthButtons errorCode={oauthError} />}
          <DemoNotice />
          {done ? (
            <div className="auth-done">
              <div className="success-icon">
                <ShieldCheck size={30} />
              </div>
              <h2>
                {isReset
                  ? "Đã hoàn tất bước minh họa"
                  : "Thông tin đăng ký hợp lệ"}
              </h2>
              <SuccessMessage>
                {isReset
                  ? "Nếu có tài khoản, hệ thống thật sẽ gửi email hướng dẫn. Bản demo không gửi email."
                  : role === "owner"
                    ? "Trong hệ thống thật, chủ sân cần xác minh hồ sơ trước khi đăng sân. Bản demo chưa tạo tài khoản."
                    : "Bản demo không tạo tài khoản hay lưu mật khẩu."}
              </SuccessMessage>
              <Link className="button full-width" href="/dang-nhap">
                Quay lại đăng nhập <ArrowRight size={16} />
              </Link>
            </div>
          ) : (
            <form onSubmit={submit} className="auth-form">
              {isRegister && (
                <fieldset className="role-fieldset">
                  <legend>Tôi muốn tham gia với vai trò</legend>
                  <div className="role-options">
                    <button
                      type="button"
                      className={role === "user" ? "active" : ""}
                      aria-pressed={role === "user"}
                      onClick={() => setRole("user")}
                    >
                      <UserRound size={17} />
                      Người chơi
                      <span className="radio-mark" />
                    </button>
                    <button
                      type="button"
                      className={role === "owner" ? "active" : ""}
                      aria-pressed={role === "owner"}
                      onClick={() => setRole("owner")}
                    >
                      <Building2 size={17} />
                      Chủ sân
                      <span className="radio-mark" />
                    </button>
                  </div>
                </fieldset>
              )}
              {isRegister && (
                <label className="field">
                  Họ và tên
                  <input
                    name="name"
                    required
                    minLength={2}
                    maxLength={80}
                    autoComplete="off"
                    placeholder="Nhập họ tên minh họa"
                  />
                </label>
              )}
              <div className={isRegister ? "form-columns" : ""}>
                <label className="field">
                  {mode === "login" ? "Email hoặc số điện thoại" : "Email"}
                  <input
                    name="email"
                    type={mode === "login" ? "text" : "email"}
                    required
                    autoComplete="off"
                    placeholder={
                      mode === "login"
                        ? "an.nguyen@example.com"
                        : "ban@example.com"
                    }
                    maxLength={150}
                  />
                </label>
                {isRegister && (
                  <label className="field">
                    Số điện thoại
                    <input
                      name="phone"
                      type="tel"
                      pattern="0[0-9]{9}"
                      required
                      placeholder="09xx xxx xxx"
                      autoComplete="off"
                      title="10 chữ số, bắt đầu bằng 0"
                    />
                  </label>
                )}
              </div>
              {!isReset && (
                <label className="field">
                  Mật khẩu minh họa
                  <span className="password-field">
                    <input
                      name="password"
                      type={visible ? "text" : "password"}
                      minLength={8}
                      maxLength={128}
                      required
                      autoComplete="off"
                      placeholder="Tối thiểu 8 ký tự · không dùng mật khẩu thật"
                    />
                    <button
                      type="button"
                      className="icon-button"
                      onClick={() => setVisible(!visible)}
                      aria-label={visible ? "Ẩn mật khẩu" : "Hiện mật khẩu"}
                    >
                      {visible ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                  </span>
                </label>
              )}
              {isRegister && (
                <label className="field">
                  Xác nhận mật khẩu
                  <input
                    name="confirm"
                    type={visible ? "text" : "password"}
                    minLength={8}
                    required
                    autoComplete="off"
                    placeholder="Nhập lại mật khẩu minh họa"
                  />
                </label>
              )}
              {mode === "login" && (
                <div className="form-between">
                  <label className="checkbox-label">
                    <input type="checkbox" /> Ghi nhớ (minh họa)
                  </label>
                  <Link href="/quen-mat-khau">Quên mật khẩu?</Link>
                </div>
              )}
              {isRegister && (
                <label className="checkbox-label terms-check">
                  <input type="checkbox" required />{" "}
                  <span>
                    Tôi đã đọc{" "}
                    <Link href="/huong-dan">
                      điều khoản & lưu ý bảo mật của bản demo
                    </Link>
                    .
                  </span>
                </label>
              )}
              {error && (
                <p className="form-error" role="alert">
                  {error}
                </p>
              )}
              <button type="submit" className="button full-width">
                {isRegister
                  ? role === "owner"
                    ? "Đăng ký tài khoản chủ sân"
                    : "Tạo tài khoản"
                  : isReset
                    ? "Gửi hướng dẫn (demo)"
                    : "Đăng nhập demo"}
                <ArrowRight size={17} />
              </button>
            </form>
          )}
          {!isReset && !done && (
            <p className="auth-switch">
              {isRegister ? "Đã có tài khoản?" : "Chưa có tài khoản?"}{" "}
              <Link href={isRegister ? "/dang-nhap" : "/dang-ky"}>
                {isRegister ? "Đăng nhập" : "Đăng ký ngay"} ↗
              </Link>
            </p>
          )}
          {mode === "login" && (
            <div className="demo-role-panel">
              <span className="eyebrow">KHÁM PHÁ BẢN MẪU THEO VAI TRÒ</span>
              <p>Các trang demo công khai, không cấp quyền tài khoản.</p>
              <div>
                <Link href="/user" className="button">
                  <UserRound size={16} />
                  Người chơi
                </Link>
                <Link href="/chu-san" className="button button-dark">
                  <Building2 size={16} />
                  Chủ sân
                </Link>
                <Link href="/admin" className="button button-neutral">
                  <ShieldCheck size={16} />
                  Admin
                </Link>
              </div>
            </div>
          )}
          {isRegister && (
            <p className="auth-footnote">
              <ShieldCheck size={15} />
              Chủ sân cần được duyệt trước khi đăng sân. Không có đăng ký admin.
            </p>
          )}
        </div>
      </section>
    </main>
  );
}
