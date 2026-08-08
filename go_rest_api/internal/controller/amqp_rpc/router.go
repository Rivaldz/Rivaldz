package v1

import (
	v1 "github.com/Rivaldz/my-template-go/internal/controller/amqp_rpc/v1"
	"github.com/Rivaldz/my-template-go/internal/usecase"
	"github.com/Rivaldz/my-template-go/pkg/jwt"
	"github.com/Rivaldz/my-template-go/pkg/logger"
	"github.com/Rivaldz/my-template-go/pkg/rabbitmq/rmq_rpc/server"
)

// NewRouter -.
func NewRouter(t usecase.Translation, u usecase.User, tk usecase.Task, j *jwt.Manager, l logger.Interface) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)

	{
		v1.NewRoutes(routes, t, u, tk, j, l)
	}

	return routes
}
