package post

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
	repo repo.PostRepo
}

// New -.
func New(r repo.PostRepo) *UseCase {
	return &UseCase{repo: r}
}

// Create -.
func (uc *UseCase) Create(ctx context.Context, userID, title, body string) (entity.Post, error) {
	now := time.Now().UTC()

	post := entity.Post{
		ID:           uuid.New().String(),
		UserID:       userID,
		Title:        title,
		Body:         body,
		CommentCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := uc.repo.Store(ctx, &post)
	if err != nil {
		return entity.Post{}, fmt.Errorf("PostUseCase - Create - uc.repo.Store: %w", err)
	}

	return post, nil
}

// Get -.
func (uc *UseCase) Get(ctx context.Context, postID string) (entity.Post, error) {
	post, err := uc.repo.GetByID(ctx, postID)
	if err != nil {
		return entity.Post{}, fmt.Errorf("PostUseCase - Get - uc.repo.GetByID: %w", err)
	}

	return post, nil
}

// List -.
func (uc *UseCase) List(ctx context.Context, limit, offset int) ([]entity.Post, int, error) {
	if limit <= 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	posts, total, err := uc.repo.List(ctx, repo.PostFilter{
		Limit:  uint64(limit),
		Offset: uint64(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("PostUseCase - List - uc.repo.List: %w", err)
	}

	return posts, total, nil
}

// Update -.
func (uc *UseCase) Update(ctx context.Context, userID, postID, title, body string) (entity.Post, error) {
	now := time.Now().UTC()

	post, err := uc.repo.GetByID(ctx, postID)
	if err != nil {
		return entity.Post{}, fmt.Errorf("PostUseCase - Update - uc.repo.GetByID: %w", err)
	}

	if post.UserID != userID {
		return entity.Post{}, entity.ErrPostForbidden
	}

	post.Title = title
	post.Body = body
	post.UpdatedAt = now

	err = uc.repo.Update(ctx, &post)
	if err != nil {
		return entity.Post{}, fmt.Errorf("PostUseCase - Update - uc.repo.Update: %w", err)
	}

	return post, nil
}

// Delete -.
func (uc *UseCase) Delete(ctx context.Context, userID, postID string) error {
	post, err := uc.repo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("PostUseCase - Delete - uc.repo.GetByID: %w", err)
	}

	if post.UserID != userID {
		return entity.ErrPostForbidden
	}

	err = uc.repo.Delete(ctx, postID, userID)
	if err != nil {
		return fmt.Errorf("PostUseCase - Delete - uc.repo.Delete: %w", err)
	}

	return nil
}
