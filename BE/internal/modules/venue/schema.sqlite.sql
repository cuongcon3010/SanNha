CREATE TABLE venue_venues (
    id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL,
    name TEXT NOT NULL,
    sport TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    amenities TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected')),
    review_comment TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);
CREATE INDEX venue_owner ON venue_venues(owner_id);
CREATE TABLE venue_courts (
    id TEXT PRIMARY KEY,
    venue_id TEXT NOT NULL REFERENCES venue_venues(id),
    name TEXT NOT NULL,
    day_price INTEGER NOT NULL CHECK(day_price BETWEEN 1000 AND 10000000),
    evening_price INTEGER NOT NULL CHECK(evening_price BETWEEN 1000 AND 10000000),
    maintenance INTEGER NOT NULL DEFAULT 0 CHECK(maintenance IN (0,1)),
    UNIQUE(venue_id,name)
);
CREATE TABLE venue_favorites (
    user_id TEXT NOT NULL,
    venue_id TEXT NOT NULL REFERENCES venue_venues(id),
    PRIMARY KEY(user_id,venue_id)
);
