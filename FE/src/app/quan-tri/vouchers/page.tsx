import { Header } from "@/components/header";
import { VoucherAdmin } from "@/components/voucher-admin";

export const metadata = { title: "Quản trị voucher" };
export default function Page() {
  return (
    <>
      <Header />
      <main id="main-content" className="container section guide">
        <VoucherAdmin />
      </main>
    </>
  );
}
