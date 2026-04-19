INSERT INTO settings (key, value)
VALUES
    ('queue_stale_after_seconds', '')
ON CONFLICT (key) DO NOTHING;
