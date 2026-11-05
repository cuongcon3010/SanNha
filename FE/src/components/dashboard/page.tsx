import { notFound } from "next/navigation";
import type { Role } from "@/lib/data";
import { navigation } from "@/lib/navigation";
import { DashboardShell } from "./shell";
import { UserDashboard } from "./user";
import { OwnerDashboard } from "./owner";
import { AdminDashboard } from "./admin";

export function DashboardPage({
  role,
  sections,
}: {
  role: Role;
  sections?: string[];
}) {
  const section = sections?.join("/") || "";
  const base = role === "owner" ? "/chu-san" : `/${role}`;
  if (
    !navigation[role].some(
      (n) => n.path === base + (section ? `/${section}` : ""),
    )
  )
    notFound();
  return (
    <DashboardShell role={role} section={section}>
      {role === "user" ? (
        <UserDashboard section={section} />
      ) : role === "owner" ? (
        <OwnerDashboard section={section} />
      ) : (
        <AdminDashboard section={section} />
      )}
    </DashboardShell>
  );
}
