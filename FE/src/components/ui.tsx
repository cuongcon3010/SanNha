import Link from "next/link";
import { ArrowUpRight, Check, CircleHelp, ShieldCheck } from "lucide-react";
import type { ReactNode } from "react";

export function Logo({ light = false }: { light?: boolean }) {
  return (
    <Link
      href="/"
      className={`logo ${light ? "logo-light" : ""}`}
      aria-label="Sân Nhà — Trang chủ"
    >
      <span className="logo-mark">
        S<span className="logo-dot" />
      </span>
      <span>
        SÂN NHÀ<span className="logo-period">.</span>
      </span>
    </Link>
  );
}
export function Arrow() {
  return <ArrowUpRight size={17} strokeWidth={1.8} aria-hidden="true" />;
}
export function Badge({
  children,
  tone = "neutral",
}: {
  children: ReactNode;
  tone?: "neutral" | "yellow" | "red";
}) {
  return (
    <span className={`badge badge-${tone}`}>
      <span className="status-dot" />
      {children}
    </span>
  );
}
export function DemoNotice({ compact = false }: { compact?: boolean }) {
  return (
    <div className={`demo-notice ${compact ? "compact" : ""}`}>
      <ShieldCheck size={16} aria-hidden="true" />
      <span>
        {compact
          ? "Dữ liệu demo"
          : "Chế độ demo · Không thanh toán, không giữ sân thực tế. Không nhập thông tin nhạy cảm."}
      </span>
    </div>
  );
}
export function Empty({
  title = "Chưa có dữ liệu",
  description = "Thử thay đổi bộ lọc hoặc quay lại sau.",
}: {
  title?: string;
  description?: string;
}) {
  return (
    <div className="empty-state">
      <CircleHelp size={32} />
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  );
}
export function SuccessMessage({ children }: { children: ReactNode }) {
  return (
    <p className="success-message" role="status">
      <Check size={18} />
      {children}
    </p>
  );
}
export function Footer() {
  return (
    <footer className="site-footer container">
      <div>
        <Logo />
        <p>Cùng nhau ra sân. Cùng nhau hết mình.</p>
      </div>
      <div className="footer-right">
        <span>© 2026 Sân Nhà · Bản demo</span>
        <Link href="/huong-dan">
          Hướng dẫn & chính sách demo <Arrow />
        </Link>
      </div>
    </footer>
  );
}
