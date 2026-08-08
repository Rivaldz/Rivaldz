-- ============================================================================
-- Component 1: Users, Profiles, and User Stats
-- Target Database: PostgreSQL 14+
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Enum Types
DO $$ BEGIN
    CREATE TYPE user_status AS ENUM ('active', 'suspended', 'deactivated', 'banned');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE privacy_level AS ENUM ('public', 'private', 'friends_only');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 1. Core Users Table (Authentication & Identity)
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(30) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    phone_number    VARCHAR(20),
    status          user_status NOT NULL DEFAULT 'active',
    is_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ,

    CONSTRAINT uq_users_username UNIQUE (username),
    CONSTRAINT uq_users_email UNIQUE (email),
    CONSTRAINT chk_username_length CHECK (char_length(username) >= 3),
    CONSTRAINT chk_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- 2. User Profiles Table (1:1 Profile Information)
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id         BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name    VARCHAR(100),
    bio             TEXT,
    avatar_url      VARCHAR(500),
    cover_photo_url VARCHAR(500),
    website_url     VARCHAR(500),
    location        VARCHAR(100),
    date_of_birth   DATE,
    privacy_level   privacy_level NOT NULL DEFAULT 'public',
    metadata        JSONB DEFAULT '{}'::jsonb,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. User Statistics Table (Denormalized Counters for O(1) Reads)
CREATE TABLE IF NOT EXISTS user_stats (
    user_id         BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    followers_count BIGINT NOT NULL DEFAULT 0,
    following_count BIGINT NOT NULL DEFAULT 0,
    posts_count     BIGINT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_positive_user_counts CHECK (
        followers_count >= 0 AND following_count >= 0 AND posts_count >= 0
    )
);
