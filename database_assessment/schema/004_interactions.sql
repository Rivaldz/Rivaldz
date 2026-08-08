-- ============================================================================
-- Component 4: Comments and Reactions
-- Target Database: PostgreSQL 14+
-- ============================================================================

DO $$ BEGIN
    CREATE TYPE reaction_type AS ENUM ('like', 'love', 'haha', 'wow', 'sad', 'angry');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE reactable_type AS ENUM ('post', 'comment');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 1. Threaded Comments Table
CREATE TABLE IF NOT EXISTS comments (
    id              BIGSERIAL PRIMARY KEY,
    post_id         BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id       BIGINT REFERENCES comments(id) ON DELETE CASCADE,  -- Threaded replies
    content         TEXT NOT NULL,
    depth           SMALLINT NOT NULL DEFAULT 0,                        -- Nesting level (0-5)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,                                         -- Soft delete

    CONSTRAINT chk_comment_depth CHECK (depth >= 0 AND depth <= 5),
    CONSTRAINT chk_comment_content CHECK (char_length(content) > 0)
);

-- 2. Polymorphic Reactions Table (Supports Posts and Comments)
CREATE TABLE IF NOT EXISTS reactions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reactable_type  reactable_type NOT NULL,
    reactable_id    BIGINT NOT NULL,                                     -- post_id or comment_id
    reaction_type   reaction_type NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_user_reaction UNIQUE (user_id, reactable_type, reactable_id)
);
