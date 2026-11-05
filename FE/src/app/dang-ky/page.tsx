import { Auth } from "@/components/auth";
export const metadata = { title: "Đăng ký" };
export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ role?: string }>;
}) {
  const { role } = await searchParams;
  return <Auth mode="register" owner={role === "owner"} />;
}
