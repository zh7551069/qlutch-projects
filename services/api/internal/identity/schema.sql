-- Identity tables for accounts, 7-step onboarding, locale / region preference,
-- and session token hashes. Applied with search_path set to the target schema.
-- OTP challenges, rate limits, and idempotency records are not stored here.

CREATE TABLE IF NOT EXISTS identity_accounts (
    user_id TEXT PRIMARY KEY,
    phone_region TEXT NOT NULL,
    phone_national TEXT NOT NULL,
    locale TEXT NOT NULL,
    region_pref TEXT NOT NULL,
    has_first_pack BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (phone_region, phone_national),
    CHECK (phone_region IN ('+86', '+852')),
    CHECK (region_pref IN ('+86', '+852')),
    CHECK (locale IN ('zh-Hans', 'zh-Hant'))
);

CREATE TABLE IF NOT EXISTS identity_onboarding_answers (
    user_id TEXT NOT NULL REFERENCES identity_accounts (user_id),
    step_id TEXT NOT NULL,
    answer TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, step_id),
    CHECK (step_id IN ('nickname', 'age', 'study_or_work', 'identity', 'how_to_address', 'exam_track', 'daily_time')),
    CHECK (char_length(answer) > 0)
);

CREATE TABLE IF NOT EXISTS identity_sessions (
    session_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES identity_accounts (user_id),
    access_hash BYTEA NOT NULL UNIQUE,
    refresh_hash BYTEA NOT NULL UNIQUE,
    access_expiry TIMESTAMPTZ NOT NULL,
    refresh_expiry TIMESTAMPTZ NOT NULL,
    CHECK (octet_length(access_hash) = 32),
    CHECK (octet_length(refresh_hash) = 32)
);
