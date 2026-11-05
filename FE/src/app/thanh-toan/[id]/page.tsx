import { Header } from "@/components/header";
import { PaymentPage } from "@/components/payment";

export const metadata = { title: "Thanh toán cọc VietQR" };
export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return (
    <>
      <Header />
      <main id="main-content" className="confirmation-page dotted-light">
        <PaymentPage key={id} id={id} />
      </main>
    </>
  );
}
