import { notFound } from "next/navigation";
import { BookingPage } from "@/components/booking";
import { venues } from "@/lib/data";

export function generateStaticParams() {
  return venues.map((v) => ({ slug: v.id }));
}
export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  return {
    title: venues.find((v) => v.id === slug)?.name || "Không tìm thấy sân",
  };
}
export default async function Page({
  params,
  searchParams,
}: {
  params: Promise<{ slug: string }>;
  searchParams: Promise<{ date?: string; period?: string }>;
}) {
  const { slug } = await params;
  const query = await searchParams;
  const venue = venues.find((v) => v.id === slug);
  if (!venue) notFound();
  return (
    <BookingPage
      venue={venue}
      initialDate={query.date}
      allHours={query.period === "all"}
    />
  );
}
