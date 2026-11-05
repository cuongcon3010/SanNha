import { Auth } from "@/components/auth";
export const metadata = { title: "Khôi phục mật khẩu" };
export default function Page() {
  return <Auth mode="reset" />;
}
