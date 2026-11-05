import { DashboardPage } from "@/components/dashboard/page";
export const metadata = { title: "Không gian người chơi" };
export default async function Page({
  params,
}: {
  params: Promise<{ section?: string[] }>;
}) {
  return <DashboardPage role="user" sections={(await params).section} />;
}
