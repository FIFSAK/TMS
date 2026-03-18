package app

import (
	"time"

	"github.com/FIFSAK/TMS/internal/config"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/FIFSAK/TMS/internal/handler"
	"github.com/FIFSAK/TMS/internal/repository"
	"github.com/FIFSAK/TMS/internal/service"
	"github.com/FIFSAK/TMS/pkg/server"
)

const (
	dbStartupDelay = time.Second
)

func initApp(logger *zap.Logger) (*App, error) {
	app := &App{logger: logger}

	if err := app.loadConfiguration(); err != nil {
		return nil, err
	}

	if err := app.initializeRepositories(); err != nil {
		return nil, err
	}

	if err := app.initializeServices(); err != nil {
		app.cleanup()
		return nil, err
	}

	if err := app.initializeHandlers(); err != nil {
		app.cleanup()
		return nil, err
	}

	if err := app.initializeServers(); err != nil {
		app.cleanup()
		return nil, err
	}

	return app, nil
}

func (app *App) loadConfiguration() error {
	configs, err := config.New()
	if err != nil {
		app.logger.Error("config load error", zap.Error(err))
		return err
	}

	app.configs = configs
	app.logger.Info("configuration loaded",
		zap.String("mode", configs.APP.Mode),
		zap.String("port", configs.APP.Port),
	)

	return nil
}

func (app *App) initializeRepositories() error {
	app.logger.Info("initializing repositories",
		zap.Duration("db startup delay", dbStartupDelay))
	time.Sleep(dbStartupDelay)

	configs := []repository.Configuration{
		repository.WithPostgresStore(app.configs.Store.DSN),
	}

	repositories, err := repository.New(configs...)
	if err != nil {
		app.logger.Error("repository init error", zap.Error(err))
		return err
	}

	app.repositories = repositories
	app.logger.Info("repositories initialized")
	return nil
}

func (app *App) initializeServices() error {
	services, err := service.New(
		service.Dependencies{
			Repositories: app.repositories,
			Configs:      app.configs,
		},
		service.WithLibraryService(),
	)
	if err != nil {
		app.logger.Error("service init error", zap.Error(err))
		return err
	}

	app.services = services
	app.logger.Info("services initialized")
	return nil
}

func (app *App) initializeHandlers() error {
	handlers, err := handler.New(
		handler.Dependencies{
			Configs:  app.configs,
			Services: app.services,
		},
		//	TODO: add with grpc handler option
	)
	if err != nil {
		app.logger.Error("handler init error", zap.Error(err))
		return err
	}

	app.handlers = handlers
	app.logger.Info("handlers initialized")
	return nil
}

func (app *App) initializeServers() error {
	servers, err := server.NewServer(
		//TODO: write gropc options
		server.WithGRPC(),
	)
	if err != nil {
		app.logger.Error("server init error", zap.Error(err))
		return err
	}

	app.servers = servers
	app.logger.Info("servers initialized")
	return nil
}

func (app *App) natsHealthHandler(msg *nats.Msg) (interface{}, error) {
	return map[string]string{
		"status":  "healthy",
		"service": "github.com/FIFSAK/TMS",
	}, nil
}

func (app *App) cleanup() {
	if app.repositories != nil {
		app.repositories.Close()
		app.logger.Info("repositories cleanup complete")
	}
}
