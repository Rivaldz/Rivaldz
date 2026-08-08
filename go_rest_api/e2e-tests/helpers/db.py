import psycopg2
from psycopg2.extras import RealDictCursor

def get_connection(dsn: str):
    conn = psycopg2.connect(dsn)
    conn.autocommit = True
    return conn

def insert_user_raw(conn, uid, username, email, password_hash) -> None:
    with conn.cursor() as cur:
        cur.execute(
            "INSERT INTO users (id, username, email, password_hash) VALUES (%s, %s, %s, %s)",
            (uid, username, email, password_hash)
        )

def delete_user(conn, user_id) -> None:
    with conn.cursor() as cur:
        cur.execute("DELETE FROM users WHERE id = %s", (user_id,))

def get_user_by_id(conn, user_id) -> dict:
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        cur.execute("SELECT * FROM users WHERE id = %s", (user_id,))
        return cur.fetchone()

def get_post_by_id(conn, post_id) -> dict:
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        cur.execute("SELECT * FROM posts WHERE id = %s", (post_id,))
        return cur.fetchone()

def get_comment_by_id(conn, comment_id) -> dict:
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        cur.execute("SELECT * FROM comments WHERE id = %s", (comment_id,))
        return cur.fetchone()

def get_comments_for_post(conn, post_id) -> list:
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        cur.execute("SELECT * FROM comments WHERE post_id = %s ORDER BY created_at ASC", (post_id,))
        return cur.fetchall()

def count_comments(conn, post_id) -> int:
    with conn.cursor() as cur:
        cur.execute("SELECT COUNT(*) FROM comments WHERE post_id = %s", (post_id,))
        return cur.fetchone()[0]

def post_exists(conn, post_id) -> bool:
    with conn.cursor() as cur:
        cur.execute("SELECT 1 FROM posts WHERE id = %s", (post_id,))
        return cur.fetchone() is not None

def comment_exists(conn, comment_id) -> bool:
    with conn.cursor() as cur:
        cur.execute("SELECT 1 FROM comments WHERE id = %s", (comment_id,))
        return cur.fetchone() is not None

def truncate_test_data(conn) -> None:
    with conn.cursor() as cur:
        cur.execute("DELETE FROM comments; DELETE FROM posts; DELETE FROM users;")
