import { DashboardPage } from "@/components/dashboard/page";
export const metadata = { title: "Không gian chủ sân" };
export default async function Page({
  params,
}: {
  params: Promise<{ section?: string[] }>;
}) {
  return <DashboardPage role="owner" sections={(await params).section} />;
}
