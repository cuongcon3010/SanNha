"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  api,
  bookingStatus,
  type LiveBooking,
  type LiveQuote,
} from "@/lib/api";
import { dateLabel, money, timeLabel, today } from "@/lib/data";

type LiveVenue = {
  id: string;
  name: string;
  address: string;
  courts: { id: string; name: string; maintenance: boolean }[];
};
type Slot = {
  courtId: string;
  hour: number;
  price: number;
  available: boolean;
};

export function LiveBookingForm() {
  const router = useRouter();
  const [venues, setVenues] = useState<LiveVenue[]>([]);
  const [venueId, setVenueId] = useState("");
  const [courtId, setCourtId] = useState("");
  const [date, setDate] = useState("");
  const [slots, setSlots] = useState<Slot[]>([]);
  const [hour, setHour] = useState<number | null>(null);
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [paymentEnv, setPaymentEnv] = useState("");
  const [voucherCode, setVoucherCode] = useState("");
  const [quote, setQuote] = useState<{ key: string; value: LiveQuote } | null>(
    null,
  );
  const [quoting, setQuoting] = useState(false);
  const [voucherError, setVoucherError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    Promise.all([
      api<LiveVenue[]>("/venues", { signal: controller.signal }),
      api<{ name: string; phone: string }>("/me", {
        signal: controller.signal,
      }),
      api<{ enabled: boolean; environment: string }>("/payments/config", {
        signal: controller.signal,
      }),
    ])
      .then(([catalog, me, payments]) => {
        setVenues(catalog);
        setName(me.name);
        setPhone(me.phone);
        setDate(today());
        if (catalog[0]) {
          setVenueId(catalog[0].id);
          setCourtId(catalog[0].courts[0]?.id || "");
        }
        setPaymentEnv(payments.enabled ? payments.environment : "");
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
  useEffect(() => {
    if (!venueId || !date) return;
    const controller = new AbortController();
    api<Slot[]>(
      `/venues/${encodeURIComponent(venueId)}/availability?date=${date}`,
      { signal: controller.signal },
    )
      .then(setSlots)
      .catch((e) => {
        if (e.name !== "AbortError") setError(e.message);
      });
    return () => controller.abort();
  }, [venueId, date]);
  const venue = venues.find((v) => v.id === venueId);
  const selected = slots.find(
    (s) => s.courtId === courtId && s.hour === hour && s.available,
  );
  const quoteInput = {
    courtId,
    date,
    hour,
    voucherCode: voucherCode.trim().toUpperCase(),
  };
  const quoteKey = JSON.stringify(quoteInput);
  const preview = quote?.key === quoteKey ? quote.value : null;
  async function applyVoucher() {
    if (!selected || quoting) return;
    setQuoting(true);
    setVoucherError("");
    try {
      const value = await api<LiveQuote>("/bookings/quote", {
        method: "POST",
        body: JSON.stringify(quoteInput),
      });
      setQuote({ key: quoteKey, value });
    } catch (e) {
      setQuote(null);
      setVoucherError(
        e instanceof Error ? e.message : "Không áp dụng được voucher.",
      );
    } finally {
      setQuoting(false);
    }
  }
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!selected || busy || (voucherCode.trim() && !preview)) return;
    if (
      !window.confirm(
        "Tạo lịch trên hệ thống thật và chuyển sang trang thanh toán cọc?",
      )
    )
      return;
    setBusy(true);
    setError("");
    try {
      const b = await api<LiveBooking>("/bookings", {
        method: "POST",
        body: JSON.stringify({
          courtId,
          date,
          hour,
          name,
          phone,
          voucherCode: preview?.voucherCode || "",
        }),
      });
      router.push(`/thanh-toan/${b.id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không thể tạo lịch.");
      if (voucherCode.trim()) setQuote(null);
      setBusy(false);
    }
  }
  return (
    <section className="panel">
      <span className="eyebrow">DỮ LIỆU API · KHÔNG PHẢI DEMO</span>
      <h1>Đặt sân & thanh toán cọc</h1>
      <p className="muted">
        Cần đăng nhập tài khoản thật. Tiền cọc nhận vào tài khoản merchant chung
        của Sân Nhà.
      </p>
      {error && (
        <p role="alert" className="form-error">
          {error}
        </p>
      )}
      {loading ? (
        <p role="status">Đang tải dữ liệu…</p>
      ) : (
        <>
          {paymentEnv && (
            <p
              className={
                paymentEnv === "sandbox" ? "warning-note" : "success-message"
              }
            >
              {paymentEnv === "sandbox"
                ? "SANDBOX: chỉ dùng giao dịch giả lập của VietQR, không chuyển tiền thật."
                : "PRODUCTION: chuyển tiền cọc thật. Kiểm tra kỹ tài khoản nhận trước khi trả tiền."}
            </p>
          )}
          {!paymentEnv && (
            <p className="warning-note">
              Thanh toán chưa được cấu hình. Không tạo QR hoặc yêu cầu chuyển
              tiền.
            </p>
          )}
          {venues.length > 0 ? (
            <form className="stack-form" onSubmit={submit}>
              <label className="field">
                Cơ sở sân
                <select
                  value={venueId}
                  onChange={(e) => {
                    const v = venues.find((v) => v.id === e.target.value);
                    setVenueId(e.target.value);
                    setCourtId(v?.courts[0]?.id || "");
                    setSlots([]);
                    setHour(null);
                    setQuote(null);
                    setVoucherError("");
                    setError("");
                  }}
                >
                  {venues.map((v) => (
                    <option key={v.id} value={v.id}>
                      {v.name}
                    </option>
                  ))}
                </select>
              </label>
              <p className="muted">{venue?.address}</p>
              <div className="form-columns">
                <label className="field">
                  Sân con
                  <select
                    value={courtId}
                    onChange={(e) => {
                      setCourtId(e.target.value);
                      setHour(null);
                      setQuote(null);
                      setVoucherError("");
                    }}
                  >
                    {venue?.courts.map((c) => (
                      <option key={c.id} value={c.id} disabled={c.maintenance}>
                        {c.name}
                        {c.maintenance ? " · Bảo trì" : ""}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="field">
                  Ngày chơi
                  <input
                    type="date"
                    required
                    value={date}
                    min={date ? today() : undefined}
                    onChange={(e) => {
                      setDate(e.target.value);
                      setSlots([]);
                      setHour(null);
                      setQuote(null);
                      setVoucherError("");
                    }}
                  />
                </label>
              </div>
              <fieldset>
                <legend>Khung giờ (60 phút)</legend>
                <div className="time-slots">
                  {slots
                    .filter((s) => s.courtId === courtId)
                    .map((s) => (
                      <button
                        type="button"
                        key={s.hour}
                        disabled={!s.available}
                        aria-pressed={hour === s.hour}
                        className={
                          hour === s.hour
                            ? "selected"
                            : s.available
                              ? ""
                              : "unavailable"
                        }
                        onClick={() => {
                          setHour(s.hour);
                          setQuote(null);
                          setVoucherError("");
                        }}
                      >
                        <strong>{timeLabel(s.hour)}</strong>
                        <span>
                          {s.available ? money(s.price) : "Không nhận đặt"}
                        </span>
                      </button>
                    ))}
                </div>
              </fieldset>
              <div className="form-columns">
                <label className="field">
                  Họ tên người đặt
                  <input
                    value={name}
                    required
                    minLength={2}
                    maxLength={80}
                    onChange={(e) => setName(e.target.value)}
                  />
                </label>
                <label className="field">
                  Điện thoại
                  <input
                    type="tel"
                    pattern="(?:0[0-9]{9}|\+84[0-9]{9})"
                    value={phone}
                    required
                    onChange={(e) => setPhone(e.target.value)}
                  />
                </label>
              </div>
              <fieldset>
                <legend>Mã voucher (không bắt buộc)</legend>
                <div className="toolbar">
                  <label className="field" style={{ flex: 1 }}>
                    Mã giảm giá
                    <input
                      value={voucherCode}
                      maxLength={24}
                      placeholder="VD: SANNHA50"
                      autoComplete="off"
                      onChange={(e) => {
                        setVoucherCode(e.target.value);
                        setVoucherError("");
                        setQuote(null);
                      }}
                    />
                  </label>
                  <button
                    className="button button-neutral"
                    type="button"
                    disabled={!selected || !voucherCode.trim() || quoting}
                    onClick={applyVoucher}
                  >
                    {quoting ? "Đang kiểm tra…" : "Áp dụng mã"}
                  </button>
                  {voucherCode && (
                    <button
                      className="text-link"
                      type="button"
                      onClick={() => {
                        setVoucherCode("");
                        setQuote(null);
                        setVoucherError("");
                      }}
                    >
                      Bỏ mã
                    </button>
                  )}
                </div>
                {voucherError && (
                  <p role="alert" className="form-error">
                    {voucherError}
                  </p>
                )}
                {preview?.voucherCode && (
                  <p role="status" className="success-message">
                    Mã {preview.voucherCode}: giảm {money(preview.discount)}.
                    Lượt dùng chỉ được giữ khi tạo lịch thành công.
                  </p>
                )}
                {voucherCode.trim() && !preview && !voucherError && (
                  <p className="muted">
                    Bấm áp dụng để kiểm tra mã. Đổi sân/ngày/giờ cần áp dụng
                    lại.
                  </p>
                )}
              </fieldset>
              {selected && (
                <p>
                  {preview?.discount ? (
                    <>
                      Giá gốc: {money(preview.subtotal)} · Giảm:{" "}
                      {money(preview.discount)} ·{" "}
                    </>
                  ) : null}
                  Tổng dự kiến:{" "}
                  <strong>{money(preview?.total ?? selected.price)}</strong> ·
                  Cọc 30%:{" "}
                  <strong>
                    {money(
                      preview?.deposit ?? Math.round(selected.price * 0.3),
                    )}
                  </strong>
                  . Server kiểm tra lại mã/lượt dùng và giá khi tạo lịch. Tổng
                  sau giảm tối thiểu 1.000đ.
                </p>
              )}
              <button
                className="button"
                disabled={
                  !selected ||
                  !paymentEnv ||
                  busy ||
                  quoting ||
                  (!!voucherCode.trim() && !preview)
                }
              >
                {busy ? "Đang tạo lịch…" : "Tạo lịch thật & thanh toán cọc ↗"}
              </button>
            </form>
          ) : (
            <p className="muted">
              Chưa có sân được duyệt trong backend hoặc bạn chưa đăng nhập.{" "}
              <Link href="/dang-nhap">Đăng nhập Google/GitHub ↗</Link>
            </p>
          )}
        </>
      )}
    </section>
  );
}

export function LiveBookings() {
  const [rows, setRows] = useState<LiveBooking[] | null>(null);
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    api<LiveBooking[]>("/me/bookings", { signal: controller.signal })
      .then(setRows)
      .catch((e) => {
        if (e.name !== "AbortError") setError(e.message);
      });
    return () => controller.abort();
  }, []);
  return (
    <section className="panel" style={{ marginTop: 24 }}>
      <div className="content-heading">
        <h2>Lịch đặt trên backend</h2>
        <Link href="/dat-san" className="button">
          Đặt sân thật ↗
        </Link>
      </div>
      {error && (
        <p role="alert" className="form-error">
          {error}
        </p>
      )}
      {rows === null && !error && <p>Đang tải lịch…</p>}
      {rows?.length === 0 && (
        <p className="muted">
          Chưa có lịch. Lịch demo localStorage không được thu tiền.
        </p>
      )}
      {rows?.map((b) => (
        <div key={b.id} className="pending-row">
          <div>
            <strong>
              {dateLabel(b.date)} · {timeLabel(b.hour)}
            </strong>
            <p>
              {b.id} · {bookingStatus[b.status] || b.status} · Cọc{" "}
              {money(b.deposit)} ·{" "}
              {b.depositPaid ? "Đã nhận cọc" : "Chưa nhận cọc"}
              {b.voucherCode && (
                <>
                  {" "}
                  · Mã {b.voucherCode}: giảm {money(b.discount)}
                </>
              )}
            </p>
          </div>
          <Link className="button button-neutral" href={`/thanh-toan/${b.id}`}>
            Xem thanh toán ↗
          </Link>
        </div>
      ))}
    </section>
  );
}
