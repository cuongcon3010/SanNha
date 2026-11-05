CREATE TABLE identity_oauth_accounts (
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES identity_users(id),
    PRIMARY KEY(provider,subject)
);
CREATE TABLE identity_oauth_flows (
    state_hash TEXT PRIMARY KEY,
    browser_hash TEXT NOT NULL,
    provider TEXT NOT NULL,
    verifier TEXT NOT NULL,
    expires_at BIGINT NOT NULL
);
