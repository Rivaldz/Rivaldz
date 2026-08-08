import os
import glob

def build_prompt() -> str:
    # Read swagger
    swagger_path = os.path.join(os.path.dirname(__file__), "..", "..", "docs", "swagger.json")
    swagger_content = "No swagger found"
    if os.path.exists(swagger_path):
        with open(swagger_path, "r") as f:
            swagger_content = f.read()
    else:
        # Check alternative locations
        swagger_path = os.path.join(os.path.dirname(__file__), "..", "..", "swagger.json")
        if os.path.exists(swagger_path):
            with open(swagger_path, "r") as f:
                swagger_content = f.read()

    # Read migrations
    migrations_dir = os.path.join(os.path.dirname(__file__), "..", "..", "migrations")
    migrations_content = "No migrations found"
    if os.path.exists(migrations_dir):
        sql_files = glob.glob(os.path.join(migrations_dir, "*.up.sql")) + glob.glob(os.path.join(migrations_dir, "*.sql"))
        migrations_content = ""
        for sql_file in sql_files:
            if not sql_file.endswith(".down.sql"):
                with open(sql_file, "r") as f:
                    migrations_content += f"--- {os.path.basename(sql_file)} ---\n"
                    migrations_content += f.read() + "\n\n"

    # Read helper signatures
    helpers_content = """
    ## Available Helpers
    APIClient methods:
    - register(username, email, password) -> (status, data)
    - login(email, password) -> (status, data)
    - create_post(title, body) -> (status, data)
    - list_posts(limit=10, offset=0) -> (status, data)
    - get_post(post_id) -> (status, data)
    - update_post(post_id, title, body) -> (status, data)
    - delete_post(post_id) -> status
    - create_comment(post_id, body) -> (status, data)
    - list_comments(post_id, limit=10, offset=0) -> (status, data)
    - get_comment(comment_id) -> (status, data)
    - update_comment(comment_id, body) -> (status, data)
    - delete_comment(comment_id) -> status
    - set_token(token) -> None

    DB helper methods:
    - get_user_by_id(conn, user_id) -> dict
    - get_post_by_id(conn, post_id) -> dict
    - get_comment_by_id(conn, comment_id) -> dict
    - get_comments_for_post(conn, post_id) -> list[dict]
    - count_comments(conn, post_id) -> int
    - post_exists(conn, post_id) -> bool
    - comment_exists(conn, comment_id) -> bool
    """

    prompt = f"""
## API Endpoints & Schemas (Swagger)
{swagger_content}

## Database Schema (Migrations)
{migrations_content}

{helpers_content}

## Task
Generate pytest test functions covering all CRUD operations, authorization,
validation, error cases, and database state verification for the blog API.
Include at least these test scenarios:
- Register success, duplicate
- Login success
- Posts: CRUD happy path, pagination, list
- Posts: Authorization (forbidden on update/delete for other users)
- Posts: Validation & Error (missing fields, max lengths, not found)
- Comments: CRUD happy path, list
- Comments: Count verification in DB and API
- Comments: Authorization (forbidden on update/delete for other users)
- Comments: Validation & Error (missing fields, max lengths, post not found)
- Unauthenticated access tests
- DB State Verification (transaction cascade on delete)
"""
    return prompt
