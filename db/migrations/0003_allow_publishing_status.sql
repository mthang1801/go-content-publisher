ALTER TABLE content_items
    DROP CONSTRAINT IF EXISTS content_items_status_check;

ALTER TABLE content_items
    ADD CONSTRAINT content_items_status_check
    CHECK (status IN ('pending', 'processing', 'rewritten', 'publishing', 'published', 'failed', 'skipped'));
