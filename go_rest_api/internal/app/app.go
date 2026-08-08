// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rivaldz/my-template-go/config"
	amqprpc "github.com/Rivaldz/my-template-go/internal/controller/amqp_rpc"
	"github.com/Rivaldz/my-template-go/internal/controller/grpc"
	grpcmw "github.com/Rivaldz/my-template-go/internal/controller/grpc/middleware"
	natsrpc "github.com/Rivaldz/my-template-go/internal/controller/nats_rpc"
	"github.com/Rivaldz/my-template-go/internal/controller/restapi"
	"github.com/Rivaldz/my-template-go/internal/repo/persistent"
	"github.com/Rivaldz/my-template-go/internal/repo/webapi"
	"github.com/Rivaldz/my-template-go/internal/usecase/comment"
	"github.com/Rivaldz/my-template-go/internal/usecase/post"
	"github.com/Rivaldz/my-template-go/internal/usecase/task"
	"github.com/Rivaldz/my-template-go/internal/usecase/translation"
	"github.com/Rivaldz/my-template-go/internal/usecase/user"
	"github.com/Rivaldz/my-template-go/pkg/grpcserver"
	"github.com/Rivaldz/my-template-go/pkg/httpserver"
	"github.com/Rivaldz/my-template-go/pkg/jwt"
	"github.com/Rivaldz/my-template-go/pkg/logger"
	natsRPCServer "github.com/Rivaldz/my-template-go/pkg/nats/nats_rpc/server"
	"github.com/Rivaldz/my-template-go/pkg/postgres"
	rmqRPCServer "github.com/Rivaldz/my-template-go/pkg/rabbitmq/rmq_rpc/server"
	pbgrpc "google.golang.org/grpc"
)

type useCases struct {
	translation *translation.UseCase
	user        *user.UseCase
	task        *task.UseCase
	post        *post.UseCase
	comment     *comment.UseCase
}

type servers struct {
	rmq  *rmqRPCServer.Server
	nats *natsRPCServer.Server
	grpc *grpcserver.Server
	http *httpserver.Server
}

func initUseCases(pg *postgres.Postgres, jwtManager *jwt.Manager) useCases {
	userRepo := persistent.NewUserRepo(pg)
	taskRepo := persistent.NewTaskRepo(pg)
	translationRepo := persistent.NewTranslationRepo(pg)
	postRepo := persistent.NewPostRepo(pg)
	commentRepo := persistent.NewCommentRepo(pg)

	return useCases{
		user:        user.New(userRepo, jwtManager),
		task:        task.New(taskRepo),
		translation: translation.New(translationRepo, webapi.New()),
		post:        post.New(postRepo),
		comment:     comment.New(commentRepo),
	}
}

func initServers(cfg *config.Config, uc useCases, jwtManager *jwt.Manager, l logger.Interface) servers {
	var rmqServer *rmqRPCServer.Server
	var natsServer *natsRPCServer.Server
	var grpcServer *grpcserver.Server
	var httpServer *httpserver.Server
	var err error

	// RabbitMQ RPC Server
	if cfg.Features.EnableRabbitMQ {
		rmqRouter := amqprpc.NewRouter(uc.translation, uc.user, uc.task, jwtManager, l)
		rmqServer, err = rmqRPCServer.New(cfg.RMQ.URL, cfg.RMQ.ServerExchange, rmqRouter, l)
		if err != nil {
			l.Fatal(fmt.Errorf("app - Run - rmqServer - server.New: %w", err))
		}
	}

	// NATS RPC Server
	if cfg.Features.EnableNATS {
		natsRouter := natsrpc.NewRouter(uc.translation, uc.user, uc.task, jwtManager, l)
		natsServer, err = natsRPCServer.New(cfg.NATS.URL, cfg.NATS.ServerExchange, natsRouter, l)
		if err != nil {
			l.Fatal(fmt.Errorf("app - Run - natsServer - server.New: %w", err))
		}
	}

	// gRPC Server
	if cfg.Features.EnableGRPC {
		grpcServer = grpcserver.New(
			l,
			grpcserver.Port(cfg.GRPC.Port),
			grpcserver.ServerOptions(pbgrpc.UnaryInterceptor(grpcmw.AuthInterceptor(jwtManager))),
		)
		grpc.NewRouter(grpcServer.App, uc.translation, uc.user, uc.task, l)
	}

	// HTTP Server
	if cfg.Features.EnableREST {
		httpServer = httpserver.New(l, httpserver.Addr(cfg.HTTP.URL), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
		restapi.NewRouter(httpServer.App, cfg, uc.translation, uc.user, uc.task, uc.post, uc.comment, jwtManager, l)
	}

	return servers{
		rmq:  rmqServer,
		nats: natsServer,
		grpc: grpcServer,
		http: httpServer,
	}
}

func (s *servers) startServers() {
	if s.rmq != nil {
		s.rmq.Start()
	}
	if s.nats != nil {
		s.nats.Start()
	}
	if s.grpc != nil {
		s.grpc.Start()
	}
	if s.http != nil {
		s.http.Start()
	}
}

func (s *servers) waitForShutdown(l logger.Interface) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var err error
	var rmqNotify, natsNotify, grpcNotify, httpNotify <-chan error

	if s.rmq != nil {
		rmqNotify = s.rmq.Notify()
	}
	if s.nats != nil {
		natsNotify = s.nats.Notify()
	}
	if s.grpc != nil {
		grpcNotify = s.grpc.Notify()
	}
	if s.http != nil {
		httpNotify = s.http.Notify()
	}

	select {
	case sig := <-interrupt:
		l.Info("app - Run - signal: %s", sig.String())
	case err = <-httpNotify:
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	case err = <-grpcNotify:
		l.Error(fmt.Errorf("app - Run - grpcServer.Notify: %w", err))
	case err = <-rmqNotify:
		l.Error(fmt.Errorf("app - Run - rmqServer.Notify: %w", err))
	case err = <-natsNotify:
		l.Error(fmt.Errorf("app - Run - natsServer.Notify: %w", err))
	}

	s.shutdownServers(l)
}

func (s *servers) shutdownServers(l logger.Interface) {
	if s.http != nil {
		if err := s.http.Shutdown(); err != nil {
			l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
		}
	}

	if s.grpc != nil {
		if err := s.grpc.Shutdown(); err != nil {
			l.Error(fmt.Errorf("app - Run - grpcServer.Shutdown: %w", err))
		}
	}

	if s.rmq != nil {
		if err := s.rmq.Shutdown(); err != nil {
			l.Error(fmt.Errorf("app - Run - rmqServer.Shutdown: %w", err))
		}
	}

	if s.nats != nil {
		if err := s.nats.Shutdown(); err != nil {
			l.Error(fmt.Errorf("app - Run - natsServer.Shutdown: %w", err))
		}
	}
}

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// JWT
	jwtManager := jwt.New(cfg.JWT.Secret, cfg.JWT.TokenExpiry)

	uc := initUseCases(pg, jwtManager)
	s := initServers(cfg, uc, jwtManager, l)
	s.startServers()
	s.waitForShutdown(l)
}
