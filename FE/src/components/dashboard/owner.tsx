"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import {
  CalendarDays,
  Plus,
  LockKeyhole,
  ArrowUpRight,
  Pencil,
  ShieldCheck,
} from "lucide-react";
import {
  dateLabel,
  demoDate,
  deposit,
  hours,
  isAvailable,
  money,
  timeLabel,
  venues,
  type Booking,
} from "@/lib/data";
import { useDemo } from "@/lib/demo-store";
import { Badge, Empty, SuccessMessage } from "../ui";
import { Modal, PageHeading, Stats } from "./shell";

const venue = venues[0];
export function OwnerDashboard({ section }: { section: string }) {
  if (section === "san") return <Courts />;
  if (section === "doanh-thu") return <Revenue />;
  return <Schedule />;
}

function Schedule() {
  const { bookings, addBooking, setStatus } = useDemo();
  const [date, setDate] = useState(demoDate);
  const [selected, setSelected] = useState<Booking | null>(null);
  const [locking, setLocking] = useState<{
    court: string;
    hour: number;
  } | null>(null);
  const [error, setError] = useState("");
  const day = bookings.filter(
    (b) => b.venueId === venue.id && b.date === date && b.status !== "Đã hủy",
  );
  const pending = day.filter((b) => b.status === "Chờ cọc");
  const active = day.filter((b) => b.status !== "Khóa giờ");
  function lock(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    const court = String(f.get("court")),
      hour = Number(f.get("hour"));
    if (!isAvailable(bookings, venue.id, court, date, hour)) {
      setError("Khung giờ này đã có lịch, không thể khóa.");
      return;
    }
    addBooking({
      id: `BLOCK-${crypto.randomUUID().slice(0, 8)}`,
      venueId: venue.id,
      court,
      date,
      hour,
      total: 0,
      name: String(f.get("reason") || "Bảo trì"),
      phone: "",
      status: "Khóa giờ",
      playerBooking: false,
    });
    setLocking(null);
    setError("");
  }
  return (
    <>
      <PageHeading
        eyebrow="VẬN HÀNH / THÀNH ĐỒNG"
        title="Một ngày kín lịch, trọn niềm vui."
        description={`Sân bóng Thành Đồng · Lịch demo ngày ${dateLabel(date)}`}
      />
      <Stats
        items={[
          {
            label: "Giá trị lịch trong ngày",
            value: money(active.reduce((s, b) => s + b.total, 0)),
            caption: "Giá trị mẫu, không phải tiền đã thu",
          },
          {
            label: "Lịch đặt trong ngày",
            value: String(active.length).padStart(2, "0"),
            caption: `${pending.length} lịch đang chờ xác nhận`,
          },
          {
            label: "Tỷ lệ lấp đầy",
            value: `${Math.round((day.length / 18) * 100)}%`,
            caption: "03 sân · 06 khung giờ mỗi sân",
          },
          {
            label: "Cọc cần đối soát",
            value: money(pending.reduce((s, b) => s + deposit(b.total), 0)),
            caption: "Số tiền minh họa",
          },
        ]}
      />
      <div className="content-heading">
        <h2>Lịch sân trong ngày</h2>
        <div className="toolbar">
          <label className="date-filter">
            <CalendarDays size={16} />
            <input
              aria-label="Ngày vận hành"
              type="date"
              value={date}
              onChange={(e) => {
                if (e.target.value) setDate(e.target.value);
              }}
            />
          </label>
          <Link
            className="button"
            href={`/san/${venue.id}?date=${date}&period=all`}
          >
            <Plus size={16} />
            Tạo lịch đặt
          </Link>
          <button
            className="button button-dark"
            onClick={() => {
              setLocking({ court: venue.courts[0], hour: hours[0] });
              setError("");
            }}
          >
            <LockKeyhole size={15} />
            Khóa giờ
          </button>
        </div>
      </div>
      <div className="schedule-scroll panel">
        <div className="schedule-grid">
          <div className="schedule-label">SÂN / GIỜ</div>
          {hours.map((h) => (
            <div className="schedule-time" key={h}>
              {h}:00
            </div>
          ))}
          {venue.courts.map((c) => (
            <div className="schedule-row" key={c}>
              <strong>
                {c}
                <small>{c === "Sân C" ? "5 người" : "7 người"}</small>
              </strong>
              {hours.map((h) => {
                const b = day.find((b) => b.court === c && b.hour === h);
                return (
                  <button
                    key={h}
                    className={`schedule-cell ${b ? (b.status === "Khóa giờ" ? "locked" : b.status === "Chờ cọc" ? "pending" : "booked") : ""}`}
                    onClick={() => {
                      setError("");
                      if (b) setSelected(b);
                      else setLocking({ court: c, hour: h });
                    }}
                    aria-label={`${c}, ${timeLabel(h)}, ${b?.status || "Còn trống"}`}
                  >
                    {b ? (
                      <>
                        <strong>{b.name}</strong>
                        <small>
                          {b.status === "Khóa giờ"
                            ? "Không nhận đặt"
                            : b.status}
                        </small>
                      </>
                    ) : (
                      <>
                        <Plus size={17} />
                        <small>Còn trống</small>
                      </>
                    )}
                  </button>
                );
              })}
            </div>
          ))}
        </div>
      </div>
      <div className="slot-legend">
        <span>
          <i className="yellow" />
          Đã xác nhận
        </span>
        <span>
          <i className="red" />
          Chờ cọc
        </span>
        <span>
          <i className="black" />
          Khóa giờ
        </span>
        <span>
          <i />
          Trống
        </span>
      </div>
      <section className="panel pending-panel">
        <div className="content-heading">
          <h2>
            Cần xử lý <Badge tone="yellow">{pending.length} lịch chờ</Badge>
          </h2>
        </div>
        {pending.map((b) => (
          <div className="pending-row" key={b.id}>
            <div>
              <strong>{b.name}</strong>
              <p>
                {b.id} · {b.court} · {timeLabel(b.hour)} · Cọc{" "}
                {money(deposit(b.total))}
              </p>
            </div>
            <button className="button" onClick={() => setSelected(b)}>
              Xem chi tiết <ArrowUpRight size={16} />
            </button>
          </div>
        ))}
        {pending.length === 0 && (
          <p className="muted">Không có lịch chờ xác nhận trong ngày này.</p>
        )}
      </section>
      {selected && (
        <Modal
          title={
            selected.status === "Khóa giờ"
              ? "Khung giờ đã khóa"
              : "Chi tiết lịch đặt"
          }
          onClose={() => setSelected(null)}
        >
          <div className="booking-modal-info">
            <Badge tone="yellow">{selected.status}</Badge>
            <h3>{selected.name}</h3>
            <p>
              {selected.court} · {dateLabel(selected.date)}
            </p>
            <p>
              {timeLabel(selected.hour)} · {money(selected.total)}
            </p>
            <p>{selected.phone}</p>
            <small>{selected.id}</small>
          </div>
          <div className="modal-actions">
            {selected.status === "Chờ cọc" && (
              <button
                className="button"
                onClick={() => {
                  setStatus(selected.id, "Đã xác nhận");
                  setSelected(null);
                }}
              >
                Xác nhận lịch demo
              </button>
            )}
            {selected.status === "Đã xác nhận" && (
              <button
                className="button"
                onClick={() => {
                  setStatus(selected.id, "Hoàn thành");
                  setSelected(null);
                }}
              >
                Đánh dấu hoàn thành
              </button>
            )}
            <button
              className="button button-danger"
              onClick={() => {
                if (window.confirm("Hủy lịch / mở khóa khung giờ demo này?")) {
                  setStatus(selected.id, "Đã hủy");
                  setSelected(null);
                }
              }}
            >
              {selected.status === "Khóa giờ" ? "Mở khóa giờ" : "Hủy lịch demo"}
            </button>
          </div>
        </Modal>
      )}
      {locking && (
        <Modal title="Khóa khung giờ" onClose={() => setLocking(null)}>
          <p className="muted">
            Khóa một giờ cho bảo trì hoặc nhu cầu nội bộ. Không thể khóa đè lên
            lịch đã có.
          </p>
          <form onSubmit={lock} className="stack-form">
            <label className="field">
              Sân con
              <select name="court" defaultValue={locking.court}>
                {venue.courts.map((c) => (
                  <option key={c}>{c}</option>
                ))}
              </select>
            </label>
            <label className="field">
              Khung giờ
              <select name="hour" defaultValue={locking.hour}>
                {hours.map((h) => (
                  <option value={h} key={h}>
                    {timeLabel(h)}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              Lý do
              <input
                name="reason"
                defaultValue="Bảo trì"
                maxLength={80}
                required
              />
            </label>
            {error && (
              <p className="form-error" role="alert">
                {error}
              </p>
            )}
            <button className="button" type="submit">
              Khóa giờ demo <LockKeyhole size={15} />
            </button>
          </form>
        </Modal>
      )}
    </>
  );
}

function Courts() {
  const { rates, saveRates } = useDemo();
  const [editing, setEditing] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!editing) return;
    const f = new FormData(e.currentTarget);
    saveRates({
      ...rates,
      [editing]: {
        day: Number(f.get("day")),
        evening: Number(f.get("evening")),
        maintenance: f.get("maintenance") === "on",
      },
    });
    setEditing(null);
    setSaved(true);
  }
  return (
    <>
      <PageHeading
        eyebrow="THÔNG TIN CƠ SỞ / SÂN CON"
        title="Sân đẹp. Giá rõ. Dễ quản lý."
        description="Quản lý bảng giá và trạng thái sân con trong dữ liệu demo."
      />
      <section className="panel venue-settings">
        <div>
          <Badge tone="yellow">Đang hoạt động</Badge>
          <h2>{venue.name}</h2>
          <p>{venue.address}</p>
          <p>05:00 – 23:00 · 03 sân con · Bãi đỗ xe, đèn LED, phòng thay đồ</p>
        </div>
        <Link className="button button-neutral" href={`/san/${venue.id}`}>
          Xem trang sân <ArrowUpRight size={16} />
        </Link>
      </section>
      <div className="content-heading">
        <h2>Sân con & bảng giá</h2>
        <span className="muted">Nhấn chỉnh sửa để cập nhật</span>
      </div>
      {saved && (
        <SuccessMessage>
          Đã cập nhật giá demo. Giá lịch đã đặt không thay đổi.
        </SuccessMessage>
      )}
      <div className="table-panel">
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Tên sân</th>
                <th>Loại sân</th>
                <th>16:00 – 17:00</th>
                <th>17:00 – 22:00</th>
                <th>Trạng thái</th>
                <th>Thao tác</th>
              </tr>
            </thead>
            <tbody>
              {venue.courts.map((c) => (
                <tr key={c}>
                  <td>
                    <strong>{c}</strong>
                  </td>
                  <td>{c === "Sân C" ? "5 người" : "7 người"}</td>
                  <td>{money(rates[c].day)}</td>
                  <td>{money(rates[c].evening)}</td>
                  <td>
                    <Badge tone={rates[c].maintenance ? "red" : "neutral"}>
                      {rates[c].maintenance ? "Bảo trì" : "Hoạt động"}
                    </Badge>
                  </td>
                  <td>
                    <button
                      className="table-link"
                      onClick={() => {
                        setEditing(c);
                        setSaved(false);
                      }}
                    >
                      <Pencil size={14} /> Chỉnh sửa
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      <div className="warning-note">
        <ShieldCheck size={22} />
        <div>
          <strong>Giá mới chỉ áp dụng cho lịch đặt sau khi cập nhật.</strong>
          <p>
            Lịch đã xác nhận giữ nguyên giá. Sân bảo trì sẽ không nhận lịch mới
            trong bản demo.
          </p>
        </div>
      </div>
      {editing && (
        <Modal title={`Chỉnh sửa ${editing}`} onClose={() => setEditing(null)}>
          <form className="stack-form" onSubmit={submit}>
            <label className="field">
              Giá 16:00 – 17:00 (đ/giờ)
              <input
                name="day"
                type="number"
                min={1000}
                max={10000000}
                step={1000}
                required
                defaultValue={rates[editing].day}
              />
            </label>
            <label className="field">
              Giá 17:00 – 22:00 (đ/giờ)
              <input
                name="evening"
                type="number"
                min={1000}
                max={10000000}
                step={1000}
                required
                defaultValue={rates[editing].evening}
              />
            </label>
            <label className="checkbox-label">
              <input
                name="maintenance"
                type="checkbox"
                defaultChecked={rates[editing].maintenance}
              />
              Tạm dừng nhận lịch để bảo trì
            </label>
            <button className="button" type="submit">
              Lưu bảng giá demo ↗
            </button>
          </form>
        </Modal>
      )}
    </>
  );
}

function Revenue() {
  const { bookings } = useDemo();
  const [month, setMonth] = useState("2026-11");
  const rows = bookings.filter(
    (b) =>
      b.venueId === venue.id &&
      b.date.startsWith(month) &&
      b.status !== "Khóa giờ" &&
      b.status !== "Đã hủy",
  );
  const completed = rows
    .filter((b) => b.status === "Hoàn thành")
    .reduce((s, b) => s + b.total, 0);
  const confirmed = rows
    .filter((b) => b.status === "Đã xác nhận")
    .reduce((s, b) => s + b.total, 0);
  const weeks = [0, 1, 2, 3].map((w) =>
    rows
      .filter(
        (b) => Math.min(3, Math.floor((Number(b.date.slice(8)) - 1) / 7)) === w,
      )
      .reduce((s, b) => s + b.total, 0),
  );
  return (
    <>
      <PageHeading
        eyebrow="BÁO CÁO / ĐỐI SOÁT"
        title="Nắm dòng tiền. An tâm vận hành."
        description="Giá trị lịch mẫu tại Thành Đồng. Không phản ánh giao dịch hoặc tiền thật."
        action={
          <label className="date-filter">
            <CalendarDays size={17} />
            <input
              type="month"
              aria-label="Tháng báo cáo"
              value={month}
              onChange={(e) => {
                if (e.target.value) setMonth(e.target.value);
              }}
            />
          </label>
        }
      />
      <Stats
        items={[
          {
            label: "Tổng giá trị lịch",
            value: money(rows.reduce((s, b) => s + b.total, 0)),
            caption: "Không bao gồm lịch hủy / khóa",
          },
          {
            label: "Giá trị đã hoàn thành",
            value: money(completed),
            caption: "Dữ liệu demo, không phải tiền đã thu",
          },
          {
            label: "Giá trị đã xác nhận",
            value: money(confirmed),
            caption: "Chưa đánh dấu hoàn thành",
          },
          {
            label: "Cọc đang chờ",
            value: money(
              rows
                .filter((b) => b.status === "Chờ cọc")
                .reduce((s, b) => s + deposit(b.total), 0),
            ),
            caption: "Minh họa 30% giá tiền sân",
          },
        ]}
      />
      <section className="panel revenue-chart">
        <h2>Giá trị lịch theo tuần</h2>
        <p className="muted">
          Tổng giá tiền các lịch không bị hủy trong tháng.
        </p>
        {weeks.map((v, i) => (
          <div className="horizontal-bar" key={i}>
            <span>Tuần {i + 1}</span>
            <div>
              <i
                style={{
                  width: `${v ? Math.max(3, (v / Math.max(...weeks, 1)) * 100) : 0}%`,
                }}
              />
            </div>
            <strong>{money(v)}</strong>
          </div>
        ))}
      </section>
      <div className="content-heading">
        <h2>Chi tiết lịch trong tháng</h2>
        <Badge>{rows.length} lịch</Badge>
      </div>
      <div className="table-panel">
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Mã lịch</th>
                <th>Ngày / sân</th>
                <th>Người đặt</th>
                <th>Giá trị</th>
                <th>Trạng thái</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((b) => (
                <tr key={b.id}>
                  <td>
                    <strong>{b.id}</strong>
                  </td>
                  <td>
                    {dateLabel(b.date)}
                    <small>
                      {b.court} · {timeLabel(b.hour)}
                    </small>
                  </td>
                  <td>{b.name}</td>
                  <td>
                    <strong>{money(b.total)}</strong>
                  </td>
                  <td>
                    <Badge tone={b.status === "Chờ cọc" ? "yellow" : "neutral"}>
                      {b.status}
                    </Badge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {rows.length === 0 && <Empty title="Không có lịch trong tháng này" />}
      </div>
    </>
  );
}
