import assert from "node:assert/strict";
import test from "node:test";
import {
  bookingError,
  courtPrice,
  deposit,
  filterVenues,
  initialBookings,
  initialRates,
  isAvailable,
  isDate,
  normalize,
  today,
  venues,
} from "../src/lib/data.ts";

test("Tìm kiếm không dấu và kết hợp môn, thành phố", () => {
  assert.equal(normalize("Cầu Giấy, Hà Nội"), "cau giay, ha noi");
  assert.deepEqual(
    filterVenues(venues, "Bóng đá", "Hà Nội", "cau giay").map((v) => v.id),
    ["thanh-dong"],
  );
  assert.equal(filterVenues(venues, "Tennis", "", " ").length, 0);
  assert.equal(filterVenues(venues, "Tất cả", "Đà Nẵng", "").length, 0);
  assert.equal(filterVenues(venues, "Tất cả", "", "").length, 3);
});
test("Không đặt trùng sân / ngày / giờ, lịch hủy được giải phóng", () => {
  assert.equal(
    isAvailable(initialBookings, "thanh-dong", "Sân A", "2026-11-06", 19),
    false,
  );
  assert.equal(
    isAvailable(initialBookings, "thanh-dong", "Sân B", "2026-11-06", 19),
    true,
  );
  assert.equal(
    isAvailable(initialBookings, "thanh-dong", "Sân C", "2026-11-06", 17),
    false,
  );
  assert.equal(
    isAvailable(
      initialBookings.map((b) => ({ ...b, status: "Đã hủy" as const })),
      "thanh-dong",
      "Sân A",
      "2026-11-06",
      19,
    ),
    true,
  );
});
test("Kiểm tra sân, ngày, tên, số điện thoại tại bước đặt", () => {
  const booking = {
    venueId: "thanh-dong",
    court: "Sân A",
    date: today(),
    hour: 20,
    name: "Nguyễn Văn An",
    phone: "0901234567",
  };
  assert.equal(bookingError(booking, []), "");
  assert.equal(bookingError({ ...booking, phone: "+84901234567" }, []), "");
  for (const invalid of [
    { venueId: "khong-ton-tai" },
    { court: "Sân Z" },
    { hour: 25 },
    { date: "invalid" },
    { date: "2000-01-01" },
    { phone: "12345" },
    { name: " " },
  ]) {
    assert.notEqual(bookingError({ ...booking, ...invalid }, []), "");
  }
  const duplicate = {
    ...booking,
    id: "duplicate",
    total: 220000,
    status: "Đã xác nhận" as const,
    playerBooking: true,
  };
  assert.match(bookingError(booking, [duplicate]), /đã có lịch/);
});
test("Giá theo sân và giờ, tiền cọc làm tròn bằng số nguyên", () => {
  assert.equal(courtPrice(venues[0], "Sân A", 16, initialRates), 180000);
  assert.equal(courtPrice(venues[0], "Sân A", 19, initialRates), 220000);
  assert.equal(courtPrice(venues[0], "Sân C", 19, initialRates), 180000);
  assert.equal(courtPrice(venues[1], "Sân 1", 19, initialRates), 150000);
  assert.equal(deposit(220000), 66000);
  assert.equal(deposit(150001), 45000);
});
test("Ngày lịch phải tồn tại, bao gồm quy tắc năm nhuận", () => {
  assert.equal(isDate("2026-02-30"), false);
  assert.equal(isDate("2026-02-29"), false);
  assert.equal(isDate("2028-02-29"), true);
  assert.equal(isDate("2026-13-01"), false);
});
