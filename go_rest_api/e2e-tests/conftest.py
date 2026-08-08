import pytest
from uuid import uuid4
import time
from config import PG_URL, BASE_URL
from helpers.db import get_connection, truncate_test_data
from helpers.api import APIClient

@pytest.fixture(scope="session")
def db_conn():
    """Session-scoped DB connection. Closed after all tests."""
    conn = get_connection(PG_URL)
    yield conn
    conn.close()

@pytest.fixture(scope="session")
def api_client():
    """Session-scoped API client. Reusable across tests."""
    # Wait for API to be ready
    client = APIClient(BASE_URL)
    for _ in range(20):
        try:
            status = client.health()
            if status == 200:
                break
        except Exception:
            pass
        time.sleep(1)
    return client

@pytest.fixture
def auth_users(api_client):
    """
    Register 2 unique users, return {owner, other}.
    owner  = user A (pemilik resource)
    other  = user B (untuk test forbidden)
    Usernames pakai UUID suffix agar unik.
    A client sudah di-set token owner.
    """
    suffix_owner = uuid4().hex[:8]
    suffix_other = uuid4().hex[:8]
    
    # Register owner
    status, _ = api_client.register(f"owner_{suffix_owner}", f"owner_{suffix_owner}@example.com", "password")
    status, login_data = api_client.login(f"owner_{suffix_owner}@example.com", "password")
    # Handle possible different token keys based on swagger
    owner_token = login_data.get("token", login_data.get("access_token")) if isinstance(login_data, dict) else None
        
    # Register other
    status, _ = api_client.register(f"other_{suffix_other}", f"other_{suffix_other}@example.com", "password")
    status, login_data = api_client.login(f"other_{suffix_other}@example.com", "password")
    other_token = login_data.get("token", login_data.get("access_token")) if isinstance(login_data, dict) else None
        
    api_client.set_token(owner_token)
    return {"owner": owner_token, "other": other_token}

def pytest_sessionfinish(session, exitstatus):
    """Truncate test data after all tests run."""
    try:
        conn = get_connection(PG_URL)
        truncate_test_data(conn)
        conn.close()
    except Exception as e:
        print(f"Error truncating data: {e}")
