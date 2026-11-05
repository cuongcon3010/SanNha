ALTER TABLE booking_bookings ADD COLUMN deposit_paid BOOLEAN NOT NULL DEFAULT FALSE;
CREATE TABLE booking_payments (
    booking_id TEXT PRIMARY KEY REFERENCES booking_bookings(id),
    order_id TEXT NOT NULL UNIQUE,
    amount BIGINT NOT NULL CHECK(amount>0),
    bank_code TEXT NOT NULL,
    bank_account TEXT NOT NULL,
    bank_name TEXT NOT NULL,
    content TEXT NOT NULL,
    environment TEXT NOT NULL CHECK(environment IN ('sandbox','production')),
    qr_code TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','paid','review')),
    created_at BIGINT NOT NULL
);
CREATE TABLE booking_payment_events (
    transaction_id TEXT PRIMARY KEY,
    bank_account TEXT NOT NULL,
    reference_number TEXT NOT NULL,
    booking_id TEXT NOT NULL REFERENCES booking_bookings(id),
    fingerprint TEXT NOT NULL,
    amount BIGINT NOT NULL,
    trans_type TEXT NOT NULL,
    transaction_time BIGINT NOT NULL,
    result TEXT NOT NULL CHECK(result IN ('paid','review')),
    reason TEXT NOT NULL,
    received_at BIGINT NOT NULL,
    UNIQUE(bank_account,reference_number)
);
CREATE TABLE booking_callback_tokens (
    token_hash TEXT PRIMARY KEY,
    expires_at BIGINT NOT NULL
);
