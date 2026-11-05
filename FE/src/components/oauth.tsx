"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { LiveBookings } from "./live-booking";
import { apiOrigin } from "@/lib/api";

export function OAuthButtons({ errorCode = "" }: { errorCode?: string }) {
  const [providers, setProviders] = useState<string[]>([]);
  const errors: Record<string, string> = {
    invalid_state:
      "Phiên đăng nhập đã hết hạn hoặc không hợp lệ. Vui lòng thử lại.",
    cancelled: "Bạn đã hủy đăng nhập.",
    provider_failed: "Không kết nối được nền tảng đăng nhập.",
    email_unverified: "Tài khoản cần có email chính đã xác minh.",
    account_exists:
      "Email đã được đăng ký bằng phương thức khác. Vui lòng dùng phương thức ban đầu; chưa hỗ trợ ghép tài khoản.",
    login_failed: "Không thể đăng nhập. Tài khoản có thể đã bị khóa.",
  };
  const error = errorCode
    ? errors[errorCode] || "Không thể đăng nhập OAuth."
    : "";
  useEffect(() => {
    const controller = new AbortController();
    fetch(`${apiOrigin}/api/v1/auth/oauth/providers`, {
      signal: controller.signal,
    })
      .then((r) => {
        if (!r.ok) throw new Error();
        return r.json();
      })
      .then((data) => {
        if (Array.isArray(data))
          setProviders(data.filter((p) => p === "google" || p === "github"));
      })
      .catch(() => {});
    return () => controller.abort();
  }, []);
  return (
    <>
      {error && (
        <p role="alert" className="form-error">
          {error}
        </p>
      )}
      {providers.length > 0 && (
        <section className="demo-role-panel">
          <span className="eyebrow">ĐĂNG NHẬP THẬT QUA TÀI KHOẢN CỦA BẠN</span>
          <p>
            Google/GitHub dùng phiên backend, tách biệt với bản demo bên dưới.
          </p>
          <div>
            {providers.map((p) => (
              <a
                key={p}
                className="button button-white"
                href={`${apiOrigin}/api/v1/auth/oauth/${p}/start`}
              >
                {p === "google" ? "Google" : "GitHub"} ↗
              </a>
            ))}
          </div>
        </section>
      )}
    </>
  );
}

export function OAuthAccount() {
  const router = useRouter();
  const [user, setUser] = useState<{
    name: string;
    email: string;
    role: string;
  } | null>(null);
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    fetch(`${apiOrigin}/api/v1/me`, {
      credentials: "include",
      signal: controller.signal,
    })
      .then(async (r) => {
        if (!r.ok)
          throw new Error(
            "Phiên đăng nhập không còn hợp lệ. Vui lòng đăng nhập lại.",
          );
        return r.json();
      })
      .then(setUser)
      .catch((e) => {
        if (e.name !== "AbortError") setError(e.message);
      });
    return () => controller.abort();
  }, []);
  async function logout() {
    setError("");
    try {
      const r = await fetch(`${apiOrigin}/api/v1/auth/logout`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: "{}",
      });
      if (!r.ok) throw new Error("Không thể đăng xuất. Vui lòng thử lại.");
      router.replace("/dang-nhap");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không thể đăng xuất.");
    }
  }
  return (
    <section className="panel">
      <h1>Tài khoản đã xác thực</h1>
      {error && (
        <p role="alert" className="form-error">
          {error}
        </p>
      )}
      {user ? (
        <>
          <p className="success-message">
            Xin chào {user.name}. Bạn đã đăng nhập vào backend.
          </p>
          <p>{user.email}</p>
          <p>Vai trò: {user.role}</p>
          {user.role === "admin" && (
            <p>
              <a className="button" href="/quan-tri/vouchers">
                Quản lý voucher thật ↗
              </a>
            </p>
          )}
          <p className="muted">
            Dashboard và trang chủ dùng demo. Đặt sân/thanh toán qua API ở danh
            sách bên dưới.
          </p>
          <button className="button" onClick={logout}>
            Đăng xuất tài khoản thật
          </button>
          <LiveBookings />
        </>
      ) : (
        !error && <p role="status">Đang kiểm tra phiên…</p>
      )}
      <p>
        <a className="back-link" href="/dang-nhap">
          ← Về trang đăng nhập
        </a>
      </p>
    </section>
  );
}
