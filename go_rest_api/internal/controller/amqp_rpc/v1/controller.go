package v1

import (
	"github.com/Rivaldz/my-template-go/internal/usecase"
	"github.com/Rivaldz/my-template-go/pkg/jwt"
	"github.com/Rivaldz/my-template-go/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// V1 -.
type V1 struct {
	t  usecase.Translation
	u  usecase.User
	tk usecase.Task
	j  *jwt.Manager
	l  logger.Interface
	v  *validator.Validate
}
