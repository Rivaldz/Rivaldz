-- ============================================================================
-- Component 5: Private Messaging System
-- Target Database: PostgreSQL 14+
-- ============================================================================

DO $$ BEGIN
    CREATE TYPE message_type AS ENUM ('text', 'image', 'video', 'audio', 'file', 'system');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 1. Conversations Table
CREATE TABLE IF NOT EXISTS conversations (
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(255),                                        -- Optional title for group chats
    is_group        BOOLEAN NOT NULL DEFAULT FALSE,
    created_by      BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Conversation Participants Table
CREATE TABLE IF NOT EXISTS conversation_participants (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL DEFAULT 'member',               -- e.g. 'admin', 'member'
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at         TIMESTAMPTZ,
    is_muted        BOOLEAN NOT NULL DEFAULT FALSE,
    last_read_at    TIMESTAMPTZ,                                         -- Per-user read marker

    CONSTRAINT uq_conversation_user UNIQUE (conversation_id, user_id)
);

-- 3. Messages Table
CREATE TABLE IF NOT EXISTS messages (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id       BIGINT REFERENCES users(id) ON DELETE SET NULL,
    message_type    message_type NOT NULL DEFAULT 'text',
    content         TEXT,
    media_url       VARCHAR(1000),
    reply_to_id     BIGINT REFERENCES messages(id) ON DELETE SET NULL,
    metadata        JSONB DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at       TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_message_has_content CHECK (content IS NOT NULL OR media_url IS NOT NULL)
);

-- 4. Message Receipts Table (Per-user delivery & read status)
CREATE TABLE IF NOT EXISTS message_receipts (
    id              BIGSERIAL PRIMARY KEY,
    message_id      BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    delivered_at    TIMESTAMPTZ,
    read_at         TIMESTAMPTZ,

    CONSTRAINT uq_message_receipt UNIQUE (message_id, user_id)
);
