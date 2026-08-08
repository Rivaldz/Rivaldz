-- ============================================================================
-- Complex SQL Queries for Feed Generation and Discovery
-- Target Database: PostgreSQL 14+
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Query 1: Personalized Timeline Feed (Fan-out-on-Read)
-- Description: Fetches posts from followed users and the user's own posts,
--              excluding blocked relationships, with cursor-based pagination
--              and JSON aggregated media attachments.
-- Parameters:
--   :current_user_id   (BIGINT)    - Logged in user ID
--   :cursor_timestamp  (TIMESTAMPTZ)- Created timestamp of last post seen (default NOW())
--   :limit_val         (INT)       - Items per page (e.g. 20)
-- ----------------------------------------------------------------------------
PREPARE get_home_feed(bigint, timestamptz, int) AS
SELECT
    p.id AS post_id,
    p.user_id AS author_id,
    u.username AS author_username,
    up.display_name AS author_display_name,
    up.avatar_url AS author_avatar_url,
    p.content,
    p.visibility,
    p.location_name,
    p.created_at,
    COALESCE(ps.likes_count, 0) AS likes_count,
    COALESCE(ps.comments_count, 0) AS comments_count,
    COALESCE(ps.shares_count, 0) AS shares_count,
    -- User interaction check
    EXISTS (
        SELECT 1 FROM reactions r
        WHERE r.reactable_type = 'post'
          AND r.reactable_id = p.id
          AND r.user_id = $1
    ) AS user_has_liked,
    -- JSON array aggregation of media attachments
    COALESCE(
        (
            SELECT json_agg(
                json_build_object(
                    'id', pm.id,
                    'media_type', pm.media_type,
                    'url', pm.url,
                    'thumbnail_url', pm.thumbnail_url,
                    'width', pm.width,
                    'height', pm.height,
                    'duration_secs', pm.duration_secs
                ) ORDER BY pm.sort_order ASC
            )
            FROM post_media pm
            WHERE pm.post_id = p.id
        ),
        '[]'::json
    ) AS media_attachments
FROM posts p
INNER JOIN users u ON u.id = p.user_id AND u.status = 'active'
LEFT JOIN user_profiles up ON up.user_id = u.id
LEFT JOIN post_stats ps ON ps.post_id = p.id
WHERE p.deleted_at IS NULL
  AND p.created_at < $2
  AND (
      -- Include own posts
      p.user_id = $1
      OR
      -- Include posts from active followings
      (
          p.user_id IN (
              SELECT following_id 
              FROM follows 
              WHERE follower_id = $1 AND status = 'active'
          )
          AND p.visibility IN ('public', 'followers_only')
      )
  )
  -- Exclude posts from blocked or blocking users
  AND NOT EXISTS (
      SELECT 1 FROM user_blocks ub
      WHERE (ub.blocker_id = $1 AND ub.blocked_id = p.user_id)
         OR (ub.blocker_id = p.user_id AND ub.blocked_id = $1)
  )
ORDER BY p.created_at DESC
LIMIT $3;


-- ----------------------------------------------------------------------------
-- Query 2: Trending & Discovery Feed (Engagement-Weighted Decay Algorithm)
-- Description: Calculates trending score using Hacker News style algorithm:
--              Score = (Likes*1 + Comments*3 + Shares*5) / (Age_in_Hours + 2)^1.5
-- Parameters:
--   :current_user_id (BIGINT) - Logged in user ID
--   :limit_val       (INT)    - Number of top trending posts
-- ----------------------------------------------------------------------------
PREPARE get_trending_feed(bigint, int) AS
SELECT
    p.id AS post_id,
    p.user_id AS author_id,
    u.username AS author_username,
    up.display_name AS author_display_name,
    up.avatar_url AS author_avatar_url,
    p.content,
    p.created_at,
    COALESCE(ps.likes_count, 0) AS likes_count,
    COALESCE(ps.comments_count, 0) AS comments_count,
    COALESCE(ps.shares_count, 0) AS shares_count,
    -- Mathematical Gravity Score Calculation
    ROUND(
        CAST(
            (COALESCE(ps.likes_count, 0) * 1.0 +
             COALESCE(ps.comments_count, 0) * 3.0 +
             COALESCE(ps.shares_count, 0) * 5.0)
            /
            POWER(
                EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 3600.0 + 2.0,
                1.5
            ) AS numeric
        ), 4
    ) AS trending_score
FROM posts p
INNER JOIN users u ON u.id = p.user_id AND u.status = 'active'
LEFT JOIN user_profiles up ON up.user_id = u.id
LEFT JOIN post_stats ps ON ps.post_id = p.id
WHERE p.deleted_at IS NULL
  AND p.visibility = 'public'
  AND p.created_at >= NOW() - INTERVAL '72 hours'
  -- Exclude blocked users
  AND NOT EXISTS (
      SELECT 1 FROM user_blocks ub
      WHERE (ub.blocker_id = $1 AND ub.blocked_id = p.user_id)
         OR (ub.blocker_id = p.user_id AND ub.blocked_id = $1)
  )
ORDER BY trending_score DESC, p.created_at DESC
LIMIT $2;


-- ----------------------------------------------------------------------------
-- Query 3: Hashtag Feed Query
-- Description: Fetches public posts associated with a specific hashtag.
-- Parameters:
--   :hashtag_name (VARCHAR) - Target hashtag without '#'
--   :limit_val    (INT)     - Pagination limit
-- ----------------------------------------------------------------------------
PREPARE get_hashtag_feed(varchar, int) AS
SELECT
    p.id AS post_id,
    u.username AS author_username,
    up.avatar_url AS author_avatar_url,
    p.content,
    p.created_at,
    COALESCE(ps.likes_count, 0) AS likes_count,
    COALESCE(ps.comments_count, 0) AS comments_count
FROM posts p
INNER JOIN post_hashtags ph ON ph.post_id = p.id
INNER JOIN hashtags h ON h.id = ph.hashtag_id
INNER JOIN users u ON u.id = p.user_id AND u.status = 'active'
LEFT JOIN user_profiles up ON up.user_id = u.id
LEFT JOIN post_stats ps ON ps.post_id = p.id
WHERE h.name = LOWER($1)
  AND p.deleted_at IS NULL
  AND p.visibility = 'public'
ORDER BY p.created_at DESC
LIMIT $2;
