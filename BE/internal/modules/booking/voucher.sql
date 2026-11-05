CREATE TABLE booking_vouchers (
    code TEXT PRIMARY KEY CHECK(code ~ '^[A-Z0-9]{3,24}$'),
    kind TEXT NOT NULL CHECK(kind IN ('fixed','percent')),
    value BIGINT NOT NULL CHECK(value>0 AND ((kind='percent' AND value<=99) OR (kind='fixed' AND value<=10000000))),
    min_total BIGINT NOT NULL CHECK(min_total BETWEEN 0 AND 10000000),
    max_discount BIGINT NOT NULL CHECK(max_discount BETWEEN 0 AND 10000000),
    starts_at BIGINT NOT NULL,
    ends_at BIGINT NOT NULL CHECK(ends_at>starts_at),
    max_uses INTEGER NOT NULL CHECK(max_uses BETWEEN 1 AND 1000000),
    used_count INTEGER NOT NULL DEFAULT 0 CHECK(used_count>=0 AND used_count<=max_uses),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at BIGINT NOT NULL
);
ALTER TABLE booking_bookings ADD COLUMN subtotal BIGINT NOT NULL DEFAULT 0;
ALTER TABLE booking_bookings ADD COLUMN discount BIGINT NOT NULL DEFAULT 0;
ALTER TABLE booking_bookings ADD COLUMN voucher_code TEXT NOT NULL DEFAULT '';
UPDATE booking_bookings SET subtotal=total;
ALTER TABLE booking_bookings ADD CONSTRAINT booking_discount_valid CHECK(discount>=0 AND subtotal=total+discount);
CREATE TABLE booking_voucher_uses (
    booking_id TEXT PRIMARY KEY REFERENCES booking_bookings(id),
    voucher_code TEXT NOT NULL REFERENCES booking_vouchers(code),
    user_id TEXT NOT NULL,
    released BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE UNIQUE INDEX booking_voucher_user_active ON booking_voucher_uses(voucher_code,user_id) WHERE NOT released;
