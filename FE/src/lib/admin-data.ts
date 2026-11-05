import type { Role } from "./data";
export type DemoUser = {
  id: string;
  name: string;
  email: string;
  role: Role;
  joined: string;
  locked: boolean;
};
export type Review = {
  id: string;
  name: string;
  address: string;
  owner: string;
  sport: string;
  status: "Chờ duyệt" | "Đã duyệt" | "Cần bổ sung";
  comment: string;
};
export const demoUsers: DemoUser[] = [
  {
    id: "u1",
    name: "Nguyễn Văn An",
    email: "an.nguyen@example.com",
    role: "user",
    joined: "2026-10-01",
    locked: false,
  },
  {
    id: "u2",
    name: "Trần Minh Đức",
    email: "duc.tran@example.com",
    role: "owner",
    joined: "2026-09-12",
    locked: false,
  },
  {
    id: "u3",
    name: "Lê Hoàng Nam",
    email: "nam.le@example.com",
    role: "user",
    joined: "2026-08-23",
    locked: true,
  },
  {
    id: "u4",
    name: "Quản trị viên",
    email: "admin@example.com",
    role: "admin",
    joined: "2026-01-01",
    locked: false,
  },
];
export const demoReviews: Review[] = [
  {
    id: "r1",
    name: "The Yard Pickleball Club",
    address: "Mễ Trì, Nam Từ Liêm, Hà Nội",
    owner: "Trần Minh Đức",
    sport: "Pickleball · 04 sân có mái che",
    status: "Chờ duyệt",
    comment: "",
  },
  {
    id: "r2",
    name: "Sân bóng Hoàng Gia",
    address: "Thanh Xuân, Hà Nội",
    owner: "Nguyễn Hoàng",
    sport: "Bóng đá · 03 sân",
    status: "Chờ duyệt",
    comment: "",
  },
  {
    id: "r3",
    name: "CLB Cầu lông Ánh Dương",
    address: "Hải Châu, Đà Nẵng",
    owner: "Lê Minh",
    sport: "Cầu lông · 06 sân trong nhà",
    status: "Chờ duyệt",
    comment: "",
  },
  {
    id: "r4",
    name: "Tennis Riverside",
    address: "Thủ Đức, TP.HCM",
    owner: "Phạm Hà",
    sport: "Tennis · 02 sân",
    status: "Chờ duyệt",
    comment: "",
  },
  {
    id: "r5",
    name: "Sân bóng An Phú",
    address: "Bình Thạnh, TP.HCM",
    owner: "Vũ Anh",
    sport: "Bóng đá · 02 sân",
    status: "Chờ duyệt",
    comment: "",
  },
];
