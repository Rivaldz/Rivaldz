package restapi

import (
	"net/http"

	"github.com/Rivaldz/my-template-go/config"
	_ "github.com/Rivaldz/my-template-go/docs" // Swagger docs.
	"github.com/Rivaldz/my-template-go/internal/controller/restapi/middleware"
	v1 "github.com/Rivaldz/my-template-go/internal/controller/restapi/v1"
	"github.com/Rivaldz/my-template-go/internal/usecase"
	"github.com/Rivaldz/my-template-go/pkg/jwt"
	"github.com/Rivaldz/my-template-go/pkg/logger"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// NewRouter -.
// Swagger spec:
//
//	@title       Go Clean Template API
//	@description Multi-domain clean architecture template with translation, user, and task management
//	@version     1.0
//	@host        localhost:8080
//	@BasePath    /v1
//	@securityDefinitions.apikey BearerAuth
//	@in header
//	@name Authorization
func NewRouter(app *fiber.App, cfg *config.Config, t usecase.Translation, u usecase.User, tk usecase.Task, p usecase.Post, c usecase.Comment, jwtManager *jwt.Manager, l logger.Interface) {
	// Options
	app.Use(middleware.Logger(l))
	app.Use(middleware.Recovery(l))

	// Prometheus metrics
	if cfg.Metrics.Enabled {
		prometheus := fiberprometheus.New("my-service-name")
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	// Swagger
	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// K8s probe
	app.Get("/healthz", func(ctx *fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) })

	// Routers
	apiV1Group := app.Group("/v1")
	{
		v1.NewRoutes(apiV1Group, t, u, tk, p, c, jwtManager, l)
	}
}
