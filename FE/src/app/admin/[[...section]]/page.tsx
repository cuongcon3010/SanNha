import { DashboardPage } from "@/components/dashboard/page";
export const metadata = { title: "Quản trị Sân Nhà" };
export default async function Page({
  params,
}: {
  params: Promise<{ section?: string[] }>;
}) {
  return <DashboardPage role="admin" sections={(await params).section} />;
}
