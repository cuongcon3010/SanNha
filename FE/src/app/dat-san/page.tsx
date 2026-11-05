import { Header } from "@/components/header";
import { LiveBookingForm } from "@/components/live-booking";

export const metadata = { title: "Đặt sân & thanh toán VietQR" };
export default function Page() {
  return (
    <>
      <Header />
      <main id="main-content" className="container section guide">
        <LiveBookingForm />
      </main>
    </>
  );
}
