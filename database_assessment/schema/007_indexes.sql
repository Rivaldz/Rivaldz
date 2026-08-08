-- ============================================================================
-- Component 7: Indexing Strategy for Common Queries & Performance Optimization
-- Target Database: PostgreSQL 14+
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1. Users & Profiles Indexes
-- ----------------------------------------------------------------------------
-- Partial index on inactive users for admin audits
CREATE INDEX IF NOT EXISTS idx_users_status_non_active 
    ON users(status) 
    WHERE status != 'active';

CREATE INDEX IF NOT EXISTS idx_users_created_at 
    ON users(created_at);

-- Trigram index for fuzzy username search
CREATE INDEX IF NOT EXISTS idx_users_username_trgm 
    ON users USING gin(username gin_trgm_ops);

-- ----------------------------------------------------------------------------
-- 2. Follows & Blocks Indexes
-- ----------------------------------------------------------------------------
-- Composite index for looking up followers of a target user ("who follows me?")
CREATE INDEX IF NOT EXISTS idx_follows_following_status 
    ON follows(following_id, status);

-- Partial index for active follower queries
CREATE INDEX IF NOT EXISTS idx_follows_follower_active 
    ON follows(follower_id) 
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_user_blocks_blocker 
    ON user_blocks(blocker_id, blocked_id);

-- ----------------------------------------------------------------------------
-- 3. Posts, Media, and Hashtags Indexes
-- ----------------------------------------------------------------------------
-- Composite partial index for user timeline query (user's active posts by date)
CREATE INDEX IF NOT EXISTS idx_posts_user_created_active 
    ON posts(user_id, created_at DESC) 
    WHERE deleted_at IS NULL;

-- Global feed timeline index
CREATE INDEX IF NOT EXISTS idx_posts_created_at_active 
    ON posts(created_at DESC) 
    WHERE deleted_at IS NULL;

-- Visibility filtered feed index
CREATE INDEX IF NOT EXISTS idx_posts_visibility_created 
    ON posts(visibility, created_at DESC) 
    WHERE deleted_at IS NULL;

-- BRIN index for massive time-series scan optimization on posts
CREATE INDEX IF NOT EXISTS idx_posts_created_brin 
    ON posts USING brin(created_at);

-- Media order index
CREATE INDEX IF NOT EXISTS idx_post_media_post_order 
    ON post_media(post_id, sort_order);

-- Hashtags indexing
CREATE INDEX IF NOT EXISTS idx_post_hashtags_hashtag_id 
    ON post_hashtags(hashtag_id);

CREATE INDEX IF NOT EXISTS idx_hashtags_name_trgm 
    ON hashtags USING gin(name gin_trgm_ops);

-- ----------------------------------------------------------------------------
-- 4. Comments & Reactions Indexes
-- ----------------------------------------------------------------------------
-- Post comments ordering index
CREATE INDEX IF NOT EXISTS idx_comments_post_created 
    ON comments(post_id, created_at ASC) 
    WHERE deleted_at IS NULL;

-- Threaded reply lookup index
CREATE INDEX IF NOT EXISTS idx_comments_parent_id 
    ON comments(parent_id) 
    WHERE parent_id IS NOT NULL;

-- User comment history index
CREATE INDEX IF NOT EXISTS idx_comments_user_created 
    ON comments(user_id, created_at DESC);

-- Polymorphic reactions lookups
CREATE INDEX IF NOT EXISTS idx_reactions_target 
    ON reactions(reactable_type, reactable_id);

CREATE INDEX IF NOT EXISTS idx_reactions_user_target 
    ON reactions(user_id, reactable_type, reactable_id);

-- ----------------------------------------------------------------------------
-- 5. Messaging Indexes
-- ----------------------------------------------------------------------------
-- Active user conversations lookup
CREATE INDEX IF NOT EXISTS idx_conv_participants_user_active 
    ON conversation_participants(user_id) 
    WHERE left_at IS NULL;

-- Conversation messages in reverse chronological order
CREATE INDEX IF NOT EXISTS idx_messages_conv_created 
    ON messages(conversation_id, created_at DESC) 
    WHERE deleted_at IS NULL;

-- Sender message lookup
CREATE INDEX IF NOT EXISTS idx_messages_sender_created 
    ON messages(sender_id, created_at DESC);

-- Unread message receipts lookup
CREATE INDEX IF NOT EXISTS idx_message_receipts_unread 
    ON message_receipts(user_id) 
    WHERE read_at IS NULL;

-- ----------------------------------------------------------------------------
-- 6. Notifications & Activity Log Indexes
-- ----------------------------------------------------------------------------
-- User notification inbox index
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_created 
    ON notifications(recipient_id, created_at DESC);

-- Partial index for fast unread notifications badge count
CREATE INDEX IF NOT EXISTS idx_notifications_unread 
    ON notifications(recipient_id) 
    WHERE is_read = FALSE;

-- Aggregation group lookup
CREATE INDEX IF NOT EXISTS idx_notifications_group_key 
    ON notifications(group_key, recipient_id) 
    WHERE group_key IS NOT NULL;

-- User activity timeline
CREATE INDEX IF NOT EXISTS idx_activity_log_user_created 
    ON activity_log(user_id, created_at DESC);

-- BRIN index for time-series activity log queries
CREATE INDEX IF NOT EXISTS idx_activity_log_created_brin 
    ON activity_log USING brin(created_at);
