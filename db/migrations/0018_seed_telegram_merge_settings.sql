INSERT INTO settings (key, value, description)
VALUES
    ('telegram_merge_sources', COALESCE((SELECT value FROM settings WHERE key = 'telegram_merge_sources'), ''), 'Comma-separated Telegram public source handles that should be merged into grouped content items during MTProto polling, for example vnwallstreet,financialjuice.'),
    ('telegram_merge_idle_seconds', COALESCE((SELECT value FROM settings WHERE key = 'telegram_merge_idle_seconds'), '120'), 'For merge-enabled Telegram public sources, split batches when the gap between consecutive messages exceeds this many seconds.'),
    ('telegram_merge_max_seconds', COALESCE((SELECT value FROM settings WHERE key = 'telegram_merge_max_seconds'), '300'), 'For merge-enabled Telegram public sources, split batches when the total time span of a merged batch exceeds this many seconds.')
ON CONFLICT (key) DO UPDATE
SET
    value = EXCLUDED.value,
    description = EXCLUDED.description;
