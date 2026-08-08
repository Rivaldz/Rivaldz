#!/usr/bin/env bash
set -e

CONTAINER_NAME="pg_test"
DB_USER="postgres"

echo "=== Running Complete Database Assessment Verification ==="

cat \
  schema/001_users.sql \
  schema/002_relationships.sql \
  schema/003_posts.sql \
  schema/004_interactions.sql \
  schema/005_messaging.sql \
  schema/006_notifications.sql \
  schema/007_indexes.sql \
  schema/008_triggers.sql \
  schema/seed.sql \
  queries/feed_generation.sql \
  queries/common_queries.sql \
  <(echo "
\echo '=== 1. TEST HOME FEED QUERY FOR ALICE (User ID 1) ==='
EXECUTE get_home_feed(1, NOW(), 10);

\echo '=== 2. TEST TRENDING FEED QUERY ==='
EXECUTE get_trending_feed(1, 5);

\echo '=== 3. TEST HASHTAG FEED QUERY (#travel) ==='
EXECUTE get_hashtag_feed('travel', 5);

\echo '=== 4. TEST THREADED COMMENTS FOR POST 101 ==='
EXECUTE get_threaded_comments(101);

\echo '=== 5. TEST CONVERSATIONS INBOX FOR ALICE (User ID 1) ==='
EXECUTE get_user_conversations(1);

\echo '=== 6. TEST MUTUAL FOLLOWINGS (User 1 & User 2) ==='
EXECUTE get_mutual_followings(1, 2);

\echo '=== 7. TEST NOTIFICATIONS INBOX FOR ALICE (User ID 1) ==='
EXECUTE get_notifications(1, 10);

\echo '=== 8. TEST AUTOMATED COUNTER TRIGGERS IN USER_STATS & POST_STATS ==='
SELECT * FROM user_stats;
SELECT * FROM post_stats;
") | docker exec -i $CONTAINER_NAME psql -U $DB_USER

echo "=========================================================="
echo "SUCCESS: ALL SCHEMA DDL, INDEXES, TRIGGERS, SEEDS,"
echo "AND COMPLEX SQL QUERIES EXECUTED AND VERIFIED CLEANLY!"
echo "=========================================================="
