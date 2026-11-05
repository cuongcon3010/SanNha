import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";

const inter = localFont({
  src: [
    {
      path: "../../public/assets/inter-variable.woff2",
      weight: "100 900",
      style: "normal",
    },
  ],
  variable: "--font-inter",
  display: "swap",
});

export const metadata: Metadata = {
  title: {
    default: "Sân Nhà — Kèo đã có. Sân để chúng tôi.",
    template: "%s | Sân Nhà",
  },
  description:
    "Tìm sân đúng gu, chọn giờ đúng ý. Giao diện đặt sân thể thao và quản lý sân — bản demo Sân Nhà.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="vi">
      <body className={inter.variable}>
        <a className="skip-link" href="#main-content">
          Đến nội dung chính
        </a>
        {children}
      </body>
    </html>
  );
}
