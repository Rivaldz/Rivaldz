package v1

import (
	v1 "github.com/Rivaldz/my-template-go/docs/proto/v1"
	"github.com/Rivaldz/my-template-go/internal/usecase"
	"github.com/Rivaldz/my-template-go/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// TaskController -.
type TaskController struct {
	v1.UnimplementedTaskServiceServer

	tk usecase.Task
	l  logger.Interface
	v  *validator.Validate
}
