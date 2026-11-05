"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { api, ApiError, bookingStatus, type LivePayment } from "@/lib/api";
import { money } from "@/lib/data";

export function PaymentPage({ id }: { id: string }) {
  const [payment, setPayment] = useState<LivePayment | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const [canCreate, setCanCreate] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout>;
    async function refresh() {
      try {
        const p = await api<LivePayment>(
          `/bookings/${encodeURIComponent(id)}/payment`,
          { signal: controller.signal },
        );
        if (controller.signal.aborted) return;
        setPayment(p);
        setCanCreate(
          p.status === "pending" && p.bookingStatus === "pending" && !p.expired,
        );
        setError("");
        if (
          p.status === "pending" &&
          p.bookingStatus === "pending" &&
          !p.expired
        )
          timer = setTimeout(refresh, 5000);
      } catch (e) {
        if (e instanceof Error && e.name !== "AbortError") {
          setCanCreate(e instanceof ApiError && e.status === 404);
          setError(
            e instanceof ApiError && e.status === 404
              ? "Chưa có QR thanh toán. Bấm tạo QR cho lịch đang chờ cọc."
              : e.message,
          );
        }
      } finally {
        if (!controller.signal.aborted) setLoaded(true);
      }
    }
    refresh();
    return () => {
      controller.abort();
      clearTimeout(timer);
    };
  }, [id, refreshKey]);
  async function create() {
    setBusy(true);
    setError("");
    try {
      const p = await api<LivePayment>(
        `/bookings/${encodeURIComponent(id)}/payment`,
        { method: "POST", body: "{}" },
      );
      setPayment(p);
      setRefreshKey((k) => k + 1);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không thể tạo QR.");
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="panel confirmation-card">
      <span className="eyebrow">VIETQR · TINGO PAY</span>
      <h1>Thanh toán cọc sân</h1>
      <p className="muted">Mã lịch: {id}</p>
      {!loaded && <p role="status">Đang kiểm tra thanh toán…</p>}
      {error && (
        <p role="alert" className="form-error">
          {error}
        </p>
      )}
      {payment && (
        <>
          <p
            className={
              payment.environment === "sandbox"
                ? "warning-note"
                : "success-message"
            }
          >
            {payment.environment === "sandbox"
              ? "SANDBOX — không chuyển tiền thật; chỉ dùng Test Callback của VietQR."
              : "PRODUCTION — chuyển khoản thật vào tài khoản Sân Nhà."}
          </p>
          <div className="ticket">
            <p>
              Tiền sân: {money(payment.subtotal)}
              {payment.voucherCode && (
                <>
                  {" "}
                  · Mã {payment.voucherCode}: −{money(payment.discount)}
                </>
              )}
            </p>
            <p>
              Tổng sau giảm: <strong>{money(payment.total)}</strong>
            </p>
            <p>
              <strong>Số tiền cọc: {money(payment.amount)}</strong>
            </p>
            <p>Ngân hàng: {payment.bankCode}</p>
            <p>Tài khoản: {payment.bankAccount}</p>
            <p>Chủ tài khoản: {payment.bankName}</p>
            <p>
              Hạn QR:{" "}
              {new Date(payment.expiresAt * 1000).toLocaleString("vi-VN")}
            </p>
            <p>Nội dung / mã đơn: {payment.content}</p>
            <p>
              Lịch:{" "}
              {bookingStatus[payment.bookingStatus] || payment.bookingStatus}
            </p>
          </div>
          {payment.qrImage &&
            payment.status === "pending" &&
            payment.bookingStatus === "pending" && (
              <Image
                src={payment.qrImage}
                alt={`QR thanh toán cọc ${money(payment.amount)}`}
                width={320}
                height={320}
                unoptimized
                style={{ maxWidth: "100%", height: "auto" }}
              />
            )}
          {payment.status === "paid" ? (
            <p role="status" className="success-message">
              Đã nhận cọc qua callback VietQR.
              {payment.bookingStatus === "cancelled"
                ? " Lịch đã hủy; liên hệ Sân Nhà để xử lý hoàn cọc, không có hoàn tiền tự động."
                : " Lịch đã được xác nhận trên backend."}
            </p>
          ) : payment.status === "review" ? (
            <p role="status" className="warning-note">
              Giao dịch cần Sân Nhà đối soát (sai số tiền/tài khoản hoặc lịch đã
              đổi). Không chuyển thêm tiền, không tự đặt lại.
            </p>
          ) : (
            <p role="status" className="muted">
              {payment.expired
                ? "QR đã quá hạn hiển thị. Không chuyển tiền, liên hệ Sân Nhà để kiểm tra hoặc hủy lịch. Ngừng hiển thị không hủy được QR trong ứng dụng ngân hàng."
                : "Chưa ghi nhận cọc. Trang tự kiểm tra mỗi 5 giây khi lịch còn chờ. Không có nút tự xác nhận đã thanh toán."}
            </p>
          )}
        </>
      )}
      {loaded &&
        canCreate &&
        (!payment || !payment.qrImage) &&
        payment?.status !== "paid" &&
        payment?.status !== "review" &&
        (!payment ||
          (payment.bookingStatus === "pending" && !payment.expired)) && (
          <button className="button" disabled={busy} onClick={create}>
            {busy ? "Đang tạo QR…" : "Tạo / thử lại QR cọc"}
          </button>
        )}
      {loaded && (
        <button
          className="button button-neutral"
          disabled={busy}
          onClick={() => setRefreshKey((k) => k + 1)}
          style={{ marginTop: 16 }}
        >
          Kiểm tra lại trạng thái
        </button>
      )}
      <p className="photo-note">
        Không thanh toán cho mã lịch demo. Không tự chuyển khoản khi QR lỗi.
        Callback đến sau khi hủy lịch chỉ được lưu để đối soát, không phục hồi
        lịch.
      </p>
      <Link className="back-link" href="/tai-khoan">
        ← Lịch đặt & tài khoản thật
      </Link>
    </section>
  );
}
