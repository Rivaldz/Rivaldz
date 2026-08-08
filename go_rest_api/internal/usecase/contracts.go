// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/Rivaldz/my-template-go/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Translation -.
	Translation interface {
		Translate(ctx context.Context, userID string, t entity.Translation) (entity.Translation, error)
		History(ctx context.Context, userID string) (entity.TranslationHistory, error)
	}

	// User -.
	User interface {
		Register(ctx context.Context, username, email, password string) (entity.User, error)
		Login(ctx context.Context, email, password string) (string, error)
		GetUser(ctx context.Context, userID string) (entity.User, error)
	}

	// Task -.
	Task interface {
		Create(ctx context.Context, userID, title, description string) (entity.Task, error)
		Get(ctx context.Context, userID, taskID string) (entity.Task, error)
		List(ctx context.Context, userID string, status *entity.TaskStatus, limit, offset int) ([]entity.Task, int, error)
		Update(ctx context.Context, userID, taskID, title, description string) (entity.Task, error)
		Transition(ctx context.Context, userID, taskID string, newStatus entity.TaskStatus) (entity.Task, error)
		Delete(ctx context.Context, userID, taskID string) error
	}

	// Post -.
	Post interface {
		Create(ctx context.Context, userID, title, body string) (entity.Post, error)
		Get(ctx context.Context, postID string) (entity.Post, error)
		List(ctx context.Context, limit, offset int) ([]entity.Post, int, error)
		Update(ctx context.Context, userID, postID, title, body string) (entity.Post, error)
		Delete(ctx context.Context, userID, postID string) error
	}

	// Comment -.
	Comment interface {
		Create(ctx context.Context, userID, postID, body string) (entity.Comment, error)
		Get(ctx context.Context, commentID string) (entity.Comment, error)
		ListByPost(ctx context.Context, postID string, limit, offset int) ([]entity.Comment, int, error)
		Update(ctx context.Context, userID, commentID, body string) (entity.Comment, error)
		Delete(ctx context.Context, userID, commentID string) error
	}
)
