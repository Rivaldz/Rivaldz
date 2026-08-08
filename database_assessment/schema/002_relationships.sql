-- ============================================================================
-- Component 2: Follower Relationships & User Blocking
-- Target Database: PostgreSQL 14+
-- ============================================================================

DO $$ BEGIN
    CREATE TYPE follow_status AS ENUM ('active', 'pending', 'blocked');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 1. Follows Table (M:N Relationship)
CREATE TABLE IF NOT EXISTS follows (
    id              BIGSERIAL PRIMARY KEY,
    follower_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status          follow_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_follow_pair UNIQUE (follower_id, following_id),
    CONSTRAINT chk_no_self_follow CHECK (follower_id <> following_id)
);

-- 2. User Blocks Table
CREATE TABLE IF NOT EXISTS user_blocks (
    id              BIGSERIAL PRIMARY KEY,
    blocker_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_block_pair UNIQUE (blocker_id, blocked_id),
    CONSTRAINT chk_no_self_block CHECK (blocker_id <> blocked_id)
);
