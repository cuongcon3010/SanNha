CREATE TABLE identity_users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    area TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('user','owner','admin')),
    locked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TEXT NOT NULL
);
CREATE UNIQUE INDEX identity_users_email_key ON identity_users(lower(email));
CREATE TABLE identity_sessions (
    token_hash TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    expires_at BIGINT NOT NULL
);
CREATE INDEX identity_sessions_user ON identity_sessions(user_id);
