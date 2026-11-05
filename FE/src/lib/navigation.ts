import type { Role } from "./data";

export const navigation: Record<
  Role,
  { path: string; label: string; icon: string }[]
> = {
  user: [
    { path: "/user", label: "Lịch đặt của tôi", icon: "calendar" },
    { path: "/user/ho-so", label: "Thông tin cá nhân", icon: "user" },
    { path: "/user/yeu-thich", label: "Sân yêu thích", icon: "heart" },
  ],
  owner: [
    { path: "/chu-san", label: "Lịch vận hành", icon: "calendar" },
    { path: "/chu-san/san", label: "Sân & bảng giá", icon: "venue" },
    { path: "/chu-san/doanh-thu", label: "Doanh thu", icon: "chart" },
  ],
  admin: [
    { path: "/admin", label: "Tổng quan", icon: "overview" },
    { path: "/admin/nguoi-dung", label: "Người dùng", icon: "users" },
    { path: "/admin/kiem-duyet", label: "Kiểm duyệt sân", icon: "shield" },
  ],
};
