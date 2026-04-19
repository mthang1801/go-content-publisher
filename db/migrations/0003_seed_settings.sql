INSERT INTO settings (key, value)
VALUES
    ('auto_publish', 'false'),
    ('rewrite_provider', 'deepseek')
ON CONFLICT (key) DO NOTHING;
