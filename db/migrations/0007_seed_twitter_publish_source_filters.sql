INSERT INTO settings (key, value)
VALUES
    ('twitter_publish_source_tags', ''),
    ('twitter_publish_source_topics', '')
ON CONFLICT (key) DO NOTHING;
