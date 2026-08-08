package v1

import (
	"github.com/Rivaldz/my-template-go/internal/controller/restapi/middleware"
	"github.com/Rivaldz/my-template-go/internal/usecase"
	"github.com/Rivaldz/my-template-go/pkg/jwt"
	"github.com/Rivaldz/my-template-go/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// NewRoutes -.
func NewRoutes(apiV1Group fiber.Router, t usecase.Translation, u usecase.User, tk usecase.Task, p usecase.Post, c usecase.Comment, jwtManager *jwt.Manager, l logger.Interface) {
	r := &V1{t: t, u: u, tk: tk, p: p, c: c, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	// Public routes
	authGroup := apiV1Group.Group("/auth")
	{
		authGroup.Post("/register", r.register)
		authGroup.Post("/login", r.login)
	}

	// Protected routes
	protected := apiV1Group.Group("", middleware.Auth(jwtManager))

	userGroup := protected.Group("/user")
	{
		userGroup.Get("/profile", r.profile)
	}

	taskGroup := protected.Group("/tasks")
	{
		taskGroup.Post("/", r.createTask)
		taskGroup.Get("/", r.listTasks)
		taskGroup.Get("/:id", r.getTask)
		taskGroup.Put("/:id", r.updateTask)
		taskGroup.Patch("/:id/status", r.transitionTask)
		taskGroup.Delete("/:id", r.deleteTask)
	}

	postGroup := protected.Group("/posts")
	{
		postGroup.Post("/", r.createPost)
		postGroup.Get("/", r.listPosts)
		postGroup.Get("/:id", r.getPost)
		postGroup.Put("/:id", r.updatePost)
		postGroup.Delete("/:id", r.deletePost)
		postGroup.Post("/:postID/comments", r.createComment)
		postGroup.Get("/:postID/comments", r.listComments)
	}

	commentGroup := protected.Group("/comments")
	{
		commentGroup.Get("/:id", r.getComment)
		commentGroup.Put("/:id", r.updateComment)
		commentGroup.Delete("/:id", r.deleteComment)
	}

	translationGroup := protected.Group("/translation")
	{
		translationGroup.Get("/history", r.history)
		translationGroup.Post("/do-translate", r.doTranslate)
	}
}
