package comment

import (
	"context"
	"fmt"
	"time"

	"github.com/Rivaldz/my-template-go/internal/entity"
	"github.com/Rivaldz/my-template-go/internal/repo"
	"github.com/google/uuid"
)

// UseCase -.
type UseCase struct {
	repo repo.CommentRepo
}

// New -.
func New(r repo.CommentRepo) *UseCase {
	return &UseCase{repo: r}
}

// Create -.
func (uc *UseCase) Create(ctx context.Context, userID, postID, body string) (entity.Comment, error) {
	now := time.Now().UTC()

	comment := entity.Comment{
		ID:        uuid.New().String(),
		PostID:    postID,
		UserID:    userID,
		Body:      body,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := uc.repo.Store(ctx, &comment)
	if err != nil {
		return entity.Comment{}, fmt.Errorf("CommentUseCase - Create - uc.repo.Store: %w", err)
	}

	return comment, nil
}

// Get -.
func (uc *UseCase) Get(ctx context.Context, commentID string) (entity.Comment, error) {
	comment, err := uc.repo.GetByID(ctx, commentID)
	if err != nil {
		return entity.Comment{}, fmt.Errorf("CommentUseCase - Get - uc.repo.GetByID: %w", err)
	}

	return comment, nil
}

// ListByPost -.
func (uc *UseCase) ListByPost(ctx context.Context, postID string, limit, offset int) ([]entity.Comment, int, error) {
	if limit <= 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	comments, total, err := uc.repo.ListByPost(ctx, postID, repo.CommentFilter{
		Limit:  uint64(limit),
		Offset: uint64(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("CommentUseCase - ListByPost - uc.repo.ListByPost: %w", err)
	}

	return comments, total, nil
}

// Update -.
func (uc *UseCase) Update(ctx context.Context, userID, commentID, body string) (entity.Comment, error) {
	now := time.Now().UTC()

	comment, err := uc.repo.GetByID(ctx, commentID)
	if err != nil {
		return entity.Comment{}, fmt.Errorf("CommentUseCase - Update - uc.repo.GetByID: %w", err)
	}

	if comment.UserID != userID {
		return entity.Comment{}, entity.ErrCommentForbidden
	}

	comment.Body = body
	comment.UpdatedAt = now

	err = uc.repo.Update(ctx, &comment)
	if err != nil {
		return entity.Comment{}, fmt.Errorf("CommentUseCase - Update - uc.repo.Update: %w", err)
	}

	return comment, nil
}

// Delete -.
func (uc *UseCase) Delete(ctx context.Context, userID, commentID string) error {
	comment, err := uc.repo.GetByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("CommentUseCase - Delete - uc.repo.GetByID: %w", err)
	}

	if comment.UserID != userID {
		return entity.ErrCommentForbidden
	}

	err = uc.repo.Delete(ctx, commentID, userID)
	if err != nil {
		return fmt.Errorf("CommentUseCase - Delete - uc.repo.Delete: %w", err)
	}

	return nil
}
