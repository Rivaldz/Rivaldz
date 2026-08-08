// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/Rivaldz/my-template-go/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// TranslationRepo -.
	TranslationRepo interface {
		Store(ctx context.Context, userID string, t entity.Translation) error
		GetHistory(ctx context.Context, userID string) ([]entity.Translation, error)
	}

	// TranslationWebAPI -.
	TranslationWebAPI interface {
		Translate(ctx context.Context, t entity.Translation) (entity.Translation, error)
	}

	// UserRepo -.
	UserRepo interface {
		Store(ctx context.Context, user *entity.User) error
		GetByID(ctx context.Context, id string) (entity.User, error)
		GetByEmail(ctx context.Context, email string) (entity.User, error)
	}

	// TaskRepo -.
	TaskRepo interface {
		Store(ctx context.Context, task *entity.Task) error
		GetByID(ctx context.Context, userID, taskID string) (entity.Task, error)
		List(ctx context.Context, userID string, filter TaskFilter) ([]entity.Task, int, error)
		Update(ctx context.Context, task *entity.Task) error
		Delete(ctx context.Context, userID, taskID string) error
	}

	// TaskFilter -.
	TaskFilter struct {
		Status *entity.TaskStatus
		Limit  uint64
		Offset uint64
	}

	// PostRepo -.
	PostRepo interface {
		Store(ctx context.Context, post *entity.Post) error
		GetByID(ctx context.Context, postID string) (entity.Post, error)
		List(ctx context.Context, filter PostFilter) ([]entity.Post, int, error)
		Update(ctx context.Context, post *entity.Post) error
		Delete(ctx context.Context, postID, userID string) error
	}

	// PostFilter -.
	PostFilter struct {
		Limit  uint64
		Offset uint64
	}

	// CommentRepo -.
	CommentRepo interface {
		Store(ctx context.Context, comment *entity.Comment) error
		GetByID(ctx context.Context, commentID string) (entity.Comment, error)
		ListByPost(ctx context.Context, postID string, filter CommentFilter) ([]entity.Comment, int, error)
		Update(ctx context.Context, comment *entity.Comment) error
		Delete(ctx context.Context, commentID, userID string) error
	}

	// CommentFilter -.
	CommentFilter struct {
		Limit  uint64
		Offset uint64
	}
)
