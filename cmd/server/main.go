package main

import (
	"context"
	"errors"
	"log/slog"
	"myproject/internal/config"
	"myproject/internal/handler"
	"myproject/internal/model"
	"myproject/internal/repository"
	"myproject/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := config.Load()
	repo := repository.NewInMemoryItemRepo[model.Item]()
	svc := service.NewItemService(repo)
	h := handler.NewHandlerItem(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: mux,
	}

	slog.Info("сервер запущен", "port", cfg.Port)
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	case <-quit:
		slog.Info("shutting down ...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		_ = srv.Close()
		os.Exit(1)
	}
	slog.Info("server stopped gracefully")
}
