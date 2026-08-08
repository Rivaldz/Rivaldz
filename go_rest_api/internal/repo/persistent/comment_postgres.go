package persistent

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/Rivaldz/my-template-go/internal/entity"
	"github.com/Rivaldz/my-template-go/internal/repo"
	"github.com/Rivaldz/my-template-go/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// CommentRepo -.
type CommentRepo struct {
	*postgres.Postgres
}

// NewCommentRepo -.
func NewCommentRepo(pg *postgres.Postgres) *CommentRepo {
	return &CommentRepo{pg}
}

// Store inserts a comment and increments the post comment count atomically.
func (r *CommentRepo) Store(ctx context.Context, comment *entity.Comment) error {
	err := r.Transaction(ctx, func(tx pgx.Tx) error {
		sql, args, err := r.Builder.
			Insert("comments").
			Columns("id, post_id, user_id, body, created_at, updated_at").
			Values(comment.ID, comment.PostID, comment.UserID, comment.Body, comment.CreatedAt, comment.UpdatedAt).
			ToSql()
		if err != nil {
			return fmt.Errorf("CommentRepo - Store - r.Builder: %w", err)
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return entity.ErrPostNotFound
			}

			return fmt.Errorf("CommentRepo - Store - tx.Exec: %w", err)
		}

		sql, args, err = r.Builder.
			Update("posts").
			Set("comment_count", sq.Expr("comment_count + 1")).
			Set("updated_at", time.Now().UTC()).
			Where(sq.Eq{columnID: comment.PostID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("CommentRepo - Store - count r.Builder: %w", err)
		}

		if _, err = tx.Exec(ctx, sql, args...); err != nil {
			return fmt.Errorf("CommentRepo - Store - count tx.Exec: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetByID -.
func (r *CommentRepo) GetByID(ctx context.Context, commentID string) (entity.Comment, error) { //nolint:dupl // mirrors the standard repo lookup pattern
	sql, args, err := r.Builder.
		Select("id, post_id, user_id, body, created_at, updated_at").
		From("comments").
		Where(sq.Eq{columnID: commentID}).
		ToSql()
	if err != nil {
		return entity.Comment{}, fmt.Errorf("CommentRepo - GetByID - r.Builder: %w", err)
	}

	var comment entity.Comment

	err = r.Pool.QueryRow(ctx, sql, args...).
		Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Comment{}, entity.ErrCommentNotFound
		}

		return entity.Comment{}, fmt.Errorf("CommentRepo - GetByID - r.Pool.QueryRow: %w", err)
	}

	return comment, nil
}

// ListByPost -.
func (r *CommentRepo) ListByPost(ctx context.Context, postID string, filter repo.CommentFilter) ([]entity.Comment, int, error) {
	countSQL, countArgs, err := r.Builder.
		Select("COUNT(*)").
		From("comments").
		Where(sq.Eq{"post_id": postID}).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("CommentRepo - ListByPost - countBuilder: %w", err)
	}

	var total int

	err = r.Pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("CommentRepo - ListByPost - count query: %w", err)
	}

	dataSQL, dataArgs, err := r.Builder.
		Select("id, post_id, user_id, body, created_at, updated_at").
		From("comments").
		Where(sq.Eq{"post_id": postID}).
		OrderBy("created_at ASC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("CommentRepo - ListByPost - dataBuilder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("CommentRepo - ListByPost - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	comments := make([]entity.Comment, 0, filter.Limit)

	for rows.Next() {
		var c entity.Comment

		err = rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Body, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("CommentRepo - ListByPost - rows.Scan: %w", err)
		}

		comments = append(comments, c)
	}

	return comments, total, nil
}

// Update -.
func (r *CommentRepo) Update(ctx context.Context, comment *entity.Comment) error {
	sql, args, err := r.Builder.
		Update("comments").
		Set("body", comment.Body).
		Set("updated_at", comment.UpdatedAt).
		Where(sq.Eq{columnID: comment.ID, columnUserID: comment.UserID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("CommentRepo - Update - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("CommentRepo - Update - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrCommentNotFound
	}

	return nil
}

// Delete removes a comment and decrements the post comment count atomically.
func (r *CommentRepo) Delete(ctx context.Context, commentID, userID string) error {
	err := r.Transaction(ctx, func(tx pgx.Tx) error {
		return r.deleteCommentAndDecrement(ctx, tx, commentID, userID)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *CommentRepo) deleteCommentAndDecrement(ctx context.Context, tx pgx.Tx, commentID, userID string) error {
	postID, err := r.commentPostID(ctx, tx, commentID)
	if err != nil {
		return err
	}

	sql, args, err := r.Builder.
		Delete("comments").
		Where(sq.Eq{columnID: commentID, columnUserID: userID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("CommentRepo - deleteCommentAndDecrement - r.Builder: %w", err)
	}

	result, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("CommentRepo - deleteCommentAndDecrement - tx.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrCommentNotFound
	}

	sql, args, err = r.Builder.
		Update("posts").
		Set("comment_count", sq.Expr("GREATEST(comment_count - 1, 0)")).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{columnID: postID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("CommentRepo - deleteCommentAndDecrement - count r.Builder: %w", err)
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("CommentRepo - deleteCommentAndDecrement - count tx.Exec: %w", err)
	}

	return nil
}

func (r *CommentRepo) commentPostID(ctx context.Context, tx pgx.Tx, commentID string) (string, error) {
	sql, args, err := r.Builder.
		Select("post_id").
		From("comments").
		Where(sq.Eq{columnID: commentID}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("CommentRepo - commentPostID - r.Builder: %w", err)
	}

	var postID string

	err = tx.QueryRow(ctx, sql, args...).Scan(&postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", entity.ErrCommentNotFound
		}

		return "", fmt.Errorf("CommentRepo - commentPostID - tx.QueryRow: %w", err)
	}

	return postID, nil
}
