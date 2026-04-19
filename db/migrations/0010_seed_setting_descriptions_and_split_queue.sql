INSERT INTO settings (key, value, description)
VALUES
    ('auto_publish', COALESCE((SELECT value FROM settings WHERE key = 'auto_publish'), 'false'), 'Enable or disable automatic Telegram publish in the worker loop.'),
    ('rewrite_provider', COALESCE((SELECT value FROM settings WHERE key = 'rewrite_provider'), 'deepseek'), 'Default rewrite provider name used by the content pipeline.'),
    ('twitter_publish_after', COALESCE((SELECT value FROM settings WHERE key = 'twitter_publish_after'), ''), 'Only Twitter-publish content items whose published_at is on or after this RFC3339 timestamp.'),
    ('twitter_publish_source_types', COALESCE((SELECT value FROM settings WHERE key = 'twitter_publish_source_types'), ''), 'Comma-separated source types eligible for Twitter publish, for example twitter or twitter,telegram.'),
    ('twitter_publish_source_tags', COALESCE((SELECT value FROM settings WHERE key = 'twitter_publish_source_tags'), ''), 'Comma-separated source tags eligible for Twitter publish when source topics are not configured.'),
    ('twitter_publish_source_topics', COALESCE((SELECT value FROM settings WHERE key = 'twitter_publish_source_topics'), ''), 'Comma-separated source topics eligible for Twitter publish. This setting takes priority over twitter_publish_source_tags when non-empty.'),
    ('twitter_publish_topic_keywords', COALESCE((SELECT value FROM settings WHERE key = 'twitter_publish_topic_keywords'), ''), 'Comma-separated text keywords used as an additional content-level Twitter publish filter.'),
    ('queue_stale_after_seconds', COALESCE((SELECT value FROM settings WHERE key = 'queue_stale_after_seconds'), ''), 'Legacy fallback timeout in seconds for skipping stale pending or rewritten queue items when the split queue settings are empty.'),
    ('pending_stale_after_seconds', COALESCE((SELECT value FROM settings WHERE key = 'pending_stale_after_seconds'), (SELECT value FROM settings WHERE key = 'queue_stale_after_seconds'), ''), 'Timeout in seconds after which old pending items are marked skipped so newer items can be rewritten first.'),
    ('rewritten_stale_after_seconds', COALESCE((SELECT value FROM settings WHERE key = 'rewritten_stale_after_seconds'), (SELECT value FROM settings WHERE key = 'queue_stale_after_seconds'), ''), 'Timeout in seconds after which old rewritten items are marked skipped so newer items can be published first.')
ON CONFLICT (key) DO UPDATE
SET
    value = EXCLUDED.value,
    description = EXCLUDED.description;
