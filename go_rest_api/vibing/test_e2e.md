# Plan: E2E Testing Blog API (Direct Hit Endpoint)

**Target:** Blog API di codebase ini (Go / Fiber / PostgreSQL)
**Stack:** Docker (db + app di `localhost:8080`)
**DB:** PostgreSQL di `localhost:5432`, user=`user`, pass=`myAwEsOm3pa55@w0rd`, db=`db`
**Metode:** Langsung hit URL/endpoint via HTTP, verifikasi response, lalu verifikasi state DB

---

## 0. Prasyarat & Setup

```bash
# 1. Pastikan stack jalan
docker compose -f docker-compose.yml up -d db app
docker compose ps   # db: Up, app: Up

# 2. Health check
curl http://localhost:8080/healthz
# → expect: OK

# 3. (Opsional) Reset DB agar state bersih
docker compose -f docker-compose.yml down -v
docker compose -f docker-compose.yml up -d db app
```

---

## 1. Tahap 1 — Health Check & Auth

### 1.1 Health check
```bash
curl -s http://localhost:8080/healthz
# Assert: OK
```

### 1.2 Register user baru (success)
```bash
curl -s -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"e2e_user_a","email":"e2e_a@test.com","password":"testpass123"}'
# Assert: 201, {"id":"<uuid>","username":"e2e_user_a","email":"e2e_a@test.com",...}
# Simpan id → USER_A_ID
```

### 1.3 Register user kedua (untuk test authorization)
```bash
curl -s -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"e2e_user_b","email":"e2e_b@test.com","password":"testpass123"}'
# Assert: 201, id UUID → USER_B_ID
```

### 1.4 Register duplicate email (error case)
```bash
curl -s -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"e2e_user_a2","email":"e2e_a@test.com","password":"testpass123"}'
# Assert: 409, {"error":"user already exists"}
```

### 1.5 Login user A → dapat token
```bash
curl -s -X POST http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"e2e_a@test.com","password":"testpass123"}'
# Assert: 200, {"token":"<jwt>"} → TOKEN_A
```

### 1.6 Login user B → dapat token
```bash
curl -s -X POST http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"e2e_b@test.com","password":"testpass123"}'
# Assert: 200, {"token":"<jwt>"} → TOKEN_B
```

### 1.7 Login salah password (error case)
```bash
curl -s -X POST http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"e2e_a@test.com","password":"wrongpass"}'
# Assert: 401, {"error":"invalid credentials"}
```

**Variabel yang disimpan:**
```
USER_A_ID, USER_B_ID, TOKEN_A (Bearer), TOKEN_B (Bearer)
```

---

## 2. Tahap 2 — Posts CRUD (Happy Path)

### 2.1 Create post (user A)
```bash
curl -s -X POST http://localhost:8080/v1/posts/ \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"title":"Post 1","body":"Body post pertama"}'
# Assert: 201, {"id":"<uuid>","user_id":"<USER_A_ID>","title":"Post 1",
#               "body":"Body post pertama","comment_count":0,...}
# Simpan id → POST_1_ID
```

**Verifikasi DB:**
```sql
SELECT id, user_id, title, body, comment_count
FROM posts WHERE id = '<POST_1_ID>';
-- Assert: 1 row, user_id = USER_A_ID, comment_count = 0, title/body match
```

### 2.2 Create post kedua (untuk pagination)
```bash
curl -s -X POST http://localhost:8080/v1/posts/ \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"title":"Post 2","body":"Body post kedua"}'
# Assert: 201 → POST_2_ID
```

### 2.3 List posts (pagination)
```bash
curl -s 'http://localhost:8080/v1/posts/?limit=10&offset=0' \
  -H 'Authorization: Bearer <TOKEN_A>'
# Assert: 200, {"posts":[...],"total":2}
# posts[0] harus post terbaru (created_at DESC)
```

### 2.4 Get post by ID
```bash
curl -s http://localhost:8080/v1/posts/<POST_1_ID> \
  -H 'Authorization: Bearer <TOKEN_A>'
# Assert: 200, title="Post 1", user_id=USER_A_ID
```

### 2.5 Update post (owner = user A)
```bash
curl -s -X PUT http://localhost:8080/v1/posts/<POST_1_ID> \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"title":"Post 1 Updated","body":"Body baru"}'
# Assert: 200, title="Post 1 Updated", body="Body baru"
```

### 2.6 Delete post (owner = user A) + cascade comments
```bash
curl -s -X DELETE http://localhost:8080/v1/posts/<POST_1_ID> \
  -H 'Authorization: Bearer <TOKEN_A>'
# Assert: 204 No Content
```

**Verifikasi DB (cascade delete):**
```sql
SELECT id FROM posts WHERE id = '<POST_1_ID>';
-- Assert: 0 rows (post terhapus)

SELECT id FROM comments WHERE post_id = '<POST_1_ID>';
-- Assert: 0 rows (semua comment ikut terhapus)
```

---

## 3. Tahap 3 — Comments CRUD (Happy Path)

### 3.1 Create comment di post
```bash
curl -s -X POST http://localhost:8080/v1/posts/<POST_2_ID>/comments \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"body":"Komentar 1"}'
# Assert: 201, {"id":"<uuid>","post_id":"<POST_2_ID>","user_id":"<USER_A_ID>",
#               "body":"Komentar 1",...}
# Simpan id → COMMENT_1_ID
```

**Verifikasi DB (comment_count increment):**
```sql
SELECT comment_count FROM posts WHERE id = '<POST_2_ID>';
-- Assert: 1

SELECT id, post_id, user_id, body FROM comments WHERE id = '<COMMENT_1_ID>';
-- Assert: 1 row, body="Komentar 1", post_id=POST_2_ID, user_id=USER_A_ID
```

### 3.2 Create comment kedua
```bash
curl -s -X POST http://localhost:8080/v1/posts/<POST_2_ID>/comments \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"body":"Komentar 2"}'
# Assert: 201 → COMMENT_2_ID
```

### 3.3 List comments di post
```bash
curl -s 'http://localhost:8080/v1/posts/<POST_2_ID>/comments?limit=10&offset=0' \
  -H 'Authorization: Bearer <TOKEN_A>'
# Assert: 200, {"comments":[...],"total":2}
# Urut: created_at ASC (komentar 1 dulu, lalu 2)
```

### 3.4 Get comment by ID
```bash
curl -s http://localhost:8080/v1/comments/<COMMENT_1_ID> \
  -H 'Authorization: Bearer <TOKEN_A>'
# Assert: 200, body="Komentar 1", post_id=POST_2_ID, user_id=USER_A_ID
```

### 3.5 Update comment (owner = user A)
```bash
curl -s -X PUT http://localhost:8080/v1/comments/<COMMENT_1_ID> \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"body":"Komentar 1 Updated"}'
# Assert: 200, body="Komentar 1 Updated"
```

### 3.6 Delete comment (owner = user A)
```bash
curl -s -X DELETE http://localhost:8080/v1/comments/<COMMENT_1_ID> \
  -H 'Authorization: Bearer <TOKEN_A>'
# Assert: 204 No Content
```

**Verifikasi DB (comment_count decrement):**
```sql
SELECT comment_count FROM posts WHERE id = '<POST_2_ID>';
-- Assert: 1 (setelah delete 1 dari 2 komentar)

SELECT id FROM comments WHERE id = '<COMMENT_1_ID>';
-- Assert: 0 rows
```

---

## 4. Tahap 4 — Authorization (403 Forbidden)

### 4.1 Update post oleh non-owner
```bash
# User B mencoba update post milik user A
curl -s -X PUT http://localhost:8080/v1/posts/<POST_2_ID> \
  -H 'Authorization: Bearer <TOKEN_B>' -H 'Content-Type: application/json' \
  -d '{"title":"Hack","body":"Hack"}'
# Assert: 403, {"error":"forbidden"}
```

### 4.2 Delete post oleh non-owner
```bash
curl -s -X DELETE http://localhost:8080/v1/posts/<POST_2_ID> \
  -H 'Authorization: Bearer <TOKEN_B>'
# Assert: 403, {"error":"forbidden"}
```

### 4.3 Update comment oleh non-owner
```bash
# User B mencoba update comment milik user A
curl -s -X PUT http://localhost:8080/v1/comments/<COMMENT_2_ID> \
  -H 'Authorization: Bearer <TOKEN_B>' -H 'Content-Type: application/json' \
  -d '{"body":"Hack comment"}'
# Assert: 403, {"error":"forbidden"}
```

### 4.4 Delete comment oleh non-owner
```bash
curl -s -X DELETE http://localhost:8080/v1/comments/<COMMENT_2_ID> \
  -H 'Authorization: Bearer <TOKEN_B>'
# Assert: 403, {"error":"forbidden"}
```

**Verifikasi DB (resource tidak berubah):**
```sql
SELECT title, body FROM posts WHERE id = '<POST_2_ID>';
-- Assert: tetap "Post 2" / "Body post kedua" (tidak ter-update oleh user B)

SELECT id FROM comments WHERE id = '<COMMENT_2_ID>';
-- Assert: masih ada (tidak terhapus oleh user B)
```

---

## 5. Tahap 5 — Validasi Input (400) & Resource Not Found (404)

### 5.1 Create post tanpa title
```bash
curl -s -X POST http://localhost:8080/v1/posts/ \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"body":"tanpa title"}'
# Assert: 400, {"error":"invalid request body"}
```

### 5.2 Create post dengan title > 255 karakter
```bash
curl -s -X POST http://localhost:8080/v1/posts/ \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"title":"AAAA... (300x A)","body":"ok"}'
# Assert: 400
```

### 5.3 Create post tanpa body
```bash
curl -s -X POST http://localhost:8080/v1/posts/ \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"title":"Tanpa body"}'
# Assert: 400
```

### 5.4 Create comment tanpa body
```bash
curl -s -X POST http://localhost:8080/v1/posts/<POST_2_ID>/comments \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{}'
# Assert: 400
```

### 5.5 Create comment body > 2000 karakter
```bash
curl -s -X POST http://localhost:8080/v1/posts/<POST_2_ID>/comments \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"body":"B... (3000x B)"}'
# Assert: 400
```

### 5.6 Get post tidak ada → 404
```bash
curl -s http://localhost:8080/v1/posts/00000000-0000-0000-0000-000000000000 \
  -H 'Authorization: Bearer <TOKEN_A>'
# Assert: 404, {"error":"post not found"}
```

### 5.7 Update post tidak ada → 404
```bash
curl -s -X PUT http://localhost:8080/v1/posts/00000000-0000-0000-0000-000000000000 \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"title":"x","body":"y"}'
# Assert: 404
```

### 5.8 Delete post tidak ada → 404
```bash
curl -s -X DELETE http://localhost:8080/v1/posts/00000000-0000-0000-0000-000000000000 \
  -H 'Authorization: Bearer <TOKEN_A>'
# Assert: 404
```

### 5.9 Comment di post tidak ada → 404
```bash
curl -s -X POST http://localhost:8080/v1/posts/00000000-0000-0000-0000-000000000000/comments \
  -H 'Authorization: Bearer <TOKEN_A>' -H 'Content-Type: application/json' \
  -d '{"body":"orphan"}'
# Assert: 404, {"error":"post not found"}
```

### 5.10 Get comment tidak ada → 404
```bash
curl -s http://localhost:8080/v1/comments/00000000-0000-0000-0000-000000000000 \
  -H 'Authorization: Bearer <TOKEN_A>'
# Assert: 404, {"error":"comment not found"}
```

---

## 6. Tahap 6 — Unauthenticated (401)

### 6.1 Request tanpa token
```bash
curl -s http://localhost:8080/v1/posts/ \
  -H 'Content-Type: application/json'
# Assert: 401, {"error":"missing authorization header"}
```

### 6.2 Request dengan format header salah
```bash
curl -s http://localhost:8080/v1/posts/ \
  -H 'Authorization: Basic dXNlcjpwYXNz'
# Assert: 401, {"error":"invalid authorization header format"}
```

### 6.3 Request dengan token invalid
```bash
curl -s http://localhost:8080/v1/posts/ \
  -H 'Authorization: Bearer invalid.jwt.token'
# Assert: 401, {"error":"invalid or expired token"}
```

---

## 7. Tahap 7 — Verifikasi Integritas Database

Setelah semua tahap di atas, lakukan query verifikasi akhir:

### 7.1 Tidak ada orphan comments
```sql
SELECT c.id
FROM comments c
LEFT JOIN posts p ON c.post_id = p.id
WHERE p.id IS NULL;
-- Assert: 0 rows
```

### 7.2 Denormalized `comment_count` selalu akurat
```sql
SELECT p.id, p.comment_count, COUNT(c.id) AS actual_count
FROM posts p
LEFT JOIN comments c ON c.post_id = p.id
GROUP BY p.id
HAVING p.comment_count != COUNT(c.id);
-- Assert: 0 rows (tidak ada mismatch)
```

### 7.3 Soft constraint: semua post & comment punya user valid
```sql
SELECT p.id FROM posts p LEFT JOIN users u ON p.user_id = u.id WHERE u.id IS NULL;
-- Assert: 0 rows

SELECT c.id FROM comments c LEFT JOIN users u ON c.user_id = u.id WHERE u.id IS NULL;
-- Assert: 0 rows
```

### 7.4 Pagination konsisten dengan total
```sql
SELECT COUNT(*) FROM posts;
SELECT COUNT(*) FROM comments;
-- Bandingkan dengan nilai "total" dari GET /posts & GET /posts/:id/comments
```

---

## 8. Ringkasan Matriks Assert

| # | Tahap | Endpoint | Status Assert | DB Assert |
|---|-------|----------|---------------|-----------|
| 1 | Health | GET /healthz | 200 "OK" | – |
| 2 | Auth | POST /auth/register | 201 | user row ada |
| 3 | Auth | POST /auth/register (dup) | 409 | – |
| 4 | Auth | POST /auth/login | 200 token | – |
| 5 | Auth | POST /auth/login (salah pass) | 401 | – |
| 6 | Posts | POST /posts | 201 | comment_count=0 |
| 7 | Posts | GET /posts | 200 total>=1 | COUNT(*) match |
| 8 | Posts | GET /posts/:id | 200 | row match |
| 9 | Posts | PUT /posts/:id (owner) | 200 | title updated |
| 10 | Posts | DELETE /posts/:id (owner) | 204 | post + comments hilang |
| 11 | Comments | POST /posts/:pid/comments | 201 | comment_count+1 |
| 12 | Comments | GET /posts/:pid/comments | 200 total=2 | COUNT(*) match |
| 13 | Comments | GET /comments/:id | 200 | row match |
| 14 | Comments | PUT /comments/:id (owner) | 200 | body updated |
| 15 | Comments | DELETE /comments/:id (owner) | 204 | comment_count-1 |
| 16 | Authz | PUT /posts/:id (non-owner) | 403 | row tidak berubah |
| 17 | Authz | DELETE /posts/:id (non-owner) | 403 | row masih ada |
| 18 | Authz | PUT /comments/:id (non-owner) | 403 | row tidak berubah |
| 19 | Authz | DELETE /comments/:id (non-owner) | 403 | row masih ada |
| 20 | Validasi | POST /posts tanpa title | 400 | – |
| 21 | Validasi | POST /posts title>255 | 400 | – |
| 22 | Validasi | POST /posts tanpa body | 400 | – |
| 23 | Validasi | POST comments tanpa body | 400 | – |
| 24 | Validasi | POST comments body>2000 | 400 | – |
| 25 | 404 | GET/PUT/DELETE /posts/:id hilang | 404 | – |
| 26 | 404 | POST /posts/:pid/comments (post hilang) | 404 | – |
| 27 | 404 | GET /comments/:id hilang | 404 | – |
| 28 | 401 | Semua protected tanpa token | 401 | – |
| 29 | 401 | Token invalid | 401 | – |
| 30 | DB | Orphan check | – | 0 rows |
| 31 | DB | comment_count konsisten | – | 0 mismatch |

---

## 9. Cara Eksekusi

### Opsi A — Manual curl (copy-paste command di atas)
Cocok untuk debugging satu per satu.

### Opsi B — One-shot Python script (disarankan)
Buat `e2e-tests/smoke.py` yang menjalankan semua skenario di atas otomatis:

```python
# Pseudostruktur smoke.py
# 1. import requests, psycopg2
# 2. helper: expect(actual, expected, label) → print PASS/FAIL, track count
# 3. Tahap 1-7 dijalankan berurutan
# 4. Akhir: print "X/Y PASSED" dan exit code 0/1
```

Output:
```
[PASS] healthz → 200 OK
[PASS] register e2e_user_a → 201
[PASS] register duplicate → 409
...
[FAIL] PUT /posts/:id non-owner → expected 403, got 200
...
Result: 29/30 PASSED
```

### Opsi C — Jalankan Go integration test yang sudah ada
```bash
docker compose -f docker-compose.yml -f docker-compose-integration-test.yml down -v
docker compose -f docker-compose.yml -f docker-compose-integration-test.yml \
  up --build --abort-on-container-exit --exit-code-from integration-test
```

---

## 10. Cleanup

```bash
# Hapus test user dari DB
psql "postgres://user:myAwEsOm3pa55@w0rd@localhost:5432/db" \
  -c "DELETE FROM users WHERE username LIKE 'e2e_%';"

# Atau reset penuh
docker compose -f docker-compose.yml down -v
```
