-- ============================================================================
-- Component 8: Database Triggers & Stored Functions for Data Integrity & Counters
-- Target Database: PostgreSQL 14+
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1. Automatic Timestamp Update Trigger Function
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_timestamp_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply timestamp trigger to tables
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

DROP TRIGGER IF EXISTS trg_user_profiles_updated_at ON user_profiles;
CREATE TRIGGER trg_user_profiles_updated_at BEFORE UPDATE ON user_profiles FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

DROP TRIGGER IF EXISTS trg_posts_updated_at ON posts;
CREATE TRIGGER trg_posts_updated_at BEFORE UPDATE ON posts FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

DROP TRIGGER IF EXISTS trg_comments_updated_at ON comments;
CREATE TRIGGER trg_comments_updated_at BEFORE UPDATE ON comments FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

-- ----------------------------------------------------------------------------
-- 2. Follower / Following Counter Triggers (updates user_stats)
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_follow_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT' AND NEW.status = 'active') THEN
        INSERT INTO user_stats (user_id, followers_count, following_count)
            VALUES (NEW.following_id, 1, 0)
            ON CONFLICT (user_id) DO UPDATE SET followers_count = user_stats.followers_count + 1;

        INSERT INTO user_stats (user_id, followers_count, following_count)
            VALUES (NEW.follower_id, 0, 1)
            ON CONFLICT (user_id) DO UPDATE SET following_count = user_stats.following_count + 1;

    ELSIF (TG_OP = 'DELETE' AND OLD.status = 'active') THEN
        UPDATE user_stats SET followers_count = GREATEST(0, followers_count - 1) WHERE user_id = OLD.following_id;
        UPDATE user_stats SET following_count = GREATEST(0, following_count - 1) WHERE user_id = OLD.follower_id;

    ELSIF (TG_OP = 'UPDATE') THEN
        IF (OLD.status = 'pending' AND NEW.status = 'active') THEN
            UPDATE user_stats SET followers_count = followers_count + 1 WHERE user_id = NEW.following_id;
            UPDATE user_stats SET following_count = following_count + 1 WHERE user_id = NEW.follower_id;
        ELSIF (OLD.status = 'active' AND NEW.status != 'active') THEN
            UPDATE user_stats SET followers_count = GREATEST(0, followers_count - 1) WHERE user_id = NEW.following_id;
            UPDATE user_stats SET following_count = GREATEST(0, following_count - 1) WHERE user_id = NEW.follower_id;
        END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_follow_counts ON follows;
CREATE TRIGGER trg_follow_counts
AFTER INSERT OR UPDATE OR DELETE ON follows
FOR EACH ROW EXECUTE FUNCTION update_follow_counts();

-- ----------------------------------------------------------------------------
-- 3. Post Stats Counter Triggers (Likes & Comments)
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_post_likes_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT' AND NEW.reactable_type = 'post') THEN
        INSERT INTO post_stats (post_id, likes_count)
            VALUES (NEW.reactable_id, 1)
            ON CONFLICT (post_id) DO UPDATE SET likes_count = post_stats.likes_count + 1;
    ELSIF (TG_OP = 'DELETE' AND OLD.reactable_type = 'post') THEN
        UPDATE post_stats SET likes_count = GREATEST(0, likes_count - 1) WHERE post_id = OLD.reactable_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_post_likes_count ON reactions;
CREATE TRIGGER trg_post_likes_count
AFTER INSERT OR DELETE ON reactions
FOR EACH ROW EXECUTE FUNCTION update_post_likes_count();

CREATE OR REPLACE FUNCTION update_post_comments_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        INSERT INTO post_stats (post_id, comments_count)
            VALUES (NEW.post_id, 1)
            ON CONFLICT (post_id) DO UPDATE SET comments_count = post_stats.comments_count + 1;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE post_stats SET comments_count = GREATEST(0, comments_count - 1) WHERE post_id = OLD.post_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_post_comments_count ON comments;
CREATE TRIGGER trg_post_comments_count
AFTER INSERT OR DELETE ON comments
FOR EACH ROW EXECUTE FUNCTION update_post_comments_count();
