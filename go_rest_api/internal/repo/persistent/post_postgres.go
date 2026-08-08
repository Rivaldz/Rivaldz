package persistent

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/Rivaldz/my-template-go/internal/entity"
	"github.com/Rivaldz/my-template-go/internal/repo"
	"github.com/Rivaldz/my-template-go/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

// PostRepo -.
type PostRepo struct {
	*postgres.Postgres
}

// NewPostRepo -.
func NewPostRepo(pg *postgres.Postgres) *PostRepo {
	return &PostRepo{pg}
}

// Store -.
func (r *PostRepo) Store(ctx context.Context, post *entity.Post) error { //nolint:dupl // mirrors the standard repo insert pattern
	sql, args, err := r.Builder.
		Insert("posts").
		Columns("id, user_id, title, body, comment_count, created_at, updated_at").
		Values(post.ID, post.UserID, post.Title, post.Body, post.CommentCount, post.CreatedAt, post.UpdatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("PostRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PostRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

// GetByID -.
func (r *PostRepo) GetByID(ctx context.Context, postID string) (entity.Post, error) {
	sql, args, err := r.Builder.
		Select("id, user_id, title, body, comment_count, created_at, updated_at").
		From("posts").
		Where(sq.Eq{columnID: postID}).
		ToSql()
	if err != nil {
		return entity.Post{}, fmt.Errorf("PostRepo - GetByID - r.Builder: %w", err)
	}

	var post entity.Post

	err = r.Pool.QueryRow(ctx, sql, args...).
		Scan(&post.ID, &post.UserID, &post.Title, &post.Body, &post.CommentCount, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Post{}, entity.ErrPostNotFound
		}

		return entity.Post{}, fmt.Errorf("PostRepo - GetByID - r.Pool.QueryRow: %w", err)
	}

	return post, nil
}

// List -.
func (r *PostRepo) List(ctx context.Context, filter repo.PostFilter) ([]entity.Post, int, error) {
	countSQL, countArgs, err := r.Builder.
		Select("COUNT(*)").
		From("posts").
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("PostRepo - List - countBuilder: %w", err)
	}

	var total int

	err = r.Pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("PostRepo - List - count query: %w", err)
	}

	dataSQL, dataArgs, err := r.Builder.
		Select("id, user_id, title, body, comment_count, created_at, updated_at").
		From("posts").
		OrderBy("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("PostRepo - List - dataBuilder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("PostRepo - List - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	posts := make([]entity.Post, 0, filter.Limit)

	for rows.Next() {
		var p entity.Post

		err = rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Body, &p.CommentCount, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("PostRepo - List - rows.Scan: %w", err)
		}

		posts = append(posts, p)
	}

	return posts, total, nil
}

// Update -.
func (r *PostRepo) Update(ctx context.Context, post *entity.Post) error {
	sql, args, err := r.Builder.
		Update("posts").
		Set("title", post.Title).
		Set("body", post.Body).
		Set("updated_at", post.UpdatedAt).
		Where(sq.Eq{columnID: post.ID, columnUserID: post.UserID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("PostRepo - Update - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PostRepo - Update - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrPostNotFound
	}

	return nil
}

// Delete removes a post and all of its comments in a single transaction.
func (r *PostRepo) Delete(ctx context.Context, postID, userID string) error {
	err := r.Transaction(ctx, func(tx pgx.Tx) error {
		sql, args, err := r.Builder.
			Delete("comments").
			Where(sq.Eq{"post_id": postID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("PostRepo - Delete - comments r.Builder: %w", err)
		}

		if _, err = tx.Exec(ctx, sql, args...); err != nil {
			return fmt.Errorf("PostRepo - Delete - tx.Exec comments: %w", err)
		}

		sql, args, err = r.Builder.
			Delete("posts").
			Where(sq.Eq{columnID: postID, columnUserID: userID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("PostRepo - Delete - posts r.Builder: %w", err)
		}

		result, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("PostRepo - Delete - tx.Exec posts: %w", err)
		}

		if result.RowsAffected() == 0 {
			return entity.ErrPostNotFound
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
