import { Confirmation } from "@/components/booking";
export const metadata = { title: "Xác nhận đặt sân" };
export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ id?: string }>;
}) {
  return <Confirmation id={(await searchParams).id} />;
}
