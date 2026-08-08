package v1

import (
	v1 "github.com/Rivaldz/my-template-go/docs/proto/v1"
	"github.com/Rivaldz/my-template-go/internal/usecase"
	"github.com/Rivaldz/my-template-go/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// AuthController -.
type AuthController struct {
	v1.UnimplementedAuthServiceServer

	u usecase.User
	l logger.Interface
	v *validator.Validate
}
