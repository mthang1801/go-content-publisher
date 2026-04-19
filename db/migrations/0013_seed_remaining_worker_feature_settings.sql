INSERT INTO settings (key, value, description)
VALUES
    ('enable_telegram_crawler', COALESCE((SELECT value FROM settings WHERE key = 'enable_telegram_crawler'), ''), 'Enable or disable the Telegram crawler loop in the worker runtime.'),
    ('enable_twitter_publish_vi', COALESCE((SELECT value FROM settings WHERE key = 'enable_twitter_publish_vi'), ''), 'Enable or disable the Vietnamese Twitter publish loop in the worker runtime.'),
    ('enable_twitter_publish_en', COALESCE((SELECT value FROM settings WHERE key = 'enable_twitter_publish_en'), ''), 'Enable or disable the English Twitter publish loop in the worker runtime.')
ON CONFLICT (key) DO UPDATE
SET
    value = EXCLUDED.value,
    description = EXCLUDED.description;
