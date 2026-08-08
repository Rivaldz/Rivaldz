import requests
import psycopg2
import sys

base_url = "http://localhost:8080/v1"
db_dsn = "postgresql://user:myAwEsOm3pa55%40w0rd@localhost:5432/db"

passed = 0
total = 0

def expect(actual, expected, label):
    global passed, total
    total += 1
    if actual == expected:
        print(f"[PASS] {label}")
        passed += 1
    else:
        print(f"[FAIL] {label} -> expected {expected}, got {actual}")

def expect_true(condition, label):
    global passed, total
    total += 1
    if condition:
        print(f"[PASS] {label}")
        passed += 1
    else:
        print(f"[FAIL] {label}")

def run_tests():
    global passed, total

    print("Running E2E tests...")

    try:
        conn = psycopg2.connect(db_dsn)
        conn.autocommit = True
        cur = conn.cursor()
    except Exception as e:
        print(f"Failed to connect to DB: {e}")
        sys.exit(1)

    # 1.1 Health check
    try:
        r = requests.get("http://localhost:8080/healthz")
        expect(r.status_code, 200, "healthz status")
    except Exception as e:
        print(f"[FAIL] healthz status -> {e}")
        sys.exit(1)

    # 1.2 Register user A
    r = requests.post(f"{base_url}/auth/register", json={"username":"e2e_user_a","email":"e2e_a@test.com","password":"testpass123"})
    expect(r.status_code, 201, "register user A")
    user_a_id = r.json().get("id") if r.status_code == 201 else None

    # 1.3 Register user B
    r = requests.post(f"{base_url}/auth/register", json={"username":"e2e_user_b","email":"e2e_b@test.com","password":"testpass123"})
    expect(r.status_code, 201, "register user B")
    
    # 1.4 Register duplicate email
    r = requests.post(f"{base_url}/auth/register", json={"username":"e2e_user_a2","email":"e2e_a@test.com","password":"testpass123"})
    expect(r.status_code, 409, "register duplicate")

    # 1.5 Login user A
    r = requests.post(f"{base_url}/auth/login", json={"email":"e2e_a@test.com","password":"testpass123"})
    expect(r.status_code, 200, "login user A")
    token_a = r.json().get("token") if r.status_code == 200 else ""
    headers_a = {"Authorization": f"Bearer {token_a}"}

    # 1.6 Login user B
    r = requests.post(f"{base_url}/auth/login", json={"email":"e2e_b@test.com","password":"testpass123"})
    expect(r.status_code, 200, "login user B")
    token_b = r.json().get("token") if r.status_code == 200 else ""
    headers_b = {"Authorization": f"Bearer {token_b}"}

    # 1.7 Login wrong pass
    r = requests.post(f"{base_url}/auth/login", json={"email":"e2e_a@test.com","password":"wrongpass"})
    expect(r.status_code, 401, "login wrong pass")

    # 2.1 Create post
    r = requests.post(f"{base_url}/posts/", headers=headers_a, json={"title":"Post 1","body":"Body post pertama"})
    expect(r.status_code, 201, "create post 1")
    post_1_id = r.json().get("id") if r.status_code == 201 else None

    if post_1_id:
        cur.execute("SELECT user_id, title, comment_count FROM posts WHERE id = %s", (post_1_id,))
        row = cur.fetchone()
        expect_true(row is not None and str(row[0]) == str(user_a_id) and row[1] == "Post 1" and row[2] == 0, "DB verify post 1")

    # 2.2 Create post 2
    r = requests.post(f"{base_url}/posts/", headers=headers_a, json={"title":"Post 2","body":"Body post kedua"})
    expect(r.status_code, 201, "create post 2")
    post_2_id = r.json().get("id") if r.status_code == 201 else None

    # 2.3 List posts
    r = requests.get(f"{base_url}/posts/?limit=10&offset=0", headers=headers_a)
    expect(r.status_code, 200, "list posts")
    if r.status_code == 200:
        expect_true(r.json().get("total", 0) >= 2, "list posts total >= 2")

    # 2.4 Get post
    if post_1_id:
        r = requests.get(f"{base_url}/posts/{post_1_id}", headers=headers_a)
        expect(r.status_code, 200, "get post 1")
        if r.status_code == 200:
            expect_true(r.json().get("title") == "Post 1", "get post 1 title match")

    # 2.5 Update post
    if post_1_id:
        r = requests.put(f"{base_url}/posts/{post_1_id}", headers=headers_a, json={"title":"Post 1 Updated","body":"Body baru"})
        expect(r.status_code, 200, "update post 1")

    # 2.6 Delete post 1
    if post_1_id:
        r = requests.delete(f"{base_url}/posts/{post_1_id}", headers=headers_a)
        expect(r.status_code, 204, "delete post 1")
        cur.execute("SELECT id FROM posts WHERE id = %s", (post_1_id,))
        expect_true(cur.fetchone() is None, "DB verify post 1 deleted")
        cur.execute("SELECT id FROM comments WHERE post_id = %s", (post_1_id,))
        expect_true(cur.fetchone() is None, "DB verify comments for post 1 deleted")

    # 3.1 Create comment 1
    comment_1_id = None
    if post_2_id:
        r = requests.post(f"{base_url}/posts/{post_2_id}/comments", headers=headers_a, json={"body":"Komentar 1"})
        expect(r.status_code, 201, "create comment 1")
        comment_1_id = r.json().get("id") if r.status_code == 201 else None
        if comment_1_id:
            cur.execute("SELECT comment_count FROM posts WHERE id = %s", (post_2_id,))
            row = cur.fetchone()
            expect_true(row is not None and row[0] == 1, "DB verify post 2 comment_count = 1")

    # 3.2 Create comment 2
    comment_2_id = None
    if post_2_id:
        r = requests.post(f"{base_url}/posts/{post_2_id}/comments", headers=headers_a, json={"body":"Komentar 2"})
        expect(r.status_code, 201, "create comment 2")
        comment_2_id = r.json().get("id") if r.status_code == 201 else None

    # 3.3 List comments
    if post_2_id:
        r = requests.get(f"{base_url}/posts/{post_2_id}/comments?limit=10&offset=0", headers=headers_a)
        expect(r.status_code, 200, "list comments")
        if r.status_code == 200:
            expect_true(r.json().get("total") == 2, "list comments total == 2")

    # 3.4 Get comment
    if post_2_id and comment_1_id:
        r = requests.get(f"{base_url}/comments/{comment_1_id}", headers=headers_a)
        expect(r.status_code, 200, "get comment 1")

    # 3.5 Update comment
    if post_2_id and comment_1_id:
        r = requests.put(f"{base_url}/comments/{comment_1_id}", headers=headers_a, json={"body":"Komentar 1 Updated"})
        expect(r.status_code, 200, "update comment 1")

    # 3.6 Delete comment
    if post_2_id and comment_1_id:
        r = requests.delete(f"{base_url}/comments/{comment_1_id}", headers=headers_a)
        expect(r.status_code, 204, "delete comment 1")
        cur.execute("SELECT comment_count FROM posts WHERE id = %s", (post_2_id,))
        row = cur.fetchone()
        expect_true(row is not None and row[0] == 1, "DB verify post 2 comment_count = 1 after delete")

    # 4.1 Update post non-owner
    if post_2_id:
        r = requests.put(f"{base_url}/posts/{post_2_id}", headers=headers_b, json={"title":"Hack","body":"Hack"})
        expect(r.status_code, 403, "update post non-owner")

    # 4.2 Delete post non-owner
    if post_2_id:
        r = requests.delete(f"{base_url}/posts/{post_2_id}", headers=headers_b)
        expect(r.status_code, 403, "delete post non-owner")

    # 4.3 Update comment non-owner
    if post_2_id and comment_2_id:
        r = requests.put(f"{base_url}/comments/{comment_2_id}", headers=headers_b, json={"body":"Hack"})
        expect(r.status_code, 403, "update comment non-owner")

    # 4.4 Delete comment non-owner
    if post_2_id and comment_2_id:
        r = requests.delete(f"{base_url}/comments/{comment_2_id}", headers=headers_b)
        expect(r.status_code, 403, "delete comment non-owner")

    # 5.1 Create post without title
    r = requests.post(f"{base_url}/posts/", headers=headers_a, json={"body":"tanpa title"})
    expect(r.status_code, 400, "create post without title")

    # 5.2 Create post with title > 255
    r = requests.post(f"{base_url}/posts/", headers=headers_a, json={"title":"A"*300, "body":"ok"})
    expect(r.status_code, 400, "create post with title > 255")

    # 5.3 Create post without body
    r = requests.post(f"{base_url}/posts/", headers=headers_a, json={"title":"Tanpa body"})
    expect(r.status_code, 400, "create post without body")

    # 5.4 Create comment without body
    if post_2_id:
        r = requests.post(f"{base_url}/posts/{post_2_id}/comments", headers=headers_a, json={})
        expect(r.status_code, 400, "create comment without body")

    # 5.5 Create comment body > 2000
    if post_2_id:
        r = requests.post(f"{base_url}/posts/{post_2_id}/comments", headers=headers_a, json={"body":"B"*3000})
        expect(r.status_code, 400, "create comment with body > 2000")

    # 5.6 Get post 404
    r = requests.get(f"{base_url}/posts/00000000-0000-0000-0000-000000000000", headers=headers_a)
    expect(r.status_code, 404, "get post 404")

    # 5.7 Update post 404
    r = requests.put(f"{base_url}/posts/00000000-0000-0000-0000-000000000000", headers=headers_a, json={"title":"x","body":"y"})
    expect(r.status_code, 404, "update post 404")

    # 5.8 Delete post 404
    r = requests.delete(f"{base_url}/posts/00000000-0000-0000-0000-000000000000", headers=headers_a)
    expect(r.status_code, 404, "delete post 404")

    # 5.9 Comment on 404 post
    r = requests.post(f"{base_url}/posts/00000000-0000-0000-0000-000000000000/comments", headers=headers_a, json={"body":"orphan"})
    expect(r.status_code, 404, "comment on 404 post")

    # 5.10 Get comment 404
    r = requests.get(f"{base_url}/comments/00000000-0000-0000-0000-000000000000", headers=headers_a)
    expect(r.status_code, 404, "get comment 404")

    # 6.1 Request without token
    r = requests.get(f"{base_url}/posts/")
    expect(r.status_code, 401, "request without token")

    # 6.2 Request wrong format
    r = requests.get(f"{base_url}/posts/", headers={"Authorization": "Basic dXNlcjpwYXNz"})
    expect(r.status_code, 401, "request with Basic token")

    # 6.3 Request invalid token
    r = requests.get(f"{base_url}/posts/", headers={"Authorization": "Bearer invalid.jwt.token"})
    expect(r.status_code, 401, "request with invalid token")

    # 7 DB checks
    cur.execute("SELECT c.id FROM comments c LEFT JOIN posts p ON c.post_id = p.id WHERE p.id IS NULL;")
    expect_true(len(cur.fetchall()) == 0, "DB check orphan comments")

    cur.execute("SELECT p.id, p.comment_count, COUNT(c.id) AS actual_count FROM posts p LEFT JOIN comments c ON c.post_id = p.id GROUP BY p.id HAVING p.comment_count != COUNT(c.id);")
    expect_true(len(cur.fetchall()) == 0, "DB check comment_count consistency")

    cur.execute("SELECT p.id FROM posts p LEFT JOIN users u ON p.user_id = u.id WHERE u.id IS NULL;")
    expect_true(len(cur.fetchall()) == 0, "DB check post user valid")

    cur.execute("SELECT c.id FROM comments c LEFT JOIN users u ON c.user_id = u.id WHERE u.id IS NULL;")
    expect_true(len(cur.fetchall()) == 0, "DB check comment user valid")

    # Cleanup
    cur.execute("DELETE FROM users WHERE username LIKE 'e2e_%';")

    print(f"\nResult: {passed}/{total} PASSED")
    if passed == total:
        sys.exit(0)
    else:
        sys.exit(1)

if __name__ == '__main__':
    run_tests()
