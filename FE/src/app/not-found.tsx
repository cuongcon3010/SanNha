import Link from "next/link";
import { Logo } from "@/components/ui";

export default function NotFound() {
  return (
    <main id="main-content" className="not-found">
      <Logo />
      <p className="eyebrow">404 / NGOÀI ĐƯỜNG BIÊN</p>
      <h1>Hình như bạn đã đi lạc sân.</h1>
      <p>Trang hoặc sân này không có trong bản demo.</p>
      <Link className="button" href="/">
        Về sân nhà ↗
      </Link>
    </main>
  );
}
