package app

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/httpapi"
	"gophkeeper/internal/server/repository/sqlite"
	"gophkeeper/internal/server/service"
)

type App struct {
	version   string
	buildDate string
	logger    *log.Logger
	server    *http.Server
}

func New(version, buildDate string, logger *log.Logger) (*App, error) {
	cfg := config.Load()
	repo, err := sqlite.New(cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	services := service.NewServices(repo, cfg)
	router := httpapi.NewRouter(services, logger)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return &App{version: version, buildDate: buildDate, logger: logger, server: server}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Printf("http server error: %v", err)
		}
	}()

	a.logger.Printf("GophKeeper server %s (%s) listening on %s", a.version, a.buildDate, a.server.Addr)

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.server.Shutdown(shutdownCtx)
}
