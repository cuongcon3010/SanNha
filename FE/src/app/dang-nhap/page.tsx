import { Auth } from "@/components/auth";
export const metadata = { title: "Đăng nhập" };
export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ oauth_error?: string }>;
}) {
  return <Auth mode="login" oauthError={(await searchParams).oauth_error} />;
}
