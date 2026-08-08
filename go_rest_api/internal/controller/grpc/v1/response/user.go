package response

import (
	v1 "github.com/Rivaldz/my-template-go/docs/proto/v1"
	"github.com/Rivaldz/my-template-go/internal/entity"
)

// NewRegisterResponse -.
func NewRegisterResponse(user *entity.User) *v1.RegisterResponse {
	return &v1.RegisterResponse{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// NewGetProfileResponse -.
func NewGetProfileResponse(user *entity.User) *v1.GetProfileResponse {
	return &v1.GetProfileResponse{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
