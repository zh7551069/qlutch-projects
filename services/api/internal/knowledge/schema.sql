-- Knowledge packs, material metadata, import jobs, skip records, and
-- idempotency keys. Material bytes are not stored here.
-- Map nodes, prefetch stems, and usage rows are added for the import worker.
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

CREATE TABLE IF NOT EXISTS knowledge_map_nodes (
    node_id TEXT PRIMARY KEY,
    pack_id TEXT NOT NULL REFERENCES knowledge_packs (pack_id),
    job_id TEXT NOT NULL REFERENCES knowledge_jobs (job_id),
    parent_id TEXT,
    kind TEXT NOT NULL,
    title TEXT NOT NULL,
    position INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (kind IN ('catalog', 'chapter', 'point')),
    CHECK (char_length(title) > 0),
    CHECK (char_length(title) <= 80),
    CHECK (position >= 0)
);

CREATE INDEX IF NOT EXISTS knowledge_map_nodes_pack
    ON knowledge_map_nodes (pack_id, position);

CREATE TABLE IF NOT EXISTS knowledge_prefetch (
    prefetch_id TEXT PRIMARY KEY,
    pack_id TEXT NOT NULL REFERENCES knowledge_packs (pack_id),
    job_id TEXT NOT NULL REFERENCES knowledge_jobs (job_id),
    node_id TEXT NOT NULL,
    position INT NOT NULL,
    stem TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (char_length(stem) > 0),
    CHECK (char_length(stem) <= 80),
    CHECK (position >= 0)
);

CREATE TABLE IF NOT EXISTS knowledge_usage_ledger (
    entry_id TEXT PRIMARY KEY,
    scene_code TEXT NOT NULL,
    subclass TEXT NOT NULL,
    tier TEXT NOT NULL,
    cache_hit BOOLEAN NOT NULL,
    success BOOLEAN NOT NULL,
    retry_index INT NOT NULL,
    tokens INT NOT NULL,
    metric_kind TEXT NOT NULL,
    metric_value INT NOT NULL,
    pack_id TEXT NOT NULL,
    job_id TEXT NOT NULL,
    template_version TEXT NOT NULL DEFAULT '',
    caller TEXT NOT NULL,
    fingerprint TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (scene_code IN ('S01', 'S02')),
    CHECK (char_length(subclass) > 0),
    CHECK (subclass <> '其他'),
    CHECK (caller = 'learner'),
    CHECK (retry_index >= 0),
    CHECK (tokens >= 0),
    CHECK (metric_value >= 0),
    CHECK (char_length(pack_id) > 0),
    CHECK (char_length(job_id) > 0),
    CHECK (char_length(tier) > 0)
);

CREATE INDEX IF NOT EXISTS knowledge_usage_ledger_scene
    ON knowledge_usage_ledger (scene_code, subclass, cache_hit);
