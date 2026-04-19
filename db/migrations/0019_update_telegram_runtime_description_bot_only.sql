UPDATE settings
SET description = 'JSON Telegram runtime configuration with bot_token, publish_targets, ingest_targets, and admin_user_ids for Bot API operations.'
WHERE key = 'telegram_runtime';
