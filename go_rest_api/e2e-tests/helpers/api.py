import requests
from typing import Tuple, Dict, Any, Optional

class APIClient:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip("/")
        self.token: Optional[str] = None
        self.timeout = 10

    def set_token(self, token: str):
        self.token = token

    def health(self) -> int:
        """Check root health endpoint (outside /v1)."""
        url = self.base_url.rsplit("/v1", 1)[0] + "/healthz"
        resp = requests.get(url, timeout=self.timeout)
        return resp.status_code

    def _request(self, method: str, path: str, **kwargs) -> Tuple[int, Any]:
        url = f"{self.base_url}{path}"
        headers = kwargs.get("headers", {})
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        kwargs["headers"] = headers
        kwargs.setdefault("timeout", self.timeout)

        resp = requests.request(method, url, **kwargs)
        try:
            data = resp.json()
        except Exception:
            data = resp.text
        return resp.status_code, data

    # ── Auth ──
    def register(self, username, email, password) -> Tuple[int, Any]:
        return self._request("POST", "/auth/register", json={"username": username, "email": email, "password": password})
        
    def login(self, email, password) -> Tuple[int, Any]:
        return self._request("POST", "/auth/login", json={"email": email, "password": password})

    # ── Posts ──
    def create_post(self, title, body) -> Tuple[int, Any]:
        return self._request("POST", "/posts", json={"title": title, "body": body})
        
    def list_posts(self, limit=10, offset=0) -> Tuple[int, Any]:
        return self._request("GET", "/posts", params={"limit": limit, "offset": offset})
        
    def get_post(self, post_id) -> Tuple[int, Any]:
        return self._request("GET", f"/posts/{post_id}")
        
    def update_post(self, post_id, title, body) -> Tuple[int, Any]:
        return self._request("PUT", f"/posts/{post_id}", json={"title": title, "body": body})
        
    def delete_post(self, post_id) -> int:
        status, _ = self._request("DELETE", f"/posts/{post_id}")
        return status

    # ── Comments ──
    def create_comment(self, post_id, body) -> Tuple[int, Any]:
        return self._request("POST", f"/posts/{post_id}/comments", json={"body": body})
        
    def list_comments(self, post_id, limit=10, offset=0) -> Tuple[int, Any]:
        return self._request("GET", f"/posts/{post_id}/comments", params={"limit": limit, "offset": offset})
        
    def get_comment(self, comment_id) -> Tuple[int, Any]:
        # Usually comment endpoint is /posts/:post_id/comments/:id or /comments/:id
        # Assuming /comments/{comment_id} based on generic patterns, adjust if needed
        return self._request("GET", f"/comments/{comment_id}")
        
    def update_comment(self, comment_id, body) -> Tuple[int, Any]:
        return self._request("PUT", f"/comments/{comment_id}", json={"body": body})
        
    def delete_comment(self, comment_id) -> int:
        status, _ = self._request("DELETE", f"/comments/{comment_id}")
        return status
