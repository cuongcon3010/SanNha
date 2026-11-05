CREATE TABLE booking_bookings (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    venue_id TEXT NOT NULL,
    court_id TEXT NOT NULL,
    date TEXT NOT NULL,
    hour INTEGER NOT NULL CHECK(hour BETWEEN 16 AND 21),
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    total BIGINT NOT NULL CHECK(total>=0),
    deposit BIGINT NOT NULL CHECK(deposit>=0),
    status TEXT NOT NULL CHECK(status IN ('pending','confirmed','completed','cancelled','blocked')),
    created_at TEXT NOT NULL
);
CREATE UNIQUE INDEX booking_active_slot ON booking_bookings(court_id,date,hour) WHERE status<>'cancelled';
CREATE INDEX booking_user ON booking_bookings(user_id,date,hour);
CREATE INDEX booking_venue ON booking_bookings(venue_id,date,hour);
