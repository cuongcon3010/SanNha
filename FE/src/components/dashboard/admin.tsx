"use client";

import { useState } from "react";
import Link from "next/link";
import {
  ArrowUpRight,
  Search,
  ShieldCheck,
  MapPin,
  Check,
  LockKeyhole,
} from "lucide-react";
import { dateLabel, normalize, roleLabels } from "@/lib/data";
import type { DemoUser } from "@/lib/admin-data";
import { useDemo } from "@/lib/demo-store";
import { Badge, Empty, SuccessMessage } from "../ui";
import { Modal, PageHeading, Stats } from "./shell";

export function AdminDashboard({ section }: { section: string }) {
  if (section === "nguoi-dung") return <Users />;
  if (section === "kiem-duyet") return <Moderation />;
  return <Overview />;
}
function Overview() {
  const { users, reviews, bookings, audit } = useDemo();
  const pending = reviews.filter((r) => r.status === "Chờ duyệt").length;
  return (
    <>
      <PageHeading
        eyebrow="QUẢN TRỊ / TỔNG QUAN"
        title="Toàn cảnh Sân Nhà."
        description="Theo dõi nền tảng, kiểm duyệt sân và xử lý các vấn đề cần ưu tiên."
      />
      <Stats
        items={[
          {
            label: "Người dùng mẫu",
            value: String(users.length).padStart(2, "0"),
            caption: "Tài khoản trong bộ dữ liệu demo",
          },
          {
            label: "Hồ sơ sân",
            value: String(reviews.length).padStart(2, "0"),
            caption: `${reviews.filter((r) => r.status === "Đã duyệt").length} hồ sơ đã duyệt`,
          },
          {
            label: "Lượt đặt mẫu",
            value: String(
              bookings.filter((b) => b.status !== "Khóa giờ").length,
            ).padStart(2, "0"),
            caption: "Bao gồm lịch tạo trong trình duyệt",
          },
          {
            label: "Cần xử lý",
            value: String(pending).padStart(2, "0"),
            caption: "Hồ sơ đang chờ kiểm duyệt",
          },
        ]}
      />
      <div className="overview-grid">
        <section className="panel admin-chart">
          <div className="content-heading">
            <h2>Lượt đặt trong 7 ngày</h2>
            <Badge>Thống kê minh họa</Badge>
          </div>
          <div className="vertical-chart">
            {[92, 125, 109, 156, 133, 186, 171].map((v, i) => (
              <div key={i}>
                <span>{v}</span>
                <i
                  style={{ height: `${v}px` }}
                  className={i === 5 ? "highlight" : ""}
                />
                <small>
                  {
                    [
                      "31/10",
                      "01/11",
                      "02/11",
                      "03/11",
                      "04/11",
                      "05/11",
                      "06/11",
                    ][i]
                  }
                </small>
              </div>
            ))}
          </div>
        </section>
        <section className="panel priority-panel dotted-dark">
          <span className="eyebrow">ƯU TIÊN HÔM NAY</span>
          <h2>Hộp thư xử lý</h2>
          <p>
            <strong>{String(pending).padStart(2, "0")}</strong> sân chờ kiểm
            duyệt
          </p>
          <p>
            <strong>
              {String(users.filter((u) => u.locked).length).padStart(2, "0")}
            </strong>{" "}
            tài khoản đang tạm khóa
          </p>
          <Link className="button full-width" href="/admin/kiem-duyet">
            Mở kiểm duyệt <ArrowUpRight size={16} />
          </Link>
        </section>
      </div>
      <div className="content-heading">
        <h2>Hoạt động gần đây</h2>
        <span className="muted">Nhật ký demo trong trình duyệt</span>
      </div>
      <div className="table-panel">
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Hoạt động</th>
                <th>Nguồn</th>
                <th>Trạng thái</th>
              </tr>
            </thead>
            <tbody>
              {audit.slice(0, 6).map((a, i) => (
                <tr key={i}>
                  <td>
                    <strong>{a}</strong>
                  </td>
                  <td>Phiên demo Sân Nhà</td>
                  <td>
                    <Badge>Đã ghi nhận</Badge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

function Users() {
  const { users, toggleUser } = useDemo();
  const [query, setQuery] = useState("");
  const [role, setRole] = useState("");
  const [status, setStatus] = useState("");
  const [selected, setSelected] = useState<DemoUser | null>(null);
  const [notice, setNotice] = useState("");
  const filtered = users.filter(
    (u) =>
      normalize(u.name + " " + u.email).includes(normalize(query)) &&
      (!role || u.role === role) &&
      (!status || (status === "locked" ? u.locked : !u.locked)),
  );
  function toggle() {
    if (!selected) return;
    if (
      window.confirm(
        `${selected.locked ? "Mở khóa" : "Tạm khóa"} tài khoản demo ${selected.name}? Hành động chỉ cập nhật dữ liệu mẫu, không áp dụng quyền truy cập thật.`,
      )
    ) {
      toggleUser(selected.id);
      setNotice(
        `Đã ${selected.locked ? "mở khóa" : "tạm khóa"} tài khoản demo ${selected.name}.`,
      );
      setSelected(null);
    }
  }
  return (
    <>
      <PageHeading
        eyebrow="QUẢN TRỊ / NGƯỜI DÙNG"
        title="Đúng người. Đúng quyền."
        description="Quản lý tài khoản mẫu, vai trò và trạng thái. Không có tài khoản thật."
      />
      <Stats
        items={[
          {
            label: "Tổng tài khoản",
            value: String(users.length).padStart(2, "0"),
            caption: "Trong bộ dữ liệu mẫu",
          },
          {
            label: "Người chơi",
            value: String(
              users.filter((u) => u.role === "user").length,
            ).padStart(2, "0"),
            caption: "Đang tham gia cộng đồng",
          },
          {
            label: "Chủ sân",
            value: String(
              users.filter((u) => u.role === "owner").length,
            ).padStart(2, "0"),
            caption: "Đối tác vận hành sân",
          },
          {
            label: "Tạm khóa",
            value: String(users.filter((u) => u.locked).length).padStart(
              2,
              "0",
            ),
            caption: "Trạng thái minh họa",
          },
        ]}
      />
      <div className="panel user-filters">
        <label className="search-input">
          <Search size={17} />
          <input
            aria-label="Tìm người dùng"
            placeholder="Tìm theo tên hoặc email…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </label>
        <label className="sr-only" htmlFor="role-filter">
          Vai trò
        </label>
        <select
          id="role-filter"
          value={role}
          onChange={(e) => setRole(e.target.value)}
        >
          <option value="">Vai trò: Tất cả</option>
          <option value="user">Người chơi</option>
          <option value="owner">Chủ sân</option>
          <option value="admin">Admin</option>
        </select>
        <label className="sr-only" htmlFor="status-filter">
          Trạng thái
        </label>
        <select
          id="status-filter"
          value={status}
          onChange={(e) => setStatus(e.target.value)}
        >
          <option value="">Trạng thái: Tất cả</option>
          <option value="active">Hoạt động</option>
          <option value="locked">Tạm khóa</option>
        </select>
      </div>
      {notice && <SuccessMessage>{notice}</SuccessMessage>}
      <div className="table-panel">
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Tài khoản</th>
                <th>Vai trò</th>
                <th>Ngày tham gia</th>
                <th>Trạng thái</th>
                <th>Thao tác</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((u) => (
                <tr key={u.id}>
                  <td>
                    <div className="user-cell">
                      <span className="avatar">
                        {u.name
                          .split(" ")
                          .slice(-2)
                          .map((n) => n[0])
                          .join("")}
                      </span>
                      <div>
                        <strong>{u.name}</strong>
                        <small>{u.email}</small>
                      </div>
                    </div>
                  </td>
                  <td>{roleLabels[u.role]}</td>
                  <td>{dateLabel(u.joined)}</td>
                  <td>
                    <Badge tone={u.locked ? "red" : "neutral"}>
                      {u.locked ? "Tạm khóa" : "Hoạt động"}
                    </Badge>
                  </td>
                  <td>
                    <button
                      className="table-link"
                      onClick={() => setSelected(u)}
                    >
                      Chi tiết <ArrowUpRight size={14} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {filtered.length === 0 && (
          <Empty title="Không tìm thấy tài khoản phù hợp" />
        )}
        <div className="table-footer">
          {filtered.length} tài khoản phù hợp
          <span>Không hiển thị mật khẩu</span>
        </div>
      </div>
      <div className="warning-note dark">
        <ShieldCheck size={24} />
        <div>
          <strong>
            Thay đổi quyền và khóa tài khoản cần được xử lý phía server khi vận
            hành thật.
          </strong>
          <p>
            Bản demo chỉ thay đổi trạng thái mẫu và ghi nhật ký, không ngăn truy
            cập vào trang.
          </p>
        </div>
      </div>
      {selected && (
        <Modal
          title="Thông tin tài khoản mẫu"
          onClose={() => setSelected(null)}
        >
          <div className="booking-modal-info">
            <Badge>{roleLabels[selected.role]}</Badge>
            <h3>{selected.name}</h3>
            <p>{selected.email}</p>
            <p>Tham gia: {dateLabel(selected.joined)}</p>
            <p>Trạng thái: {selected.locked ? "Tạm khóa" : "Hoạt động"}</p>
          </div>
          {selected.role === "admin" ? (
            <p className="muted">
              Không cho phép tự khóa tài khoản quản trị viên trong demo.
            </p>
          ) : (
            <button
              className={`button ${selected.locked ? "" : "button-danger"}`}
              onClick={toggle}
            >
              <LockKeyhole size={16} />
              {selected.locked
                ? "Mở khóa tài khoản demo"
                : "Tạm khóa tài khoản demo"}
            </button>
          )}
        </Modal>
      )}
    </>
  );
}

function Moderation() {
  const { reviews, reviewVenue } = useDemo();
  const [id, setId] = useState(reviews[0]?.id);
  const [filter, setFilter] = useState("Tất cả");
  const [checked, setChecked] = useState(false);
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const rows = reviews.filter(
    (r) => filter === "Tất cả" || r.status === filter,
  );
  const selected = rows.find((r) => r.id === id) || rows[0];
  function review(approve: boolean) {
    if (!selected) return;
    if (approve && !checked) {
      setError("Vui lòng xác nhận đã kiểm tra hồ sơ sở hữu / ủy quyền.");
      return;
    }
    if (!approve && comment.trim().length < 10) {
      setError("Vui lòng nhập lý do cần bổ sung (ít nhất 10 ký tự).");
      return;
    }
    if (
      !window.confirm(
        `${approve ? "Phê duyệt" : "Yêu cầu bổ sung"} hồ sơ demo ${selected.name}?`,
      )
    )
      return;
    reviewVenue(
      selected.id,
      approve ? "Đã duyệt" : "Cần bổ sung",
      comment.trim(),
    );
    setNotice(
      approve
        ? "Đã duyệt hồ sơ mẫu. Không đăng sân thực tế."
        : "Đã ghi nhận yêu cầu bổ sung trong demo. Không gửi thông báo thật.",
    );
    setError("");
  }
  return (
    <>
      <PageHeading
        eyebrow="QUẢN TRỊ / KIỂM DUYỆT"
        title="Sân chất lượng. Cuộc chơi an tâm."
        description="Xem hồ sơ cụm sân trước khi cho phép xuất hiện công khai."
      />
      {notice && <SuccessMessage>{notice}</SuccessMessage>}
      <div className="moderation-grid">
        <section className="panel review-queue">
          <div className="content-heading">
            <h2>Hồ sơ sân</h2>
            <Badge tone="yellow">
              {reviews.filter((r) => r.status === "Chờ duyệt").length} chờ
            </Badge>
          </div>
          <label className="sr-only" htmlFor="review-filter">
            Lọc hồ sơ
          </label>
          <select
            id="review-filter"
            value={filter}
            onChange={(e) => {
              setFilter(e.target.value);
              setChecked(false);
              setComment("");
              setError("");
            }}
          >
            <option>Tất cả</option>
            <option>Chờ duyệt</option>
            <option>Đã duyệt</option>
            <option>Cần bổ sung</option>
          </select>
          {rows.map((r) => (
            <button
              key={r.id}
              className={`queue-item ${selected?.id === r.id ? "active" : ""}`}
              onClick={() => {
                setId(r.id);
                setChecked(false);
                setComment(r.comment);
                setError("");
                setNotice("");
              }}
            >
              <strong>{r.name}</strong>
              <span>
                <MapPin size={12} />
                {r.address}
              </span>
              <small>{r.status}</small>
            </button>
          ))}
          {rows.length === 0 && <Empty title="Không có hồ sơ trong mục này" />}
        </section>
        <section className="panel review-detail">
          {selected ? (
            <>
              <Badge
                tone={selected.status === "Cần bổ sung" ? "red" : "yellow"}
              >
                {selected.status}
              </Badge>
              <h2>{selected.name}</h2>
              <p className="muted">
                <MapPin size={14} />
                {selected.address}
              </p>
              <dl>
                <div>
                  <dt>Chủ sân</dt>
                  <dd>{selected.owner}</dd>
                </div>
                <div>
                  <dt>Loại hình</dt>
                  <dd>{selected.sport}</dd>
                </div>
                <div>
                  <dt>Giờ mở cửa</dt>
                  <dd>05:00 – 23:00</dd>
                </div>
                <div>
                  <dt>Giá niêm yết mẫu</dt>
                  <dd>150.000đ – 250.000đ / giờ</dd>
                </div>
              </dl>
              <div className="review-checklist">
                <p>
                  <Check size={16} />
                  Thông tin liên hệ đã xác minh (mẫu)
                </p>
                <p>
                  <Check size={16} />
                  Đã nộp thông tin ảnh và địa chỉ (mẫu)
                </p>
                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={(e) => {
                      setChecked(e.target.checked);
                      setError("");
                    }}
                  />
                  Đã kiểm tra hồ sơ sở hữu / ủy quyền quản lý
                </label>
              </div>
              <label className="field">
                Nhận xét cho chủ sân
                <textarea
                  rows={3}
                  maxLength={1000}
                  value={comment}
                  onChange={(e) => setComment(e.target.value)}
                  placeholder="Nhập lý do yêu cầu bổ sung…"
                />
              </label>
              {selected.comment && (
                <p className="muted">Nhận xét đã lưu: {selected.comment}</p>
              )}
              {error && (
                <p className="form-error" role="alert">
                  {error}
                </p>
              )}
              <div className="modal-actions">
                <button
                  className="button"
                  onClick={() => review(true)}
                  disabled={selected.status === "Đã duyệt"}
                >
                  <ShieldCheck size={16} />
                  Phê duyệt sân
                </button>
                <button
                  className="button button-danger"
                  onClick={() => review(false)}
                >
                  Yêu cầu bổ sung
                </button>
              </div>
              <p className="photo-note">
                Hồ sơ, xác minh và thao tác đều là dữ liệu demo.
              </p>
            </>
          ) : (
            <Empty title="Không có hồ sơ được chọn" />
          )}
        </section>
      </div>
    </>
  );
}
