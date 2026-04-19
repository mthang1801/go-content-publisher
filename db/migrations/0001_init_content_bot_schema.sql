CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS sources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    type text NOT NULL,
    handle text NOT NULL,
    name text NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (type IN ('telegram', 'twitter'))
);

CREATE TABLE IF NOT EXISTS content_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id uuid NULL REFERENCES sources (id) ON DELETE SET NULL,
    external_id text NOT NULL DEFAULT '',
    original_text text NOT NULL,
    author_name text NOT NULL DEFAULT 'Unknown',
    source_url text NULL,
    crawled_at timestamptz NOT NULL DEFAULT now(),
    status text NOT NULL DEFAULT 'pending',
    rewritten_text text NULL,
    rewritten_text_en text NULL,
    tweet_text_vi text NULL,
    tweet_text_en text NULL,
    fact_check_note text NULL,
    fail_reason text NULL,
    tweet_vi_id text NULL,
    tweet_en_id text NULL,
    published_at timestamptz NULL,
    published_msg_id text NULL,
    CHECK (status IN ('pending', 'processing', 'rewritten', 'published', 'failed', 'skipped'))
);

CREATE TABLE IF NOT EXISTS settings (
    key text PRIMARY KEY,
    value text NOT NULL
);

CREATE TABLE IF NOT EXISTS logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    level text NOT NULL,
    module text NOT NULL,
    message text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
