INSERT INTO settings (key, value)
VALUES
    ('twitter_publish_after', ''),
    ('twitter_publish_source_types', ''),
    ('twitter_publish_topic_keywords', '')
ON CONFLICT (key) DO NOTHING;
