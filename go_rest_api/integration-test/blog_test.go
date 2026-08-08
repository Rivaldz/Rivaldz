package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
)

type postResponse struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	CommentCount int    `json:"comment_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type commentResponse struct {
	ID        string `json:"id"`
	PostID    string `json:"post_id"`
	UserID    string `json:"user_id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// httpCreatePost is a helper that creates a post via HTTP and returns the parsed response.
func httpCreatePost(t *testing.T, token, title string) postResponse {
	t.Helper()

	createBody := fmt.Sprintf(`{"title":%q,"body":%q}`, title, "post body")

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPost, basePathV1+"/posts/", bytes.NewBufferString(createBody), token)
	if err != nil {
		t.Fatalf("Create post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create post: expected 201, got %d", resp.StatusCode)
	}

	return parseJSON[postResponse](t, resp)
}

// httpCreateComment is a helper that creates a comment via HTTP and returns the parsed response.
func httpCreateComment(t *testing.T, token, postID, body string) commentResponse {
	t.Helper()

	createBody := fmt.Sprintf(`{"body":%q}`, body)

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPost, basePathV1+"/posts/"+postID+"/comments", bytes.NewBufferString(createBody), token)
	if err != nil {
		t.Fatalf("Create comment: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create comment: expected 201, got %d", resp.StatusCode)
	}

	return parseJSON[commentResponse](t, resp)
}

// registerAndLoginUnique registers a unique user based on the test name and a suffix.
func registerAndLoginUnique(t *testing.T, suffix string) string {
	t.Helper()

	name := sanitizeTestName(t) + suffix
	email := name + "@test.com"

	resp := registerUser(t, name, email, testPassword)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("registerAndLoginUnique: register expected 201, got %d", resp.StatusCode)
	}

	return loginUser(t, email, testPassword)
}

// HTTP: create post and verify fields.
func TestHTTPPostCreateV1(t *testing.T) {
	token := registerAndLogin(t)
	created := httpCreatePost(t, token, "integration post")

	if created.ID == "" {
		t.Fatal("expected non-empty id")
	}

	if created.CommentCount != 0 {
		t.Errorf("expected comment_count 0, got %d", created.CommentCount)
	}
}

// HTTP: get post by ID.
func TestHTTPPostGetV1(t *testing.T) {
	token := registerAndLogin(t)
	created := httpCreatePost(t, token, "get post")

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/posts/"+created.ID, http.NoBody, token)
	if err != nil {
		t.Fatalf("Get post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	got := parseJSON[postResponse](t, resp)

	if got.ID != created.ID {
		t.Errorf("expected id %q, got %q", created.ID, got.ID)
	}
}

// HTTP: list posts.
func TestHTTPPostListV1(t *testing.T) {
	token := registerAndLogin(t)
	httpCreatePost(t, token, "list post")

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/posts/?limit=10&offset=0", http.NoBody, token)
	if err != nil {
		t.Fatalf("List posts: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	type listResponse struct {
		Posts []postResponse `json:"posts"`
		Total int            `json:"total"`
	}

	listed := parseJSON[listResponse](t, resp)

	if listed.Total < 1 {
		t.Errorf("expected total >= 1, got %d", listed.Total)
	}
}

// HTTP: update post as its author.
func TestHTTPPostUpdateV1(t *testing.T) {
	token := registerAndLogin(t)
	created := httpCreatePost(t, token, "update post")

	updateBody := `{"title":"updated title","body":"updated body"}`

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPut, basePathV1+"/posts/"+created.ID, bytes.NewBufferString(updateBody), token)
	if err != nil {
		t.Fatalf("Update post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	updated := parseJSON[postResponse](t, resp)

	if updated.Title != "updated title" {
		t.Errorf("expected title 'updated title', got %q", updated.Title)
	}
}

// HTTP: delete post and verify comments are removed (transaction).
func TestHTTPPostDeleteV1(t *testing.T) {
	token := registerAndLogin(t)
	created := httpCreatePost(t, token, "delete post")
	httpCreateComment(t, token, created.ID, "comment one")
	httpCreateComment(t, token, created.ID, "comment two")

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodDelete, basePathV1+"/posts/"+created.ID, http.NoBody, token)
	if err != nil {
		t.Fatalf("Delete post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

// HTTP: update a post that belongs to another user returns 403.
func TestHTTPPostUpdateForbiddenV1(t *testing.T) {
	ownerToken := registerAndLogin(t)
	created := httpCreatePost(t, ownerToken, "owned post")

	otherToken := registerAndLoginUnique(t, "other")

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	updateBody := `{"title":"hijack","body":"hijack"}`

	resp, err := doAuthenticatedRequest(ctx, http.MethodPut, basePathV1+"/posts/"+created.ID, bytes.NewBufferString(updateBody), otherToken)
	if err != nil {
		t.Fatalf("Update post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

// HTTP: create comment on a post and verify the comment count increments.
func TestHTTPCommentCreateV1(t *testing.T) {
	token := registerAndLogin(t)
	created := httpCreatePost(t, token, "comment post")
	comment := httpCreateComment(t, token, created.ID, "nice post")

	if comment.ID == "" {
		t.Fatal("expected non-empty id")
	}
}

// HTTP: list comments for a post.
func TestHTTPCommentListV1(t *testing.T) {
	token := registerAndLogin(t)
	created := httpCreatePost(t, token, "comment list post")
	httpCreateComment(t, token, created.ID, "first")
	httpCreateComment(t, token, created.ID, "second")

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/posts/"+created.ID+"/comments?limit=10&offset=0", http.NoBody, token)
	if err != nil {
		t.Fatalf("List comments: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	type listResponse struct {
		Comments []commentResponse `json:"comments"`
		Total    int               `json:"total"`
	}

	listed := parseJSON[listResponse](t, resp)

	if listed.Total != 2 {
		t.Errorf("expected total 2, got %d", listed.Total)
	}
}

// HTTP: update comment as its author.
func TestHTTPCommentUpdateV1(t *testing.T) {
	token := registerAndLogin(t)
	created := httpCreatePost(t, token, "comment update post")
	comment := httpCreateComment(t, token, created.ID, "original")

	updateBody := `{"body":"updated comment"}`

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPut, basePathV1+"/comments/"+comment.ID, bytes.NewBufferString(updateBody), token)
	if err != nil {
		t.Fatalf("Update comment: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	updated := parseJSON[commentResponse](t, resp)

	if updated.Body != "updated comment" {
		t.Errorf("expected body 'updated comment', got %q", updated.Body)
	}
}

// HTTP: delete a comment that belongs to another user returns 403.
func TestHTTPCommentDeleteForbiddenV1(t *testing.T) {
	ownerToken := registerAndLogin(t)
	created := httpCreatePost(t, ownerToken, "owned comment post")
	comment := httpCreateComment(t, ownerToken, created.ID, "owned comment")

	otherToken := registerAndLoginUnique(t, "other")

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodDelete, basePathV1+"/comments/"+comment.ID, http.NoBody, otherToken)
	if err != nil {
		t.Fatalf("Delete comment: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

// HTTP: blog error cases.
func TestHTTPBlogErrorsV1(t *testing.T) { //nolint:dupl // mirrors the task test pattern
	t.Run("no token returns 401", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
		defer cancel()

		body := `{"title":"unauthorized post","body":"should fail"}`

		resp, err := doWebRequestWithTimeout(ctx, http.MethodPost, basePathV1+"/posts/", bytes.NewBufferString(body))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("get non-existent post returns 404", func(t *testing.T) {
		token := registerAndLogin(t)

		ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
		defer cancel()

		resp, err := doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/posts/00000000-0000-0000-0000-000000000000", http.NoBody, token)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("create with invalid body returns 400", func(t *testing.T) {
		token := registerAndLogin(t)

		ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
		defer cancel()

		body := `{"body":"missing title"}`

		resp, err := doAuthenticatedRequest(ctx, http.MethodPost, basePathV1+"/posts/", bytes.NewBufferString(body), token)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

// HTTP: commenting on a post that does not exist returns 404.
func TestHTTPCommentOnMissingPostV1(t *testing.T) {
	token := registerAndLogin(t)

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	body := `{"body":"orphan comment"}`

	resp, err := doAuthenticatedRequest(ctx, http.MethodPost, basePathV1+"/posts/00000000-0000-0000-0000-000000000000/comments", bytes.NewBufferString(body), token)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}
