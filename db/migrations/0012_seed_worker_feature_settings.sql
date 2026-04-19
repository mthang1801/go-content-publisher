INSERT INTO settings (key, value, description)
VALUES
    ('enable_twitter_crawler', COALESCE((SELECT value FROM settings WHERE key = 'enable_twitter_crawler'), ''), 'Enable or disable the Twitter crawler loop in the worker runtime.'),
    ('enable_rewrite_processor', COALESCE((SELECT value FROM settings WHERE key = 'enable_rewrite_processor'), ''), 'Enable or disable the rewrite processor loop in the worker runtime.')
ON CONFLICT (key) DO UPDATE
SET
    value = EXCLUDED.value,
    description = EXCLUDED.description;
