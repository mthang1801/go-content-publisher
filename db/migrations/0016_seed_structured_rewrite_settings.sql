INSERT INTO settings (key, value, description)
VALUES
    ('rewrite_duplicate_window_hours', COALESCE((SELECT value FROM settings WHERE key = 'rewrite_duplicate_window_hours'), '48'), 'Lookback window in hours for rewrite duplicate detection before and after AI rewriting.'),
    ('rewrite_duplicate_original_threshold', COALESCE((SELECT value FROM settings WHERE key = 'rewrite_duplicate_original_threshold'), '0.70'), 'Similarity threshold for skipping content whose original text duplicates recent processed content.'),
    ('rewrite_duplicate_rewritten_threshold', COALESCE((SELECT value FROM settings WHERE key = 'rewrite_duplicate_rewritten_threshold'), '0.75'), 'Similarity threshold for skipping content whose rewritten text duplicates recent rewritten or published content.')
ON CONFLICT (key) DO UPDATE
SET
    value = EXCLUDED.value,
    description = EXCLUDED.description;
