ALTER TABLE settings
ADD COLUMN IF NOT EXISTS json_value jsonb NULL;
