INSERT INTO settings (key, value, description)
VALUES
    ('crawl_interval_seconds', COALESCE((SELECT value FROM settings WHERE key = 'crawl_interval_seconds'), '300'), 'Worker crawl loop interval in seconds. Runs source revalidation, Telegram ingest, and Twitter crawl.'),
    ('process_interval_seconds', COALESCE((SELECT value FROM settings WHERE key = 'process_interval_seconds'), '30'), 'Worker rewrite loop interval in seconds. Processes pending content through the configured AI provider.'),
    ('publish_interval_seconds', COALESCE((SELECT value FROM settings WHERE key = 'publish_interval_seconds'), '10'), 'Worker Telegram publish loop interval in seconds. Publishes rewritten content when auto_publish is true.'),
    ('twitter_publish_interval_seconds', COALESCE((SELECT value FROM settings WHERE key = 'twitter_publish_interval_seconds'), '600'), 'Worker Twitter publish loop interval in seconds.')
ON CONFLICT (key) DO UPDATE
SET
    value = EXCLUDED.value,
    description = EXCLUDED.description;
