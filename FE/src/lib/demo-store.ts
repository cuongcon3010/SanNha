"use client";

import { useSyncExternalStore } from "react";
import {
  demoUsers,
  demoReviews,
  type DemoUser,
  type Review,
} from "./admin-data";
import {
  initialBookings,
  initialRates,
  venues,
  type Booking,
  type CourtRates,
} from "./data";

export type DemoProfile = {
  name: string;
  email: string;
  phone: string;
  area: string;
};
type DemoState = {
  bookings: Booking[];
  favorites: string[];
  profile: DemoProfile;
  rates: CourtRates;
  users: DemoUser[];
  reviews: Review[];
  audit: string[];
  storageError: boolean;
};
const initial: DemoState = {
  bookings: initialBookings,
  favorites: ["thanh-dong", "the-yard"],
  profile: {
    name: "Nguyễn Văn An",
    email: "an.nguyen@example.com",
    phone: "0901234567",
    area: "Cầu Giấy, Hà Nội",
  },
  storageError: false,
  rates: initialRates,
  users: demoUsers,
  reviews: demoReviews,
  audit: [
    "The Yard Club gửi hồ sơ sân mới",
    "Thành Đồng cập nhật bảng giá",
    "Người chơi gửi báo cáo thông tin sân",
  ],
};
let state = initial;
let loaded = false;
const listeners = new Set<() => void>();
const key = "san-nha-demo-v3";
function valid(value: unknown): value is DemoState {
  if (!value || typeof value !== "object") return false;
  const s = value as DemoState;
  return (
    Array.isArray(s.bookings) &&
    s.bookings.every(
      (b) =>
        b &&
        typeof b.id === "string" &&
        venues.some((v) => v.id === b.venueId && v.courts.includes(b.court)) &&
        /^\d{4}-\d{2}-\d{2}$/.test(b.date) &&
        Number.isInteger(b.hour) &&
        b.hour >= 16 &&
        b.hour <= 21 &&
        typeof b.name === "string" &&
        typeof b.phone === "string" &&
        typeof b.playerBooking === "boolean" &&
        Number.isFinite(b.total) &&
        b.total >= 0 &&
        ["Đã xác nhận", "Chờ cọc", "Hoàn thành", "Đã hủy", "Khóa giờ"].includes(
          b.status,
        ),
    ) &&
    Array.isArray(s.favorites) &&
    s.favorites.every((id) => venues.some((v) => v.id === id)) &&
    !!s.profile &&
    [s.profile.name, s.profile.email, s.profile.phone, s.profile.area].every(
      (v) => typeof v === "string",
    ) &&
    !!s.rates &&
    Object.keys(initialRates).every(
      (c) =>
        s.rates[c] &&
        Number.isSafeInteger(s.rates[c].day) &&
        s.rates[c].day > 0 &&
        Number.isSafeInteger(s.rates[c].evening) &&
        s.rates[c].evening > 0 &&
        typeof s.rates[c].maintenance === "boolean",
    ) &&
    Array.isArray(s.users) &&
    s.users.every(
      (u) =>
        u &&
        [u.id, u.name, u.email, u.joined].every((v) => typeof v === "string") &&
        ["user", "owner", "admin"].includes(u.role) &&
        typeof u.locked === "boolean",
    ) &&
    Array.isArray(s.reviews) &&
    s.reviews.every(
      (r) =>
        r &&
        [r.id, r.name, r.address, r.owner, r.sport, r.comment].every(
          (v) => typeof v === "string",
        ) &&
        ["Chờ duyệt", "Đã duyệt", "Cần bổ sung"].includes(r.status),
    ) &&
    Array.isArray(s.audit) &&
    s.audit.every((v) => typeof v === "string")
  );
}
function subscribe(listener: () => void) {
  listeners.add(listener);
  if (!loaded) {
    loaded = true;
    try {
      const saved = localStorage.getItem(key);
      if (saved) {
        const parsed: unknown = JSON.parse(saved);
        if (valid(parsed)) state = { ...parsed, storageError: false };
        else state = { ...initial, storageError: true };
      }
    } catch {
      state = { ...initial, storageError: true };
    }
    listeners.forEach((fn) => fn());
  }
  return () => {
    listeners.delete(listener);
  };
}
function update(next: DemoState) {
  state = next;
  try {
    localStorage.setItem(key, JSON.stringify(next));
  } catch {
    state = { ...next, storageError: true };
  }
  listeners.forEach((fn) => fn());
}
// shortcut: Dữ liệu dùng chung trong một trình duyệt, thay bằng API và phân quyền server trước khi vận hành thật.
export function useDemo() {
  const snapshot = useSyncExternalStore(
    subscribe,
    () => state,
    () => initial,
  );
  return {
    ...snapshot,
    addBooking: (booking: Booking) =>
      update({ ...state, bookings: [...state.bookings, booking] }),
    setStatus: (id: string, status: Booking["status"]) =>
      update({
        ...state,
        bookings: state.bookings.map((b) =>
          b.id === id ? { ...b, status } : b,
        ),
      }),
    saveProfile: (profile: DemoProfile) => update({ ...state, profile }),
    saveRates: (rates: CourtRates) => update({ ...state, rates }),
    toggleUser: (id: string) => {
      const user = state.users.find((u) => u.id === id);
      if (!user || user.role === "admin") return;
      update({
        ...state,
        users: state.users.map((u) =>
          u.id === id ? { ...u, locked: !u.locked } : u,
        ),
        audit: [
          `${user.locked ? "Mở khóa" : "Tạm khóa"} tài khoản ${user.name} (demo)`,
          ...state.audit,
        ].slice(0, 20),
      });
    },
    reviewVenue: (id: string, status: Review["status"], comment: string) => {
      const review = state.reviews.find((r) => r.id === id);
      if (!review) return;
      update({
        ...state,
        reviews: state.reviews.map((r) =>
          r.id === id ? { ...r, status, comment } : r,
        ),
        audit: [`${status}: ${review.name} (demo)`, ...state.audit].slice(
          0,
          20,
        ),
      });
    },
    toggleFavorite: (id: string) =>
      update({
        ...state,
        favorites: state.favorites.includes(id)
          ? state.favorites.filter((f) => f !== id)
          : [...state.favorites, id],
      }),
  };
}
