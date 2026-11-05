export const sports = [
  "Tất cả",
  "Bóng đá",
  "Cầu lông",
  "Pickleball",
  "Tennis",
  "Bóng rổ",
];
export const cities = ["Hà Nội", "Hồ Chí Minh", "Đà Nẵng"];
export const demoDate = "2026-11-06";
export type Role = "user" | "owner" | "admin";
export const roleLabels: Record<Role, string> = {
  user: "Người chơi",
  owner: "Chủ sân",
  admin: "Admin",
};
export type Venue = {
  id: string;
  name: string;
  sport: string;
  city: string;
  address: string;
  price: number;
  image: string;
  rating: string;
  courts: string[];
  amenities: string[];
};
export const venues: Venue[] = [
  {
    id: "thanh-dong",
    name: "Sân bóng Thành Đồng",
    sport: "Bóng đá",
    city: "Hà Nội",
    address: "Nam Trung Yên, Cầu Giấy, Hà Nội",
    price: 220000,
    image: "/assets/football.jpg",
    rating: "4.9",
    courts: ["Sân A", "Sân B", "Sân C"],
    amenities: ["Sân 7 người", "Đèn LED", "Bãi đỗ xe", "Phòng thay đồ"],
  },
  {
    id: "the-yard",
    name: "The Yard Pickleball Club",
    sport: "Pickleball",
    city: "Hà Nội",
    address: "Mễ Trì, Nam Từ Liêm, Hà Nội",
    price: 150000,
    image: "/assets/pickleball.jpg",
    rating: "4.9",
    courts: ["Sân 1", "Sân 2", "Sân 3", "Sân 4"],
    amenities: ["Có mái che", "Đèn LED", "Bãi đỗ xe", "Cho thuê vợt"],
  },
  {
    id: "ha-dong",
    name: "Sân cầu lông Hà Đông",
    sport: "Cầu lông",
    city: "Hà Nội",
    address: "Vạn Phúc, Hà Đông, Hà Nội",
    price: 100000,
    image: "/assets/badminton.jpg",
    rating: "4.8",
    courts: ["Sân 1", "Sân 2"],
    amenities: ["Sàn tiêu chuẩn", "Trong nhà", "Bãi đỗ xe", "Phòng thay đồ"],
  },
];
export const money = (value: number) =>
  new Intl.NumberFormat("vi-VN").format(value) + "đ";
export function dateLabel(date: string) {
  return new Date(date + "T12:00:00").toLocaleDateString("vi-VN", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}
export function today() {
  return new Intl.DateTimeFormat("sv-SE", {
    timeZone: "Asia/Ho_Chi_Minh",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}
export function normalize(value: string) {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/đ/g, "d")
    .replace(/Đ/g, "D")
    .toLowerCase()
    .trim();
}
export function isDate(value: string) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  const parsed = new Date(value + "T12:00:00Z");
  return (
    Number.isFinite(parsed.getTime()) &&
    parsed.toISOString().slice(0, 10) === value
  );
}
export function filterVenues(
  list: Venue[],
  sport: string,
  city: string,
  query: string,
) {
  return list.filter(
    (v) =>
      (sport === "Tất cả" || v.sport === sport) &&
      (!city || v.city === city) &&
      normalize(v.name + " " + v.address).includes(normalize(query)),
  );
}
export type BookingStatus =
  "Đã xác nhận" | "Chờ cọc" | "Hoàn thành" | "Đã hủy" | "Khóa giờ";
export type Booking = {
  id: string;
  venueId: string;
  court: string;
  date: string;
  hour: number;
  name: string;
  phone: string;
  total: number;
  status: BookingStatus;
  playerBooking: boolean;
};
export type CourtRate = { day: number; evening: number; maintenance: boolean };
export type CourtRates = Record<string, CourtRate>;
export const initialRates: CourtRates = {
  "Sân A": { day: 180000, evening: 220000, maintenance: false },
  "Sân B": { day: 180000, evening: 220000, maintenance: false },
  "Sân C": { day: 120000, evening: 180000, maintenance: false },
};
export function courtPrice(
  venue: Venue,
  court: string,
  hour: number,
  rates: CourtRates,
) {
  const rate = venue.id === "thanh-dong" ? rates[court] : undefined;
  return rate ? (hour < 17 ? rate.day : rate.evening) : venue.price;
}
export const hours = [16, 17, 18, 19, 20, 21];
export const timeLabel = (hour: number) => `${hour}:00 – ${hour + 1}:00`;
export function isAvailable(
  bookings: Booking[],
  venueId: string,
  court: string,
  date: string,
  hour: number,
) {
  return !bookings.some(
    (b) =>
      b.venueId === venueId &&
      b.court === court &&
      b.date === date &&
      b.hour === hour &&
      b.status !== "Đã hủy",
  );
}
export function bookingError(
  booking: Pick<
    Booking,
    "venueId" | "court" | "date" | "hour" | "name" | "phone"
  >,
  bookings: Booking[],
) {
  const venue = venues.find((v) => v.id === booking.venueId);
  if (
    !venue ||
    !venue.courts.includes(booking.court) ||
    !hours.includes(booking.hour)
  )
    return "Sân hoặc khung giờ không hợp lệ.";
  if (!isDate(booking.date) || booking.date < today())
    return "Vui lòng chọn ngày chơi từ hôm nay.";
  if (booking.name.trim().length < 2)
    return "Vui lòng nhập họ tên có ít nhất 2 ký tự.";
  if (!/^(?:0\d{9}|\+84\d{9})$/.test(booking.phone.replace(/\s/g, "")))
    return "Số điện thoại không hợp lệ (10 chữ số, bắt đầu bằng 0).";
  if (
    !isAvailable(
      bookings,
      booking.venueId,
      booking.court,
      booking.date,
      booking.hour,
    )
  )
    return "Khung giờ này đã có lịch. Vui lòng chọn giờ khác.";
  return "";
}
export const deposit = (total: number) => Math.round(total * 0.3);
export const initialBookings: Booking[] = [
  {
    id: "SN-0611-028",
    venueId: "thanh-dong",
    court: "Sân A",
    date: demoDate,
    hour: 19,
    name: "Nguyễn Văn An",
    phone: "0901234567",
    total: 220000,
    status: "Đã xác nhận",
    playerBooking: true,
  },
  {
    id: "SN-0811-031",
    venueId: "the-yard",
    court: "Sân 1",
    date: "2026-11-08",
    hour: 17,
    name: "Nguyễn Văn An",
    phone: "0901234567",
    total: 150000,
    status: "Chờ cọc",
    playerBooking: true,
  },
  {
    id: "SN-0111-019",
    venueId: "ha-dong",
    court: "Sân 1",
    date: "2026-11-01",
    hour: 18,
    name: "Nguyễn Văn An",
    phone: "0901234567",
    total: 100000,
    status: "Hoàn thành",
    playerBooking: true,
  },
  {
    id: "SN-0611-029",
    venueId: "thanh-dong",
    court: "Sân B",
    date: demoDate,
    hour: 16,
    name: "FC Cầu Giấy",
    phone: "0901234568",
    total: 220000,
    status: "Chờ cọc",
    playerBooking: false,
  },
  {
    id: "SN-0611-030",
    venueId: "thanh-dong",
    court: "Sân B",
    date: demoDate,
    hour: 18,
    name: "CLB Mễ Trì",
    phone: "0901234569",
    total: 220000,
    status: "Đã xác nhận",
    playerBooking: false,
  },
  {
    id: "BLOCK-1",
    venueId: "thanh-dong",
    court: "Sân C",
    date: demoDate,
    hour: 17,
    name: "Bảo trì",
    phone: "",
    total: 0,
    status: "Khóa giờ",
    playerBooking: false,
  },
];
