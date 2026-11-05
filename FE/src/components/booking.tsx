"use client";

import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import {
  MapPin,
  Star,
  Clock3,
  CalendarDays,
  Check,
  ArrowRight,
  Car,
  Lightbulb,
  ShieldCheck,
  ChevronLeft,
  Heart,
} from "lucide-react";
import {
  bookingError,
  courtPrice,
  dateLabel,
  demoDate,
  deposit,
  hours,
  isAvailable,
  isDate,
  money,
  timeLabel,
  today,
  venues,
  type Venue,
} from "@/lib/data";
import { useDemo } from "@/lib/demo-store";
import { Header } from "./header";
import { DemoNotice, Footer, Badge, Empty } from "./ui";

export function BookingPage({
  venue,
  initialDate,
  allHours = false,
}: {
  venue: Venue;
  initialDate?: string;
  allHours?: boolean;
}) {
  const router = useRouter();
  const {
    bookings,
    addBooking,
    profile,
    favorites,
    toggleFavorite,
    storageError,
    rates,
  } = useDemo();
  const minDate = today();
  const [date, setDate] = useState(
    initialDate && isDate(initialDate) && initialDate >= minDate
      ? initialDate
      : minDate > demoDate
        ? minDate
        : demoDate,
  );
  const [court, setCourt] = useState(venue.courts[0]);
  const [hour, setHour] = useState<number | null>(null);
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const total = courtPrice(venue, court, hour ?? 19, rates);
  const maintenance = venue.id === "thanh-dong" && rates[court]?.maintenance;
  const dateOptions = Array.from({ length: 7 }, (_, i) => {
    const d = new Date(date + "T12:00:00");
    d.setDate(d.getDate() + i);
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
  });
  function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (hour === null) {
      setError("Vui lòng chọn một khung giờ còn trống.");
      return;
    }
    const f = new FormData(e.currentTarget);
    const data = {
      venueId: venue.id,
      court,
      date,
      hour,
      name: String(f.get("name") || "").trim(),
      phone: String(f.get("phone") || "").trim(),
    };
    const message = maintenance
      ? "Sân này đang bảo trì. Vui lòng chọn sân khác."
      : bookingError(data, bookings);
    if (message) {
      setError(message);
      return;
    }
    const id = "SN-" + crypto.randomUUID().slice(0, 8).toUpperCase();
    setPending(true);
    addBooking({
      ...data,
      id,
      status: "Đã xác nhận",
      total,
      playerBooking: true,
    });
    router.push(`/xac-nhan?id=${id}`);
  }
  return (
    <>
      <Header />
      <main id="main-content" className="container detail-page">
        <nav className="breadcrumbs" aria-label="Đường dẫn">
          <Link href="/">Khám phá sân</Link>
          <span>/</span>
          <span>{venue.sport}</span>
          <span>/</span>
          <span>{venue.name}</span>
        </nav>
        <div className="detail-heading">
          <div>
            <span className="eyebrow">KÈO HAY BẮT ĐẦU Ở ĐÂY</span>
            <h1>{venue.name}</h1>
            <p>
              <Star size={15} fill="currentColor" />{" "}
              <strong>{venue.rating}</strong> <span>(128 đánh giá mẫu)</span>
              <span className="separator">·</span>
              <MapPin size={15} />
              {venue.address}
            </p>
          </div>
          <button
            className={`button button-white ${favorites.includes(venue.id) ? "is-liked" : ""}`}
            onClick={() => toggleFavorite(venue.id)}
            aria-pressed={favorites.includes(venue.id)}
          >
            <Heart
              size={16}
              fill={favorites.includes(venue.id) ? "currentColor" : "none"}
            />
            {favorites.includes(venue.id) ? "Đã yêu thích" : "Yêu thích"}
          </button>
        </div>
        <div className="booking-layout">
          <div className="booking-main">
            <div className="detail-image">
              <Image
                src={venue.image}
                fill
                priority
                sizes="(max-width: 900px) 100vw, 65vw"
                alt={`Ảnh thể thao minh họa cho ${venue.name}`}
              />
              <Badge tone="yellow">
                {venue.sport} · {venue.courts.length} sân con
              </Badge>
              <span className="image-caption">Ảnh minh họa thể thao</span>
            </div>
            <div className="amenities">
              {venue.amenities.map((a, i) => (
                <span key={a}>
                  {i === 1 ? (
                    <Lightbulb size={18} />
                  ) : i === 2 ? (
                    <Car size={18} />
                  ) : (
                    <Check size={18} />
                  )}
                  {a}
                </span>
              ))}
            </div>
            <section className="slot-section">
              <div className="section-heading">
                <div>
                  <span className="eyebrow">CHỌN GIỜ ĐÚNG Ý</span>
                  <h2>Chọn lịch ra sân</h2>
                </div>
                <label className="date-filter">
                  <CalendarDays size={16} />
                  <input
                    type="date"
                    aria-label="Ngày đặt sân"
                    value={date}
                    min={minDate}
                    onChange={(e) => {
                      if (e.target.value) {
                        setDate(e.target.value);
                        setHour(null);
                      }
                    }}
                  />
                </label>
              </div>
              <div className="date-options">
                {dateOptions.map((d, i) => (
                  <button
                    key={d}
                    className={date === d ? "active" : ""}
                    onClick={() => {
                      setDate(d);
                      setHour(null);
                    }}
                  >
                    <span>
                      {i === 0
                        ? "Đang xem"
                        : new Date(d + "T12:00:00").toLocaleDateString(
                            "vi-VN",
                            { weekday: "short" },
                          )}
                    </span>
                    <strong>
                      {d.slice(8)} / {d.slice(5, 7)}
                    </strong>
                  </button>
                ))}
              </div>
              <div className="court-tabs" aria-label="Chọn sân con">
                {venue.courts.map((c) => (
                  <button
                    key={c}
                    onClick={() => {
                      setCourt(c);
                      setHour(null);
                    }}
                    className={court === c ? "active" : ""}
                    aria-pressed={court === c}
                  >
                    {c}
                  </button>
                ))}
              </div>
              <div className="time-slots">
                {hours
                  .filter((h) => allHours || h >= 17)
                  .map((h) => {
                    const available =
                      !maintenance &&
                      isAvailable(bookings, venue.id, court, date, h);
                    return (
                      <button
                        key={h}
                        disabled={!available}
                        className={
                          !available
                            ? "unavailable"
                            : hour === h
                              ? "selected"
                              : ""
                        }
                        onClick={() => {
                          setHour(h);
                          setError("");
                        }}
                        aria-pressed={hour === h}
                      >
                        <strong>{timeLabel(h)}</strong>
                        <span>
                          {available
                            ? money(courtPrice(venue, court, h, rates))
                            : maintenance
                              ? "Đang bảo trì"
                              : "Đã có lịch"}
                        </span>
                      </button>
                    );
                  })}
              </div>
              <p className="slot-legend">
                <span>
                  <i />
                  Còn trống
                </span>
                <span>
                  <i className="yellow" />
                  Đang chọn
                </span>
                <span>
                  <i className="red" />
                  Đã đặt / khóa giờ
                </span>
              </p>
            </section>
            <section className="booking-notes">
              <ShieldCheck size={24} />
              <div>
                <h3>Lưu ý trước khi đặt</h3>
                <p>Đến trước 15 phút. Liên hệ chủ sân nếu cần đổi lịch.</p>
                <p>
                  Bản demo không thu tiền và không đặt chỗ thực tế. Giờ trống
                  chỉ phản ánh dữ liệu của trình duyệt này.
                </p>
              </div>
            </section>
          </div>
          <aside>
            <form onSubmit={submit} className="booking-summary panel">
              <span className="eyebrow">KÈO CỦA BẠN</span>
              <h2>
                {money(total)}
                <small> / giờ</small>
              </h2>
              <div className="summary-details">
                <p>
                  <MapPin size={17} />
                  {court} · {venue.sport}
                </p>
                <p>
                  <CalendarDays size={17} />
                  {dateLabel(date)}
                </p>
                <p>
                  <Clock3 size={17} />
                  {hour !== null
                    ? `${timeLabel(hour)} · 60 phút`
                    : "Chưa chọn khung giờ"}
                </p>
              </div>
              <div className="summary-form">
                <span className="eyebrow">THÔNG TIN NGƯỜI ĐẶT (MINH HỌA)</span>
                <label className="field">
                  Họ và tên
                  <input
                    name="name"
                    defaultValue={profile.name}
                    minLength={2}
                    maxLength={80}
                    required
                    autoComplete="off"
                  />
                </label>
                <label className="field">
                  Số điện thoại
                  <input
                    name="phone"
                    defaultValue={profile.phone}
                    type="tel"
                    required
                    autoComplete="off"
                    placeholder="0901234567"
                  />
                </label>
              </div>
              <div className="summary-total">
                <p>
                  <span>Tổng tiền sân</span>
                  <strong>{money(total)}</strong>
                </p>
                <p>
                  <span>Cọc giữ chỗ (30%)</span>
                  <strong className="text-red">{money(deposit(total))}</strong>
                </p>
              </div>
              {error && (
                <p role="alert" className="form-error">
                  {error}
                </p>
              )}
              <button
                type="submit"
                className="button full-width"
                disabled={pending}
              >
                {pending ? "Đang tạo lịch mẫu…" : "Xác nhận đặt thử nghiệm"}
                <ArrowRight size={17} />
              </button>
              <DemoNotice />
              <Link className="back-link" href="/dat-san">
                Đặt sân & thanh toán VietQR bằng tài khoản thật ↗
              </Link>
              {storageError && (
                <p className="form-error" role="status">
                  Trình duyệt không lưu được dữ liệu. Lịch chỉ có trong phiên
                  hiện tại.
                </p>
              )}
            </form>
            <div className="booking-reassurance">
              <ShieldCheck size={17} />
              <span>Giá rõ ràng. Không phát sinh phí demo.</span>
            </div>
          </aside>
        </div>
      </main>
      <Footer />
    </>
  );
}

export function Confirmation({ id }: { id?: string }) {
  const { bookings, storageError } = useDemo();
  const booking = bookings.find((b) => b.id === id);
  const venue = venues.find((v) => v.id === booking?.venueId);
  return (
    <>
      <Header />
      <main id="main-content" className="confirmation-page dotted-light">
        {booking && venue ? (
          <section className="confirmation-card panel">
            <div className="success-icon">
              <Check size={36} />
            </div>
            <span className="eyebrow">SÂN NHÀ / XÁC NHẬN DEMO</span>
            <h1>
              Kèo đã chốt.
              <br />
              Hẹn bạn ở sân!
            </h1>
            <p className="muted">Đã tạo lịch mẫu trong trình duyệt của bạn.</p>
            <div className="ticket">
              <div className="ticket-header">
                <strong>{venue.name}</strong>
                <Badge tone="yellow">{booking.status}</Badge>
              </div>
              <p>
                <MapPin size={17} />
                {booking.court} · {venue.address}
              </p>
              <p>
                <CalendarDays size={17} />
                {dateLabel(booking.date)}
                <span>·</span>
                {timeLabel(booking.hour)}
              </p>
              <div className="ticket-total">
                <span>Tổng tiền mẫu</span>
                <strong>{money(booking.total)}</strong>
              </div>
              <small>Mã mẫu: {booking.id}</small>
            </div>
            <DemoNotice />
            {storageError && (
              <p className="form-error">
                Lịch chỉ lưu trong bộ nhớ vì trình duyệt không cho lưu dữ liệu.
              </p>
            )}
            <Link className="button full-width" href="/user">
              Xem lịch đặt của tôi <ArrowRight size={17} />
            </Link>
            <Link className="back-link" href="/">
              <ChevronLeft size={15} />
              Tiếp tục khám phá sân
            </Link>
          </section>
        ) : (
          <section className="panel">
            <Empty
              title="Chưa có lịch xác nhận"
              description="Lịch chỉ tồn tại trong trình duyệt tạo nó. Hãy chọn một sân để bắt đầu."
            />
            <Link className="button" href="/">
              Khám phá sân ↗
            </Link>
          </section>
        )}
      </main>
    </>
  );
}
