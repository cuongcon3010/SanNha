"use client";

import Link from "next/link";
import { useState } from "react";
import { Menu, X, CalendarDays } from "lucide-react";
import { Logo, Arrow } from "./ui";

export function Header() {
  const [open, setOpen] = useState(false);
  return (
    <header className="site-header">
      <div className="container header-inner">
        <Logo />
        <nav
          className={open ? "header-nav is-open" : "header-nav"}
          aria-label="Điều hướng chính"
        >
          <Link href="/#san-noi-bat" onClick={() => setOpen(false)}>
            Khám phá sân
          </Link>
          <Link href="/#cach-dat-san" onClick={() => setOpen(false)}>
            Cách đặt sân
          </Link>
          <Link href="/dang-ky?role=owner" onClick={() => setOpen(false)}>
            Dành cho chủ sân <span className="tiny-label">PRO</span>
          </Link>
        </nav>
        <div className="header-actions">
          <Link href="/user" className="my-bookings">
            <CalendarDays size={17} />
            Lịch của tôi
          </Link>
          <Link className="button button-dark" href="/dang-nhap">
            Đăng nhập <Arrow />
          </Link>
          <button
            className="icon-button menu-toggle"
            onClick={() => setOpen(!open)}
            aria-label={open ? "Đóng menu" : "Mở menu"}
            aria-expanded={open}
          >
            {open ? <X /> : <Menu />}
          </button>
        </div>
      </div>
    </header>
  );
}
