"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import {
  CalendarDays,
  ArrowUpRight,
  MapPin,
  Plus,
  Heart,
  ShieldCheck,
} from "lucide-react";
import {
  dateLabel,
  deposit,
  money,
  timeLabel,
  venues,
  type Booking,
} from "@/lib/data";
import { useDemo } from "@/lib/demo-store";
import { Badge, Empty, SuccessMessage } from "../ui";
import { VenueCard } from "../venue-card";
import { Modal, PageHeading, Stats } from "./shell";

export function UserDashboard({ section }: { section: string }) {
  if (section === "ho-so") return <Profile />;
  if (section === "yeu-thich") return <Favorites />;
  return <Bookings />;
}

function Bookings() {
  const { bookings, setStatus } = useDemo();
  const [filter, setFilter] = useState("Tất cả");
  const [selected, setSelected] = useState<Booking | null>(null);
  const mine = bookings.filter(
    (b) => b.playerBooking && b.status !== "Khóa giờ",
  );
  const upcoming = mine.filter(
    (b) => b.status === "Đã xác nhận" || b.status === "Chờ cọc",
  );
  const visible = mine.filter(
    (b) =>
      filter === "Tất cả" ||
      (filter === "Sắp diễn ra"
        ? b.status === "Đã xác nhận" || b.status === "Chờ cọc"
        : b.status === filter),
  );
  const next = upcoming.toSorted(
    (a, b) => a.date.localeCompare(b.date) || a.hour - b.hour,
  )[0];
  function cancel() {
    if (
      selected &&
      window.confirm(
        "Hủy lịch demo này? Không có khoản tiền thật nào được hoàn hoặc thu.",
      )
    ) {
      setStatus(selected.id, "Đã hủy");
      setSelected(null);
    }
  }
  return (
    <>
      <PageHeading
        eyebrow="CHÀO AN, SẴN SÀNG RA SÂN CHƯA?"
        title="Kèo tiếp theo đang chờ bạn."
        description="Xem lịch đặt, theo dõi tiền cọc và quản lý các cuộc chơi."
      />
      <Stats
        items={[
          {
            label: "Sắp diễn ra",
            value: String(upcoming.length).padStart(2, "0"),
            caption: "Lịch đặt đang hoạt động",
          },
          {
            label: "Đã hoàn thành",
            value: String(
              mine.filter((b) => b.status === "Hoàn thành").length,
            ).padStart(2, "0"),
            caption: "Những cuộc chơi hết mình",
          },
          {
            label: "Cọc chờ xác nhận",
            value: money(
              mine
                .filter((b) => b.status === "Chờ cọc")
                .reduce((s, b) => s + deposit(b.total), 0),
            ),
            caption: "Số tiền minh họa, chưa thanh toán",
          },
          {
            label: "Tổng lịch đặt",
            value: String(mine.length).padStart(2, "0"),
            caption: "Được lưu trong trình duyệt",
          },
        ]}
      />
      {next && (
        <section className="upcoming-card dotted-dark">
          <div>
            <Badge tone="yellow">SẮP DIỄN RA</Badge>
            <h2>{venues.find((v) => v.id === next.venueId)?.name}</h2>
            <p>
              <CalendarDays size={16} />
              {dateLabel(next.date)} · {timeLabel(next.hour)}
              <span>·</span>
              {next.court}
            </p>
          </div>
          <button className="button" onClick={() => setSelected(next)}>
            Xem thông tin <ArrowUpRight size={16} />
          </button>
        </section>
      )}
      <div className="content-heading">
        <h2>Lịch đặt của tôi</h2>
        <Link href="/" className="button">
          <Plus size={16} />
          Đặt sân mới
        </Link>
      </div>
      <div className="filter-row">
        {["Tất cả", "Sắp diễn ra", "Hoàn thành", "Đã hủy"].map((f) => (
          <button
            key={f}
            className={`filter-chip ${filter === f ? "active" : ""}`}
            onClick={() => setFilter(f)}
            aria-pressed={filter === f}
          >
            {f}
          </button>
        ))}
      </div>
      <div className="table-panel">
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Mã đặt / Sân</th>
                <th>Thời gian</th>
                <th>Tổng tiền</th>
                <th>Trạng thái</th>
                <th>
                  <span className="sr-only">Thao tác</span>
                </th>
              </tr>
            </thead>
            <tbody>
              {visible.map((b) => (
                <tr key={b.id}>
                  <td>
                    <strong>
                      {venues.find((v) => v.id === b.venueId)?.name}
                    </strong>
                    <small>
                      {b.id} · {b.court}
                    </small>
                  </td>
                  <td>
                    {dateLabel(b.date)}
                    <small>{timeLabel(b.hour)}</small>
                  </td>
                  <td className="no-wrap">
                    <strong>{money(b.total)}</strong>
                  </td>
                  <td>
                    <Badge
                      tone={
                        b.status === "Chờ cọc"
                          ? "yellow"
                          : b.status === "Đã hủy"
                            ? "red"
                            : "neutral"
                      }
                    >
                      {b.status}
                    </Badge>
                  </td>
                  <td>
                    <button
                      className="table-link"
                      onClick={() => setSelected(b)}
                    >
                      Chi tiết <ArrowUpRight size={14} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {visible.length === 0 && <Empty title="Chưa có lịch trong mục này" />}
        <div className="table-footer">
          {visible.length} lịch đặt · Dữ liệu demo
          <span>Thông tin chỉ lưu ở trình duyệt này</span>
        </div>
      </div>
      {selected && (
        <Modal title="Thông tin lịch đặt" onClose={() => setSelected(null)}>
          <div className="booking-modal-info">
            <Badge tone="yellow">{selected.status}</Badge>
            <h3>{venues.find((v) => v.id === selected.venueId)?.name}</h3>
            <p>
              <CalendarDays size={17} />
              {dateLabel(selected.date)} · {timeLabel(selected.hour)}
            </p>
            <p>
              <MapPin size={17} />
              {selected.court} · {selected.name}
            </p>
            <p>
              Tổng tiền mẫu: <strong>{money(selected.total)}</strong>
            </p>
            <small>Mã lịch: {selected.id}</small>
          </div>
          <p className="muted">
            Đây là lịch minh họa, không giữ sân và không phát sinh thanh toán.
          </p>
          <div className="modal-actions">
            <Link className="button" href={`/san/${selected.venueId}`}>
              Xem sân ↗
            </Link>
            {["Đã xác nhận", "Chờ cọc"].includes(selected.status) && (
              <button className="button button-danger" onClick={cancel}>
                Hủy lịch demo
              </button>
            )}
          </div>
        </Modal>
      )}
    </>
  );
}

function Profile() {
  const { profile, saveProfile } = useDemo();
  const [saved, setSaved] = useState(false);
  function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    saveProfile({
      name: String(f.get("name")).trim(),
      email: String(f.get("email")).trim(),
      phone: String(f.get("phone")).trim(),
      area: String(f.get("area")).trim(),
    });
    setSaved(true);
  }
  return (
    <>
      <PageHeading
        eyebrow="TÀI KHOẢN / NGƯỜI CHƠI"
        title="Tài khoản của bạn."
        description="Thông tin liên hệ và những sân bạn muốn quay lại."
      />
      <div className="profile-grid">
        <form
          className="panel"
          onSubmit={submit}
          onChange={() => setSaved(false)}
        >
          <h2>Thông tin cá nhân</h2>
          <p className="muted">
            Chỉ dùng thông tin giả để trải nghiệm giao diện.
          </p>
          {[
            ["name", "Họ và tên", "text"],
            ["email", "Email", "email"],
            ["phone", "Số điện thoại", "tel"],
            ["area", "Khu vực thường chơi", "text"],
          ].map(([name, label, type]) => (
            <label className="field" key={name}>
              {label}
              <input
                key={profile[name as keyof typeof profile]}
                name={name}
                type={type}
                defaultValue={profile[name as keyof typeof profile]}
                minLength={name === "name" ? 2 : undefined}
                maxLength={150}
                pattern={name === "phone" ? "0[0-9]{9}" : undefined}
                required
                autoComplete="off"
              />
            </label>
          ))}
          <button className="button" type="submit">
            Lưu thay đổi <ArrowUpRight size={16} />
          </button>
          {saved && (
            <SuccessMessage>
              Đã lưu hồ sơ demo trong trình duyệt.
            </SuccessMessage>
          )}
        </form>
        <div>
          <FavoriteList />
          <section className="panel security-panel">
            <ShieldCheck size={25} />
            <h2>Bảo mật tài khoản</h2>
            <p>
              Không chia sẻ mật khẩu hoặc mã OTP. Bản demo không lưu mật khẩu.
            </p>
            <Link href="/quen-mat-khau" className="button">
              Khôi phục mật khẩu ↗
            </Link>
          </section>
        </div>
      </div>
    </>
  );
}
function FavoriteList() {
  const { favorites, toggleFavorite } = useDemo();
  return (
    <section className="panel favorite-list">
      <h2>Sân yêu thích</h2>
      {venues
        .filter((v) => favorites.includes(v.id))
        .map((v) => (
          <div key={v.id}>
            <button
              className="icon-button text-red"
              aria-label={`Bỏ thích ${v.name}`}
              onClick={() => toggleFavorite(v.id)}
            >
              <Heart size={19} fill="currentColor" />
            </button>
            <Link href={`/san/${v.id}`}>
              <strong>{v.name}</strong>
              <small>
                {v.sport} · {money(v.price)} / giờ
              </small>
            </Link>
          </div>
        ))}
      {favorites.length === 0 && (
        <p className="muted">Chưa có sân yêu thích.</p>
      )}
      <Link href="/" className="button full-width">
        Khám phá thêm sân ↗
      </Link>
    </section>
  );
}
function Favorites() {
  const { favorites } = useDemo();
  return (
    <>
      <PageHeading
        eyebrow="NHỮNG ĐIỂM HẸN QUEN THUỘC"
        title="Lưu sân hay. Hẹn kèo mới."
        description="Chạm trái tim để lưu hoặc bỏ lưu sân yêu thích của bạn."
      />
      <div className="venue-grid dashboard-venues">
        {venues
          .filter((v) => favorites.includes(v.id))
          .map((v) => (
            <VenueCard venue={v} key={v.id} />
          ))}
      </div>
      {favorites.length === 0 && (
        <div className="panel">
          <Empty
            title="Chưa có sân yêu thích"
            description="Khám phá và lưu những sân bạn muốn quay lại."
          />
          <Link className="button" href="/">
            Tìm sân ↗
          </Link>
        </div>
      )}
    </>
  );
}
