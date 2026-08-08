-- ============================================================================
-- Component 3: Posts, Post Media, Hashtags, and Post Stats
-- Target Database: PostgreSQL 14+
-- ============================================================================

DO $$ BEGIN
    CREATE TYPE post_visibility AS ENUM ('public', 'followers_only', 'private');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE media_type AS ENUM ('image', 'video', 'audio', 'document');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 1. Posts Table
CREATE TABLE IF NOT EXISTS posts (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content         TEXT,
    visibility      post_visibility NOT NULL DEFAULT 'public',
    is_pinned       BOOLEAN NOT NULL DEFAULT FALSE,
    location_name   VARCHAR(255),
    location_lat    DECIMAL(10, 8),
    location_lng    DECIMAL(11, 8),
    metadata        JSONB DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ                          -- Soft delete timestamp
);

-- 2. Post Media Table (1:N Attachments)
CREATE TABLE IF NOT EXISTS post_media (
    id              BIGSERIAL PRIMARY KEY,
    post_id         BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    media_type      media_type NOT NULL,
    url             VARCHAR(1000) NOT NULL,
    thumbnail_url   VARCHAR(1000),
    alt_text        VARCHAR(500),
    width           INTEGER,
    height          INTEGER,
    duration_secs   INTEGER,
    file_size_bytes BIGINT,
    sort_order      SMALLINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Hashtags Table
CREATE TABLE IF NOT EXISTS hashtags (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_hashtags_name UNIQUE (name)
);

-- 4. Post Hashtags Junction Table (M:N)
CREATE TABLE IF NOT EXISTS post_hashtags (
    post_id         BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    hashtag_id      BIGINT NOT NULL REFERENCES hashtags(id) ON DELETE CASCADE,

    PRIMARY KEY (post_id, hashtag_id)
);

-- 5. Post Statistics Table (Denormalized Counters for O(1) Reads)
CREATE TABLE IF NOT EXISTS post_stats (
    post_id         BIGINT PRIMARY KEY REFERENCES posts(id) ON DELETE CASCADE,
    likes_count     BIGINT NOT NULL DEFAULT 0,
    comments_count  BIGINT NOT NULL DEFAULT 0,
    shares_count    BIGINT NOT NULL DEFAULT 0,
    views_count     BIGINT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_positive_post_stats CHECK (
        likes_count >= 0 AND comments_count >= 0 AND
        shares_count >= 0 AND views_count >= 0
    )
);
