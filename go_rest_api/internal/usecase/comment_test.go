package usecase_test

import (
	"context"
	"testing"

	"github.com/Rivaldz/my-template-go/internal/entity"
	"github.com/Rivaldz/my-template-go/internal/repo"
	"github.com/Rivaldz/my-template-go/internal/usecase/comment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newCommentUseCase(t *testing.T) (*comment.UseCase, *MockCommentRepo) {
	t.Helper()

	ctrl := gomock.NewController(t)

	mockRepo := NewMockCommentRepo(ctrl)
	useCase := comment.New(mockRepo)

	return useCase, mockRepo
}

func TestCommentCreate(t *testing.T) {
	t.Parallel()

	t.Run("create success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().Store(context.Background(), gomock.Any()).Return(nil)

		c, err := uc.Create(context.Background(), "user-id-123", "post-id-123", "Nice post!")

		require.NoError(t, err)
		assert.NotEmpty(t, c.ID)
		assert.Equal(t, "post-id-123", c.PostID)
		assert.Equal(t, "user-id-123", c.UserID)
		assert.Equal(t, "Nice post!", c.Body)
	})

	t.Run("create post not found", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().Store(context.Background(), gomock.Any()).Return(entity.ErrPostNotFound)

		_, err := uc.Create(context.Background(), "user-id-123", "missing-post", "Nice post!")

		require.ErrorIs(t, err, entity.ErrPostNotFound)
	})
}

func TestCommentGet(t *testing.T) {
	t.Parallel()

	expectedComment := entity.Comment{
		ID:     "comment-id-123",
		PostID: "post-id-123",
		UserID: "user-id-123",
		Body:   "Nice post!",
	}

	t.Run("get success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "comment-id-123").Return(expectedComment, nil)

		c, err := uc.Get(context.Background(), "comment-id-123")

		require.NoError(t, err)
		assert.Equal(t, expectedComment, c)
	})

	t.Run("get not found", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "missing-id").Return(entity.Comment{}, entity.ErrCommentNotFound)

		_, err := uc.Get(context.Background(), "missing-id")

		require.ErrorIs(t, err, entity.ErrCommentNotFound)
	})
}

func TestCommentListByPost(t *testing.T) {
	t.Parallel()

	comment1 := entity.Comment{ID: "comment-1", PostID: "post-id-123", Body: "First"}
	comment2 := entity.Comment{ID: "comment-2", PostID: "post-id-123", Body: "Second"}

	t.Run("list success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().ListByPost(context.Background(), "post-id-123", repo.CommentFilter{
			Limit:  uint64(10),
			Offset: uint64(0),
		}).Return([]entity.Comment{comment1, comment2}, 2, nil)

		comments, total, err := uc.ListByPost(context.Background(), "post-id-123", 10, 0)

		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, comments, 2)
	})

	t.Run("list defaults", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().ListByPost(context.Background(), "post-id-123", repo.CommentFilter{
			Limit:  uint64(10),
			Offset: uint64(0),
		}).Return([]entity.Comment{comment1, comment2}, 2, nil)

		comments, total, err := uc.ListByPost(context.Background(), "post-id-123", 0, -1)

		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, comments, 2)
	})

	t.Run("list error", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().ListByPost(context.Background(), "post-id-123", gomock.Any()).Return(nil, 0, errRepoGeneric)

		_, _, err := uc.ListByPost(context.Background(), "post-id-123", 10, 0)

		require.ErrorIs(t, err, errRepoGeneric)
	})
}

func TestCommentUpdate(t *testing.T) {
	t.Parallel()

	existingComment := entity.Comment{
		ID:     "comment-id-123",
		PostID: "post-id-123",
		UserID: "user-id-123",
		Body:   "Old comment",
	}

	t.Run("update success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "comment-id-123").Return(existingComment, nil)
		mockRepo.EXPECT().Update(context.Background(), gomock.Any()).Return(nil)

		updated, err := uc.Update(context.Background(), "user-id-123", "comment-id-123", "New comment")

		require.NoError(t, err)
		assert.Equal(t, "New comment", updated.Body)
	})

	t.Run("update forbidden", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "comment-id-123").Return(existingComment, nil)

		_, err := uc.Update(context.Background(), "other-user", "comment-id-123", "New comment")

		require.ErrorIs(t, err, entity.ErrCommentForbidden)
	})

	t.Run("update not found", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "missing-id").Return(entity.Comment{}, entity.ErrCommentNotFound)

		_, err := uc.Update(context.Background(), "user-id-123", "missing-id", "New comment")

		require.ErrorIs(t, err, entity.ErrCommentNotFound)
	})
}

func TestCommentDelete(t *testing.T) {
	t.Parallel()

	existingComment := entity.Comment{
		ID:     "comment-id-123",
		PostID: "post-id-123",
		UserID: "user-id-123",
		Body:   "Nice post!",
	}

	t.Run("delete success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "comment-id-123").Return(existingComment, nil)
		mockRepo.EXPECT().Delete(context.Background(), "comment-id-123", "user-id-123").Return(nil)

		err := uc.Delete(context.Background(), "user-id-123", "comment-id-123")

		require.NoError(t, err)
	})

	t.Run("delete forbidden", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "comment-id-123").Return(existingComment, nil)

		err := uc.Delete(context.Background(), "other-user", "comment-id-123")

		require.ErrorIs(t, err, entity.ErrCommentForbidden)
	})

	t.Run("delete not found", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "missing-id").Return(entity.Comment{}, entity.ErrCommentNotFound)

		err := uc.Delete(context.Background(), "user-id-123", "missing-id")

		require.ErrorIs(t, err, entity.ErrCommentNotFound)
	})

	t.Run("delete repo error", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newCommentUseCase(t)
		mockRepo.EXPECT().GetByID(context.Background(), "comment-id-123").Return(existingComment, nil)
		mockRepo.EXPECT().Delete(context.Background(), "comment-id-123", "user-id-123").Return(errRepoGeneric)

		err := uc.Delete(context.Background(), "user-id-123", "comment-id-123")

		require.ErrorIs(t, err, errRepoGeneric)
	})
}
