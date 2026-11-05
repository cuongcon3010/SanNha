import { OAuthAccount } from "@/components/oauth";
import { Header } from "@/components/header";

export const metadata = { title: "Tài khoản đã xác thực" };
export default function Page() {
  return (
    <>
      <Header />
      <main id="main-content" className="container section guide">
        <OAuthAccount />
      </main>
    </>
  );
}
