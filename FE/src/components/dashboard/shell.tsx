"use client";

import Link from "next/link";
import { useState, useEffect, useRef, useId, type ReactNode } from "react";
import {
  CalendarDays,
  UserRound,
  Heart,
  Building2,
  ChartNoAxesCombined,
  LayoutDashboard,
  UsersRound,
  ShieldCheck,
  LogOut,
  ArrowLeft,
  Menu,
  X,
  ChevronDown,
  ArrowUpRight,
} from "lucide-react";
import { navigation } from "@/lib/navigation";
import { roleLabels, type Role } from "@/lib/data";
import { useDemo } from "@/lib/demo-store";
import { Logo, DemoNotice } from "../ui";

const icons = {
  calendar: CalendarDays,
  user: UserRound,
  heart: Heart,
  venue: Building2,
  chart: ChartNoAxesCombined,
  overview: LayoutDashboard,
  users: UsersRound,
  shield: ShieldCheck,
};

export function DashboardShell({
  role,
  section,
  children,
}: {
  role: Role;
  section: string;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const { storageError } = useDemo();
  const base = role === "owner" ? "/chu-san" : `/${role}`;
  const current = base + (section ? `/${section}` : "");
  return (
    <div className="dashboard-layout">
      <aside className={`sidebar ${open ? "open" : ""}`}>
        <div className="sidebar-brand">
          <Logo light />
          <button
            className="icon-button mobile-only"
            onClick={() => setOpen(false)}
            aria-label="Đóng menu"
          >
            <X size={20} />
          </button>
        </div>
        <span className="eyebrow sidebar-role">
          KHÔNG GIAN {roleLabels[role].toUpperCase()}
        </span>
        <nav aria-label={`Menu ${roleLabels[role]}`}>
          {navigation[role].map((item) => {
            const Icon = icons[item.icon as keyof typeof icons];
            return (
              <Link
                key={item.path}
                href={item.path}
                className={current === item.path ? "active" : ""}
                aria-current={current === item.path ? "page" : undefined}
                onClick={() => setOpen(false)}
              >
                <Icon size={18} />
                {item.label}
                {current === item.path && <ArrowUpRight size={15} />}
              </Link>
            );
          })}
        </nav>
        <div className="sidebar-bottom">
          <Link href="/">
            <ArrowLeft size={15} />
            Về trang chủ
          </Link>
          <div className="sidebar-note dotted-dark">
            <span className="status-dot" />
            <strong>Cùng nhau ra sân.</strong>
            <p>Một kèo hay. Một ngày tốt.</p>
          </div>
          <Link href="/dang-nhap" className="sidebar-logout">
            <LogOut size={17} />
            Thoát không gian demo
            <ArrowUpRight size={15} />
          </Link>
        </div>
      </aside>
      {open && (
        <button
          className="sidebar-backdrop"
          onClick={() => setOpen(false)}
          aria-label="Đóng điều hướng"
        />
      )}
      <div className="dashboard-body">
        <header className="dashboard-topbar">
          <div>
            <button
              className="icon-button dashboard-menu"
              onClick={() => setOpen(true)}
              aria-label="Mở menu quản lý"
              aria-expanded={open}
            >
              <Menu size={22} />
            </button>
            <span>Không gian {roleLabels[role].toLowerCase()}</span>
          </div>
          <div>
            <DemoNotice compact />
            <details className="account-menu">
              <summary>
                <span className="avatar">
                  {role === "user" ? "AN" : role === "owner" ? "TĐ" : "AD"}
                </span>
                <strong>
                  {role === "user"
                    ? "An Nguyễn"
                    : role === "owner"
                      ? "Thành Đồng"
                      : "Quản trị viên"}
                </strong>
                <ChevronDown size={14} />
              </summary>
              <nav aria-label="Đổi vai trò demo">
                <small>KHÁM PHÁ VAI TRÒ DEMO</small>
                <Link href="/user">Người chơi</Link>
                <Link href="/chu-san">Chủ sân</Link>
                <Link href="/admin">Admin</Link>
                <Link href="/dang-nhap">Về trang đăng nhập</Link>
              </nav>
            </details>
          </div>
        </header>
        <main id="main-content" className="dashboard-content">
          {storageError && (
            <p role="status" className="form-error">
              Không thể đọc hoặc lưu dữ liệu trình duyệt. Các thay đổi mới chỉ
              tồn tại trong phiên này.
            </p>
          )}
          {children}
          <p className="dashboard-footnote">
            <ShieldCheck size={14} />
            Không gian demo công khai · Không có phân quyền hay giao dịch thực
            tế.
          </p>
        </main>
      </div>
    </div>
  );
}

export function PageHeading({
  eyebrow,
  title,
  description,
  action,
}: {
  eyebrow: string;
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="dashboard-heading">
      <div>
        <span className="eyebrow">{eyebrow}</span>
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      {action}
    </div>
  );
}
export function Stats({
  items,
}: {
  items: { label: string; value: string; caption: string }[];
}) {
  return (
    <div className="stats-grid">
      {items.map((s, i) => (
        <section
          className={`stat-card ${i === 0 ? "featured" : ""}`}
          key={s.label}
        >
          <div>
            <span>{s.label}</span>
            <ArrowUpRight size={16} />
          </div>
          <strong>{s.value}</strong>
          <small>{s.caption}</small>
        </section>
      ))}
    </div>
  );
}
export function Modal({
  title,
  children,
  onClose,
}: {
  title: string;
  children: ReactNode;
  onClose: () => void;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  const id = useId();
  useEffect(() => {
    ref.current?.showModal();
  }, []);
  return (
    <dialog
      ref={ref}
      className="modal"
      aria-labelledby={id}
      onCancel={onClose}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="modal-content">
        <div className="modal-heading">
          <h2 id={id}>{title}</h2>
          <button
            className="icon-button"
            onClick={onClose}
            aria-label="Đóng hộp thoại"
          >
            <X size={20} />
          </button>
        </div>
        {children}
      </div>
    </dialog>
  );
}
