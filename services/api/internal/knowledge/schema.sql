-- Knowledge packs, material metadata, import jobs, skip records, and
-- idempotency keys. Material bytes are not stored here.
-- Applied with search_path set to the target schema.

CREATE TABLE IF NOT EXISTS knowledge_packs (
    pack_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS knowledge_materials (
    material_id TEXT PRIMARY KEY,
    pack_id TEXT NOT NULL REFERENCES knowledge_packs (pack_id),
    user_id TEXT NOT NULL,
    source TEXT NOT NULL,
    content_type TEXT NOT NULL,
    byte_size BIGINT NOT NULL,
    filename TEXT,
    object_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (source IN ('camera', 'file', 'paste_note')),
    CHECK (byte_size > 0),
    CHECK (char_length(object_key) > 0)
);

CREATE TABLE IF NOT EXISTS knowledge_jobs (
    job_id TEXT PRIMARY KEY,
    pack_id TEXT NOT NULL REFERENCES knowledge_packs (pack_id),
    material_id TEXT NOT NULL REFERENCES knowledge_materials (material_id),
    user_id TEXT NOT NULL,
    source TEXT NOT NULL,
    status TEXT NOT NULL,
    progress_message TEXT,
    failure_code TEXT,
    failure_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (source IN ('camera', 'file', 'paste_note')),
    CHECK (status IN ('accepted', 'running', 'map_ready', 'failed_retryable', 'failed_terminal'))
);

CREATE UNIQUE INDEX IF NOT EXISTS knowledge_jobs_one_active
    ON knowledge_jobs (pack_id)
    WHERE status IN ('accepted', 'running');

CREATE TABLE IF NOT EXISTS knowledge_first_pack_skips (
    user_id TEXT PRIMARY KEY,
    skipped_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS knowledge_idempotency (
    user_id TEXT NOT NULL,
    op TEXT NOT NULL,
    idem_key TEXT NOT NULL,
    body_hash BYTEA NOT NULL,
    job_id TEXT,
    err_code TEXT,
    err_details JSONB,
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, op, idem_key),
    CHECK (op IN ('import', 'skip')),
    CHECK (octet_length(body_hash) = 32)
);
