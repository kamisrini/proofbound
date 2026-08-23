CREATE TABLE IF NOT EXISTS events (
 seq BIGSERIAL PRIMARY KEY,
 event_id TEXT NOT NULL UNIQUE,
 source TEXT NOT NULL,
 native_id TEXT NOT NULL,
 kind TEXT NOT NULL,
 occurred_at TIMESTAMPTZ NOT NULL,
 recorded_at TIMESTAMPTZ NOT NULL,
 payload JSON NOT NULL,
 content_sha TEXT NOT NULL,
 connector_version TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS events_idempotency ON events(source, native_id, content_sha);
CREATE INDEX IF NOT EXISTS events_source_kind_seq ON events(source, kind, seq);
CREATE INDEX IF NOT EXISTS events_occurred_at ON events(occurred_at);
CREATE TABLE IF NOT EXISTS sync_runs (
 id BIGSERIAL PRIMARY KEY,
 connector TEXT NOT NULL,
 cursor_json JSON,
 started_at TIMESTAMPTZ NOT NULL,
 finished_at TIMESTAMPTZ,
 events_appended BIGINT NOT NULL DEFAULT 0,
 error TEXT
);
