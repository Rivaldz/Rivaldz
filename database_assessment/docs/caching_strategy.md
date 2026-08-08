# Caching Strategy Proposal & Performance Architecture

## 1. Architectural Overview

To achieve sub-50ms feed retrieval and handle millions of concurrent read/write operations without saturating the primary relational database, the system adopts a **tiered hybrid caching architecture**. PostgreSQL serves as the persistent **Source of Truth**, while **Redis** operates as the in-memory data store for real-time state, session caches, counters, and pre-computed timelines.

```
                  ┌───────────────────────────────────────────┐
                  │               Client App                  │
                  └─────────────────────┬─────────────────────┘
                                        │
                                        ▼
                  ┌───────────────────────────────────────────┐
                  │             API Gateway / Service         │
                  └──────────────┬─────────────┬──────────────┘
                                 │             │
                    Cache Hit    │             │ Cache Miss
            ┌────────────────────┘             └────────────────────┐
            ▼                                                       ▼
  ┌───────────────────┐                                   ┌───────────────────┐
  │   Redis Cluster   │                                   │ PostgreSQL DB     │
  │ (In-Memory Data)  │                                   │ (Source of Truth) │
  └─────────┬─────────┘                                   └─────────┬─────────┘
            │                                                       │
            │             Async Write-Behind / Sync Write           │
            └───────────────────────────────────────────────────────┘
```

---

## 2. Redis Data Structure Mapping & TTL Matrix

| Domain Data | Redis Data Structure | Cache Pattern | Recommended TTL | Invalidation / Update Trigger |
| :--- | :--- | :--- | :--- | :--- |
| **User Profile Data** | `Hash` (`user:{id}:profile`) | Cache-Aside (Lazy) | 1 hour | Profile update event / DB write |
| **Active Follow List** | `Set` (`user:{id}:following`) | Cache-Aside | 30 minutes | Follow / Unfollow action |
| **User Timeline Feed** | `Sorted Set` (`feed:{id}`) | Write-Behind (Fan-Out) | 7 days (max 800 items)| New post created by following user |
| **Post Details & Media**| `Hash` (`post:{id}`) | Cache-Aside | 2 hours | Post update or soft delete |
| **Engagement Counters** | `Hash` (`post:{id}:counters`) | Write-Behind | Real-time in Redis | Periodic flush (every 30s) to DB |
| **Unread Notification Count**| `String` (`user:{id}:unread_notifs`)| Push / Invalidation | Permanent (Explicit Reset)| New notification created or marked read |
| **Global Trending Feed**| `Sorted Set` (`trending:global`)| Materialized / Scheduled| 5 minutes | Background worker refresh |

---

## 3. Feed Generation Strategy: Hybrid Fan-out Pattern

For activity feeds, pure **Fan-out-on-Read** creates slow join queries at high scale, whereas pure **Fan-out-on-Write** causes write amplification when celebrity accounts post (the "Celebrity Problem"). We employ a **Hybrid Fan-Out Strategy**:

```
                              [New Post Published]
                                       │
                         Is Author a Celebrity (>10k followers)?
                                      / \
                                YES  /   \  NO
                                    /     \
                                   ▼       ▼
               [Fan-out-on-Read Pathway]   [Fan-out-on-Write Pathway]
               Store post in DB only.     Push post_id into Redis Sorted Sets
               Fetch on demand when       of all followers:
               follower loads feed.       ZADD feed:{follower_id} <timestamp> <post_id>
```

### Feed Retrieval Workflow
When User A requests their home feed:
1. Query User A's Redis Sorted Set `feed:{User_A_ID}` via `ZREVRANGEBYSCORE`.
2. Fetch posts of any followed **celebrities** created since the cursor timestamp directly from Redis/DB.
3. Merge the lists in memory, sort by timestamp, and slice top 20 items.
4. If a cache miss occurs in Redis, fall back to PostgreSQL `queries/feed_generation.sql`, populate Redis asynchronously, and return the result.

---

## 4. Counter Write-Behind Pipeline (Likes, Views, Shares)

Writing directly to disk on every single "like" or "view" creates extreme row-lock contention and Write-Ahead Log (WAL) bloat in PostgreSQL.

### Solution: Redis Counter Aggregation + Batch DB Flush
1. **Write Action**: When a user likes post `1042`, execute atomic Redis command:
   ```redis
   HINCRBY post:1042:counters likes 1
   SADD dirty_posts 1042
   ```
2. **Read Action**: Read counts directly from Redis `HGETALL post:1042:counters` for instant feedback.
3. **Background Worker**: Every 30 seconds, a background process pops dirty IDs from `dirty_posts` and flushes accumulated delta counts to PostgreSQL in bulk:
   ```sql
   INSERT INTO post_stats (post_id, likes_count)
   VALUES (1042, 15), (1088, 3)
   ON CONFLICT (post_id) 
   DO UPDATE SET 
       likes_count = post_stats.likes_count + EXCLUDED.likes_count,
       updated_at = NOW();
   ```

---

## 5. PostgreSQL Materialized Views Integration

For complex analytical queries—such as leaderboard rankings, weekly active user statistics, or global trending topics—computing on live transactional tables is cost-prohibitive.

### Implementation: Materialized View for Trending Topics
```sql
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_trending_hashtags AS
SELECT 
    h.id AS hashtag_id,
    h.name AS hashtag_name,
    COUNT(ph.post_id) AS usage_count_72h,
    NOW() AS refreshed_at
FROM hashtags h
INNER JOIN post_hashtags ph ON ph.hashtag_id = h.id
INNER JOIN posts p ON p.id = ph.post_id
WHERE p.created_at >= NOW() - INTERVAL '72 hours'
  AND p.deleted_at IS NULL
GROUP BY h.id, h.name
ORDER BY usage_count_72h DESC
LIMIT 100;

-- Unique index required for non-blocking concurrent refresh
CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_trending_hashtags_id ON mv_trending_hashtags(hashtag_id);
```

**Refresh Schedule**:
Run every 15 minutes via cron or pg_cron without locking readers:
```sql
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_trending_hashtags;
```

---

## 6. Resilience & Protection Strategies

### A. Thundering Herd & Cache Stampede Protection
When a high-traffic cache key (e.g., global trending posts) expires, thousands of simultaneous incoming requests hit PostgreSQL at once.
- **Probabilistic Early Recomputation (XFetch Algorithm)** or **Distributed Mutex Lock**:
  When a cache miss occurs, only the first request acquires a Redis mutex lock (`SET key token NX EX 5`) to query PostgreSQL and repopulate Redis. Secondary requests wait 50ms and retry against Redis.
- **TTL Randomization (Jitter)**: Add a random variance of ±10% to all static TTL values (e.g., 300s + `random(0, 30)s`) to prevent synchronized cache mass-expiration.

### B. Cache Invalidation Patterns
- Use **Redis Pub/Sub** or **CDC (Change Data Capture)** via Postgres logical replication to broadcast events (e.g. `user_profile_updated`, `post_deleted`) to service nodes to purge local memory/Redis keys immediately.
