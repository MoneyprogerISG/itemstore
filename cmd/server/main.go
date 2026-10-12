package main

import (
	"log/slog"
	"myproject/internal/config"
	"myproject/internal/handler"
	"myproject/internal/model"
	"myproject/internal/repository"
	"myproject/internal/service"
	"net/http"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := config.Load()
	repo := repository.NewInMemoryItemRepo[model.Item]()
	svc := service.NewItemService(repo)
	h := handler.NewHandlerItem(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	slog.Info("сервер запущен", "port", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, mux); err != nil {
		slog.Error("сервер оостановлен", "error", err)
		os.Exit(1)
	}
}
