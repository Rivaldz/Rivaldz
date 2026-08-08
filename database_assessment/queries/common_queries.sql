-- ============================================================================
-- Common Complex SQL Queries for Social Media Platform
-- Target Database: PostgreSQL 14+
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Query 1: Fetch Threaded Comments for a Post (Recursive CTE)
-- Description: Recursively retrieves comments up to depth 5 for a post.
-- Parameters:
--   :post_id (BIGINT) - Target post ID
-- ----------------------------------------------------------------------------
PREPARE get_threaded_comments(bigint) AS
WITH RECURSIVE comment_tree AS (
    -- Base Case: Top-level comments (parent_id IS NULL)
    SELECT
        c.id,
        c.post_id,
        c.user_id,
        c.parent_id,
        c.content,
        c.depth,
        c.created_at,
        ARRAY[c.id] AS path
    FROM comments c
    WHERE c.post_id = $1
      AND c.parent_id IS NULL
      AND c.deleted_at IS NULL

    UNION ALL

    -- Recursive Step: Child replies
    SELECT
        child.id,
        child.post_id,
        child.user_id,
        child.parent_id,
        child.content,
        child.depth,
        child.created_at,
        ct.path || child.id
    FROM comments child
    INNER JOIN comment_tree ct ON child.parent_id = ct.id
    WHERE child.deleted_at IS NULL
      AND child.depth <= 5
)
SELECT
    ct.id AS comment_id,
    ct.parent_id,
    ct.depth,
    ct.content,
    ct.created_at,
    u.username AS author_username,
    up.avatar_url AS author_avatar_url,
    ct.path
FROM comment_tree ct
INNER JOIN users u ON u.id = ct.user_id
LEFT JOIN user_profiles up ON up.user_id = u.id
ORDER BY ct.path;


-- ----------------------------------------------------------------------------
-- Query 2: Active User Inbox (Conversations with Last Message & Unread Count)
-- Description: Lists user's conversations with latest message content and unread counter.
-- Parameters:
--   :user_id (BIGINT) - Current active user ID
-- ----------------------------------------------------------------------------
PREPARE get_user_conversations(bigint) AS
SELECT
    c.id AS conversation_id,
    c.title,
    c.is_group,
    cp.last_read_at,
    m.id AS last_message_id,
    m.content AS last_message_content,
    m.created_at AS last_message_at,
    su.username AS last_message_sender,
    (
        SELECT COUNT(*)
        FROM messages unread_m
        WHERE unread_m.conversation_id = c.id
          AND unread_m.created_at > COALESCE(cp.last_read_at, '1970-01-01'::timestamptz)
          AND unread_m.sender_id != $1
          AND unread_m.deleted_at IS NULL
    ) AS unread_count
FROM conversation_participants cp
INNER JOIN conversations c ON c.id = cp.conversation_id
LEFT JOIN LATERAL (
    SELECT id, content, created_at, sender_id
    FROM messages
    WHERE conversation_id = c.id
      AND deleted_at IS NULL
    ORDER BY created_at DESC
    LIMIT 1
) m ON TRUE
LEFT JOIN users su ON su.id = m.sender_id
WHERE cp.user_id = $1
  AND cp.left_at IS NULL
ORDER BY COALESCE(m.created_at, c.created_at) DESC;


-- ----------------------------------------------------------------------------
-- Query 3: Aggregated User Notifications Inbox
-- Description: Retrieves user notifications with actor details.
-- Parameters:
--   :user_id   (BIGINT) - Recipient user ID
--   :limit_val (INT)    - Page size
-- ----------------------------------------------------------------------------
PREPARE get_notifications(bigint, int) AS
SELECT
    n.id AS notification_id,
    n.notification_type,
    n.target_type,
    n.target_id,
    n.group_key,
    n.message,
    n.is_read,
    n.created_at,
    au.username AS actor_username,
    aup.avatar_url AS actor_avatar_url
FROM notifications n
LEFT JOIN users au ON au.id = n.actor_id
LEFT JOIN user_profiles aup ON aup.user_id = au.id
WHERE n.recipient_id = $1
ORDER BY n.created_at DESC
LIMIT $2;


-- ----------------------------------------------------------------------------
-- Query 4: Mutual Followings (Mutual Friends)
-- Description: Finds overlapping active followings between user A and user B.
-- Parameters:
--   :user_a (BIGINT)
--   :user_b (BIGINT)
-- ----------------------------------------------------------------------------
PREPARE get_mutual_followings(bigint, bigint) AS
SELECT
    u.id AS user_id,
    u.username,
    up.display_name,
    up.avatar_url
FROM follows f1
INNER JOIN follows f2 ON f1.following_id = f2.following_id
INNER JOIN users u ON u.id = f1.following_id AND u.status = 'active'
LEFT JOIN user_profiles up ON up.user_id = u.id
WHERE f1.follower_id = $1 AND f1.status = 'active'
  AND f2.follower_id = $2 AND f2.status = 'active'
  AND f1.following_id NOT IN ($1, $2)
ORDER BY u.username;
