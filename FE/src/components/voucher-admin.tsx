"use client";

import Link from "next/link";
import { useEffect, useState, type FormEvent } from "react";
import { api, type Voucher } from "@/lib/api";
import { money } from "@/lib/data";

export function VoucherAdmin() {
  const [authorized, setAuthorized] = useState(false);
  const [loading, setLoading] = useState(true);
  const [items, setItems] = useState<Voucher[]>([]);
  const [busy, setBusy] = useState(false);
  const [kind, setKind] = useState("fixed");
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    api<Voucher[]>("/admin/vouchers", { signal: controller.signal })
      .then((data) => {
        setItems(data);
        setAuthorized(true);
        setLoading(false);
      })
      .catch((e) => {
        if (e.name !== "AbortError") {
          setError(e.message);
          setLoading(false);
        }
      });
    return () => controller.abort();
  }, []);
  async function login(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    const form = e.currentTarget;
    const data = new FormData(form);
    try {
      const user = await api<{ role: string }>("/auth/login", {
        method: "POST",
        body: JSON.stringify({
          email: String(data.get("email")),
          password: String(data.get("password")),
        }),
      });
      form.reset();
      if (user.role !== "admin")
        throw new Error(
          "Tài khoản này không phải admin. Đăng nhập không cấp thêm quyền.",
        );
      const rows = await api<Voucher[]>("/admin/vouchers");
      setItems(rows);
      setAuthorized(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không đăng nhập được.");
    } finally {
      setBusy(false);
    }
  }
  async function create(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    setNotice("");
    const form = e.currentTarget;
    const f = new FormData(form);
    const body = {
      code: String(f.get("code")).trim().toUpperCase(),
      kind,
      value: Number(f.get("value")),
      minTotal: Number(f.get("minTotal")),
      maxDiscount: Number(f.get("maxDiscount")),
      startsAt: Math.floor(
        new Date(String(f.get("startsAt"))).getTime() / 1000,
      ),
      endsAt: Math.floor(new Date(String(f.get("endsAt"))).getTime() / 1000),
      maxUses: Number(f.get("maxUses")),
    };
    try {
      await api("/admin/vouchers", {
        method: "POST",
        body: JSON.stringify(body),
      });
      const rows = await api<Voucher[]>("/admin/vouchers");
      setItems(rows);
      form.reset();
      setKind("fixed");
      setNotice(`Đã tạo voucher ${body.code}.`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không tạo được voucher.");
    } finally {
      setBusy(false);
    }
  }
  async function toggle(v: Voucher) {
    if (
      !window.confirm(
        `${v.active ? "Tắt" : "Bật lại"} voucher ${v.code}? Các lịch đã đặt giữ nguyên giá.`,
      )
    )
      return;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const updated = await api<Voucher>(
        `/admin/vouchers/${encodeURIComponent(v.code)}`,
        { method: "PATCH", body: JSON.stringify({ active: !v.active }) },
      );
      setItems((rows) =>
        rows.map((row) => (row.code === v.code ? updated : row)),
      );
      setNotice(`Đã ${updated.active ? "bật" : "tắt"} ${v.code}.`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không cập nhật được voucher.");
    } finally {
      setBusy(false);
    }
  }
  return (
    <>
      <section className="panel">
        <span className="eyebrow">QUẢN TRỊ VOUCHER · API THẬT</span>
        <h1>Mã ưu đãi Sân Nhà</h1>
        <p className="muted">
          Voucher toàn nền tảng. Mỗi tài khoản một lượt/mã; không cộng dồn. Giá
          sau giảm tối thiểu 1.000đ để có thể thanh toán cọc VietQR.
        </p>
        {error && (
          <p role="alert" className="form-error">
            {error}
          </p>
        )}
        {notice && (
          <p role="status" className="success-message">
            {notice}
          </p>
        )}
        {loading ? (
          <p role="status">Đang kiểm tra quyền admin…</p>
        ) : !authorized ? (
          <form className="stack-form" onSubmit={login}>
            <h2>Đăng nhập admin thật</h2>
            <p className="muted">
              Dùng tài khoản admin được tạo bằng CLI backend. Đây không phải
              form demo; mật khẩu được gửi đến API để xác thực, không lưu ở
              trình duyệt.
            </p>
            <label className="field">
              Email admin
              <input
                name="email"
                type="email"
                required
                autoComplete="username"
              />
            </label>
            <label className="field">
              Mật khẩu
              <input
                name="password"
                type="password"
                required
                autoComplete="current-password"
              />
            </label>
            <button className="button" disabled={busy}>
              {busy ? "Đang đăng nhập…" : "Đăng nhập quản trị"}
            </button>
          </form>
        ) : (
          <form className="stack-form" onSubmit={create}>
            <h2>Tạo voucher</h2>
            <div className="form-columns">
              <label className="field">
                Mã (3–24 chữ/số, không dấu)
                <input
                  name="code"
                  required
                  pattern="[a-zA-Z0-9]{3,24}"
                  maxLength={24}
                  placeholder="SANNHA50"
                  autoComplete="off"
                />
              </label>
              <label className="field">
                Loại giảm
                <select value={kind} onChange={(e) => setKind(e.target.value)}>
                  <option value="fixed">Số tiền cố định (đ)</option>
                  <option value="percent">Phần trăm (%)</option>
                </select>
              </label>
            </div>
            <div className="form-columns">
              <label className="field">
                {kind === "percent" ? "Phần trăm (1–99)" : "Số tiền giảm (đ)"}
                <input
                  key={kind}
                  name="value"
                  type="number"
                  required
                  min={1}
                  max={kind === "percent" ? 99 : 10000000}
                  step={1}
                />
              </label>
              <label className="field">
                Giá sân tối thiểu (đ)
                <input
                  name="minTotal"
                  type="number"
                  min={0}
                  max={10000000}
                  defaultValue={0}
                  required
                />
              </label>
            </div>
            <div className="form-columns">
              <label className="field">
                Giảm tối đa (đ, 0 = không giới hạn)
                <input
                  name="maxDiscount"
                  type="number"
                  min={0}
                  max={10000000}
                  defaultValue={0}
                  required
                />
              </label>
              <label className="field">
                Tổng lượt (1–1.000.000)
                <input
                  name="maxUses"
                  type="number"
                  min={1}
                  max={1000000}
                  required
                />
              </label>
            </div>
            <div className="form-columns voucher-dates">
              <label className="field">
                Bắt đầu (giờ trên thiết bị)
                <input name="startsAt" type="datetime-local" required />
              </label>
              <label className="field">
                Kết thúc (giờ trên thiết bị)
                <input name="endsAt" type="datetime-local" required />
              </label>
            </div>
            <p className="muted">
              Không sửa giá trị/mã của voucher đã tạo. Hủy lịch chưa nhận cọc
              trả lượt; lịch đã nhận cọc không tự trả lượt. Tắt mã không đổi
              giá/QR của lịch cũ.
            </p>
            <button className="button" disabled={busy}>
              {busy ? "Đang lưu…" : "Tạo mã ưu đãi"}
            </button>
          </form>
        )}
        <p style={{ marginTop: 20 }}>
          <Link className="back-link" href="/tai-khoan">
            ← Tài khoản & đăng xuất phiên thật
          </Link>
        </p>
      </section>
      {authorized && (
        <section className="panel" style={{ marginTop: 24 }}>
          <div className="content-heading">
            <h2>Voucher đã tạo</h2>
            <button
              className="button button-neutral"
              disabled={busy}
              onClick={async () => {
                setBusy(true);
                setError("");
                try {
                  setItems(await api<Voucher[]>("/admin/vouchers"));
                } catch (e) {
                  setError(
                    e instanceof Error ? e.message : "Không tải được voucher.",
                  );
                } finally {
                  setBusy(false);
                }
              }}
            >
              Tải lại lượt dùng
            </button>
          </div>
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Mã</th>
                  <th>Giảm</th>
                  <th>Điều kiện</th>
                  <th>Thời gian</th>
                  <th>Lượt giữ/đã dùng</th>
                  <th>Trạng thái</th>
                </tr>
              </thead>
              <tbody>
                {items.map((v) => (
                  <tr key={v.code}>
                    <td>
                      <strong>{v.code}</strong>
                    </td>
                    <td>
                      {v.kind === "percent" ? `${v.value}%` : money(v.value)}
                      <small>
                        {v.maxDiscount
                          ? `Tối đa ${money(v.maxDiscount)}`
                          : "Không giới hạn giảm tối đa"}
                      </small>
                    </td>
                    <td>Từ {money(v.minTotal)}</td>
                    <td>
                      {new Date(v.startsAt * 1000).toLocaleString("vi-VN")}
                      <small>
                        Đến {new Date(v.endsAt * 1000).toLocaleString("vi-VN")}
                      </small>
                    </td>
                    <td>
                      {v.usedCount} / {v.maxUses}
                    </td>
                    <td>
                      <button
                        className={`button ${v.active ? "button-danger" : "button-neutral"}`}
                        disabled={busy}
                        onClick={() => toggle(v)}
                      >
                        {v.active ? "Tắt mã" : "Bật lại"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {items.length === 0 && (
            <p className="muted">
              Chưa có voucher. Không tự sinh mã mẫu trong production.
            </p>
          )}
        </section>
      )}
    </>
  );
}
