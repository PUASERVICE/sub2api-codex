-- 090_add_model_status_tables.sql
-- Model status monitoring targets and execution history

CREATE TABLE IF NOT EXISTS model_status_targets (
    id                      BIGSERIAL PRIMARY KEY,
    name                    VARCHAR(120) NOT NULL,
    account_id              BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    model_id                VARCHAR(200) NOT NULL,
    check_interval_seconds  INT NOT NULL DEFAULT 300,
    timeout_seconds         INT NOT NULL DEFAULT 45,
    enabled                 BOOLEAN NOT NULL DEFAULT true,
    latest_status           VARCHAR(20) NOT NULL DEFAULT 'unknown',
    latest_latency_ms       BIGINT,
    latest_error_message    TEXT NOT NULL DEFAULT '',
    latest_response_text    TEXT NOT NULL DEFAULT '',
    consecutive_failures    INT NOT NULL DEFAULT 0,
    last_checked_at         TIMESTAMPTZ,
    last_success_at         TIMESTAMPTZ,
    last_failure_at         TIMESTAMPTZ,
    next_check_at           TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_model_status_targets_account_id
    ON model_status_targets(account_id);

CREATE INDEX IF NOT EXISTS idx_model_status_targets_enabled_next_check
    ON model_status_targets(enabled, next_check_at)
    WHERE enabled = true;

CREATE INDEX IF NOT EXISTS idx_model_status_targets_latest_status
    ON model_status_targets(latest_status);

CREATE TABLE IF NOT EXISTS model_status_checks (
    id                BIGSERIAL PRIMARY KEY,
    target_id         BIGINT NOT NULL REFERENCES model_status_targets(id) ON DELETE CASCADE,
    status            VARCHAR(20) NOT NULL DEFAULT 'unknown',
    latency_ms        BIGINT,
    error_message     TEXT NOT NULL DEFAULT '',
    response_text     TEXT NOT NULL DEFAULT '',
    started_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_model_status_checks_target_created
    ON model_status_checks(target_id, created_at DESC);
