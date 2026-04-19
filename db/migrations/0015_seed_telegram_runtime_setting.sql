INSERT INTO settings (key, value, json_value, description)
VALUES
    (
        'telegram_runtime',
        '',
        NULL,
        'JSON Telegram runtime configuration with bot_token, publish_targets, ingest_targets, admin_user_ids, api_id, api_hash, and session.'
    )
ON CONFLICT (key) DO UPDATE
SET
    description = EXCLUDED.description;
