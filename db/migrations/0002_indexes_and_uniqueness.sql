CREATE UNIQUE INDEX IF NOT EXISTS idx_sources_type_handle_unique
    ON sources (type, handle);

CREATE UNIQUE INDEX IF NOT EXISTS idx_content_items_source_external_unique
    ON content_items (source_id, external_id);

CREATE INDEX IF NOT EXISTS idx_content_items_status
    ON content_items (status);

CREATE INDEX IF NOT EXISTS idx_content_items_crawled_at
    ON content_items (crawled_at DESC);

CREATE INDEX IF NOT EXISTS idx_logs_created_at
    ON logs (created_at DESC);
