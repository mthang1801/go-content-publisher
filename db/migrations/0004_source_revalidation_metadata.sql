ALTER TABLE sources
    ADD COLUMN IF NOT EXISTS last_crawled_at timestamptz NULL,
    ADD COLUMN IF NOT EXISTS last_check_at timestamptz NULL,
    ADD COLUMN IF NOT EXISTS last_error text NULL;

CREATE INDEX IF NOT EXISTS idx_sources_active_due_validation
    ON sources (created_at)
    WHERE is_active = true AND last_check_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sources_inactive_due_recheck
    ON sources ((COALESCE(last_check_at, last_crawled_at, created_at)))
    WHERE is_active = false;
