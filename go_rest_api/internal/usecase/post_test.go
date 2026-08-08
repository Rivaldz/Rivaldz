package usecase_test

import (
	"context"
	"testing"

	"github.com/Rivaldz/my-template-go/internal/entity"
	"github.com/Rivaldz/my-template-go/internal/repo"
	"github.com/Rivaldz/my-template-go/internal/usecase/post"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newPostUseCase(t *testing.T) (*post.UseCase, *MockPostRepo) {
	t.Helper()

	ctrl := gomock.NewController(t)

	mockRepo := NewMockPostRepo(ctrl)
	useCase := post.New(mockRepo)

	return useCase, mockRepo
}

func TestPostCreate(t *testing.T) {
	t.Parallel()

	t.Run("create success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().Store(context.Background(), gomock.Any()).Return(nil)

		p, err := uc.Create(context.Background(), "user-id-123", "My Post", "Post body")

		require.NoError(t, err)
		assert.NotEmpty(t, p.ID)
		assert.Equal(t, "user-id-123", p.UserID)
		assert.Equal(t, "My Post", p.Title)
		assert.Equal(t, "Post body", p.Body)
		assert.Equal(t, 0, p.CommentCount)
	})

	t.Run("store error", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().Store(context.Background(), gomock.Any()).Return(errRepoGeneric)

		_, err := uc.Create(context.Background(), "user-id-123", "My Post", "Post body")

		require.ErrorIs(t, err, errRepoGeneric)
	})
}

func TestPostGet(t *testing.T) {
	t.Parallel()

	expectedPost := entity.Post{
		ID:           "post-id-123",
		UserID:       "user-id-123",
		Title:        "My Post",
		Body:         "Post body",
		CommentCount: 2,
	}

	t.Run("get success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "post-id-123").Return(expectedPost, nil)

		p, err := uc.Get(context.Background(), "post-id-123")

		require.NoError(t, err)
		assert.Equal(t, expectedPost, p)
	})

	t.Run("get not found", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "missing-id").Return(entity.Post{}, entity.ErrPostNotFound)

		_, err := uc.Get(context.Background(), "missing-id")

		require.ErrorIs(t, err, entity.ErrPostNotFound)
	})
}

func TestPostList(t *testing.T) {
	t.Parallel()

	post1 := entity.Post{ID: "post-1", Title: "Post 1"}
	post2 := entity.Post{ID: "post-2", Title: "Post 2"}

	t.Run("list success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().List(context.Background(), repo.PostFilter{
			Limit:  uint64(10),
			Offset: uint64(0),
		}).Return([]entity.Post{post1, post2}, 2, nil)

		posts, total, err := uc.List(context.Background(), 10, 0)

		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, posts, 2)
	})

	t.Run("list defaults", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().List(context.Background(), repo.PostFilter{
			Limit:  uint64(10),
			Offset: uint64(0),
		}).Return([]entity.Post{post1, post2}, 2, nil)

		posts, total, err := uc.List(context.Background(), 0, -1)

		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, posts, 2)
	})

	t.Run("list error", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().List(context.Background(), gomock.Any()).Return(nil, 0, errRepoGeneric)

		_, _, err := uc.List(context.Background(), 10, 0)

		require.ErrorIs(t, err, errRepoGeneric)
	})
}

func TestPostUpdate(t *testing.T) {
	t.Parallel()

	existingPost := entity.Post{
		ID:     "post-id-123",
		UserID: "user-id-123",
		Title:  "Old Title",
		Body:   "Old body",
	}

	t.Run("update success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "post-id-123").Return(existingPost, nil)
		mockRepo.EXPECT().Update(context.Background(), gomock.Any()).Return(nil)

		updated, err := uc.Update(context.Background(), "user-id-123", "post-id-123", "New Title", "New body")

		require.NoError(t, err)
		assert.Equal(t, "New Title", updated.Title)
		assert.Equal(t, "New body", updated.Body)
	})

	t.Run("update forbidden", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "post-id-123").Return(existingPost, nil)

		_, err := uc.Update(context.Background(), "other-user", "post-id-123", "New Title", "New body")

		require.ErrorIs(t, err, entity.ErrPostForbidden)
	})

	t.Run("update not found", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "missing-id").Return(entity.Post{}, entity.ErrPostNotFound)

		_, err := uc.Update(context.Background(), "user-id-123", "missing-id", "New Title", "New body")

		require.ErrorIs(t, err, entity.ErrPostNotFound)
	})
}

func TestPostDelete(t *testing.T) {
	t.Parallel()

	existingPost := entity.Post{
		ID:     "post-id-123",
		UserID: "user-id-123",
		Title:  "My Post",
	}

	t.Run("delete success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "post-id-123").Return(existingPost, nil)
		mockRepo.EXPECT().Delete(context.Background(), "post-id-123", "user-id-123").Return(nil)

		err := uc.Delete(context.Background(), "user-id-123", "post-id-123")

		require.NoError(t, err)
	})

	t.Run("delete forbidden", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "post-id-123").Return(existingPost, nil)

		err := uc.Delete(context.Background(), "other-user", "post-id-123")

		require.ErrorIs(t, err, entity.ErrPostForbidden)
	})

	t.Run("delete not found", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "missing-id").Return(entity.Post{}, entity.ErrPostNotFound)

		err := uc.Delete(context.Background(), "user-id-123", "missing-id")

		require.ErrorIs(t, err, entity.ErrPostNotFound)
	})

	t.Run("delete repo error", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPostUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "post-id-123").Return(existingPost, nil)
		mockRepo.EXPECT().Delete(context.Background(), "post-id-123", "user-id-123").Return(errRepoGeneric)

		err := uc.Delete(context.Background(), "user-id-123", "post-id-123")

		require.ErrorIs(t, err, errRepoGeneric)
	})
}
