-- ============================================================================
-- Seed Dataset for Testing Schema and Complex SQL Queries
-- Target Database: PostgreSQL 14+
-- ============================================================================

-- 1. Insert Sample Users
INSERT INTO users (id, username, email, password_hash, status, is_verified) VALUES
(1, 'alice', 'alice@example.com', '$2a$12$e8.Z/samplehash1', 'active', true),
(2, 'bob', 'bob@example.com', '$2a$12$e8.Z/samplehash2', 'active', true),
(3, 'charlie', 'charlie@example.com', '$2a$12$e8.Z/samplehash3', 'active', false),
(4, 'diana', 'diana@example.com', '$2a$12$e8.Z/samplehash4', 'active', true),
(5, 'eve_celebrity', 'eve@example.com', '$2a$12$e8.Z/samplehash5', 'active', true);

SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));

-- 2. Insert User Profiles
INSERT INTO user_profiles (user_id, display_name, bio, avatar_url, privacy_level) VALUES
(1, 'Alice Smith', 'Software Architect & Tech Enthusiast', 'https://cdn.example.com/avatars/alice.png', 'public'),
(2, 'Bob Jones', 'Digital Nomad & Photographer', 'https://cdn.example.com/avatars/bob.png', 'public'),
(3, 'Charlie Brown', 'Coffee Lover', 'https://cdn.example.com/avatars/charlie.png', 'public'),
(4, 'Diana Prince', 'Cybersecurity Specialist', 'https://cdn.example.com/avatars/diana.png', 'public'),
(5, 'Eve Celebrity', 'Global Influencer with Millions of Fans', 'https://cdn.example.com/avatars/eve.png', 'public');

-- 3. Insert Follow Relationships
INSERT INTO follows (follower_id, following_id, status) VALUES
(1, 2, 'active'),  -- Alice follows Bob
(1, 3, 'active'),  -- Alice follows Charlie
(1, 5, 'active'),  -- Alice follows Eve
(2, 1, 'active'),  -- Bob follows Alice
(3, 1, 'active'),  -- Charlie follows Alice
(4, 1, 'active');  -- Diana follows Alice

-- 4. Insert Posts & Media
INSERT INTO posts (id, user_id, content, visibility, created_at) VALUES
(101, 2, 'Exploring the Swiss Alps today! #travel #nature', 'public', NOW() - INTERVAL '2 hours'),
(102, 3, 'Just baked fresh sourdough bread! #foodie', 'public', NOW() - INTERVAL '5 hours'),
(103, 5, 'Excited to announce my new tech book launch! #tech #book', 'public', NOW() - INTERVAL '1 hour'),
(104, 1, 'Designing normalized PostgreSQL schemas is fun. #database #sql', 'public', NOW() - INTERVAL '30 minutes');

SELECT setval('posts_id_seq', (SELECT MAX(id) FROM posts));

INSERT INTO post_media (post_id, media_type, url, width, height, sort_order) VALUES
(101, 'image', 'https://cdn.example.com/media/alps1.jpg', 1920, 1080, 1),
(101, 'image', 'https://cdn.example.com/media/alps2.jpg', 1920, 1080, 2),
(102, 'image', 'https://cdn.example.com/media/bread.jpg', 1080, 1080, 1);

-- 5. Insert Hashtags
INSERT INTO hashtags (id, name) VALUES
(1, 'travel'), (2, 'nature'), (3, 'foodie'), (4, 'tech'), (5, 'book'), (6, 'database'), (7, 'sql');

SELECT setval('hashtags_id_seq', (SELECT MAX(id) FROM hashtags));

INSERT INTO post_hashtags (post_id, hashtag_id) VALUES
(101, 1), (101, 2),
(102, 3),
(103, 4), (103, 5),
(104, 6), (104, 7);

-- 6. Insert Comments (Threaded)
INSERT INTO comments (id, post_id, user_id, parent_id, content, depth, created_at) VALUES
(1, 101, 1, NULL, 'Stunning views, Bob!', 0, NOW() - INTERVAL '90 minutes'),
(2, 101, 2, 1, 'Thanks Alice! Highly recommend visiting.', 1, NOW() - INTERVAL '80 minutes'),
(3, 104, 2, NULL, 'Great schema post! Partial indexes are awesome.', 0, NOW() - INTERVAL '20 minutes');

SELECT setval('comments_id_seq', (SELECT MAX(id) FROM comments));

-- 7. Insert Reactions
INSERT INTO reactions (user_id, reactable_type, reactable_id, reaction_type) VALUES
(1, 'post', 101, 'love'),
(3, 'post', 101, 'like'),
(4, 'post', 101, 'like'),
(1, 'post', 103, 'like'),
(2, 'post', 104, 'like');

-- 8. Insert Conversations & Messages
INSERT INTO conversations (id, title, is_group, created_by) VALUES
(1, NULL, false, 1),
(2, 'Tech Team Chat', true, 1);

SELECT setval('conversations_id_seq', (SELECT MAX(id) FROM conversations));

INSERT INTO conversation_participants (conversation_id, user_id, role, last_read_at) VALUES
(1, 1, 'member', NOW() - INTERVAL '10 minutes'),
(1, 2, 'member', NOW() - INTERVAL '1 hour'),
(2, 1, 'admin', NOW()),
(2, 2, 'member', NOW()),
(2, 4, 'member', NOW());

INSERT INTO messages (id, conversation_id, sender_id, message_type, content, created_at) VALUES
(1, 1, 2, 'text', 'Hey Alice, are you free for a call?', NOW() - INTERVAL '25 minutes'),
(2, 1, 1, 'text', 'Sure Bob, calling you now!', NOW() - INTERVAL '10 minutes');

SELECT setval('messages_id_seq', (SELECT MAX(id) FROM messages));

-- 9. Insert Notifications
INSERT INTO notifications (recipient_id, actor_id, notification_type, target_type, target_id, message) VALUES
(2, 1, 'post_like', 'post', 101, 'Alice loved your post.'),
(1, 2, 'comment_reply', 'comment', 1, 'Bob replied to your comment.');
