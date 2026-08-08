package v1

import (
	v1 "github.com/Rivaldz/my-template-go/docs/proto/v1"
	"github.com/Rivaldz/my-template-go/internal/usecase"
	"github.com/Rivaldz/my-template-go/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// TranslationController -.
type TranslationController struct {
	v1.UnimplementedTranslationServer

	t usecase.Translation
	l logger.Interface
	v *validator.Validate
}
