DELETE FROM settings
WHERE key IN (
    'telegram_merge_sources',
    'telegram_merge_idle_seconds',
    'telegram_merge_max_seconds'
);
