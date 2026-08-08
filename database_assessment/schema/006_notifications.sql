-- ============================================================================
-- Component 6: Activity Feeds & Notifications System
-- Target Database: PostgreSQL 14+
-- ============================================================================

DO $$ BEGIN
    CREATE TYPE notification_type AS ENUM (
        'follow', 'follow_request',
        'post_like', 'post_comment', 'post_share',
        'comment_like', 'comment_reply',
        'mention',
        'message',
        'system'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 1. Notifications Table
CREATE TABLE IF NOT EXISTS notifications (
    id                BIGSERIAL PRIMARY KEY,
    recipient_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id          BIGINT REFERENCES users(id) ON DELETE SET NULL,     -- User who triggered event
    notification_type notification_type NOT NULL,
    target_type       VARCHAR(50),                                        -- e.g. 'post', 'comment'
    target_id         BIGINT,
    group_key         VARCHAR(255),                                       -- For aggregation ("X and 5 others")
    message           TEXT,
    is_read           BOOLEAN NOT NULL DEFAULT FALSE,
    read_at           TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Activity Log Table (For user activity feeds & auditing)
CREATE TABLE IF NOT EXISTS activity_log (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action          VARCHAR(50) NOT NULL,                               -- 'posted', 'liked', 'followed', etc.
    target_type     VARCHAR(50) NOT NULL,
    target_id       BIGINT NOT NULL,
    metadata        JSONB DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
