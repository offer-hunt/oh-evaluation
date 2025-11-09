package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/offer-hunt/oh-evaluation/internal/config"
	"github.com/offer-hunt/oh-evaluation/internal/handler"
	auth_middleware "github.com/offer-hunt/oh-evaluation/internal/middleware"
)

func Run(cfg *config.Config) error {
	// Инициализация аутентификации
	auth, err := auth_middleware.NewAuthenticator(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("failed to create authenticator: %w", err)
	}

	// Настройка роутера
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Публичные эндпоинты
	r.Get("/api/ping", handler.Ping)

	// Защищенные эндпоинты
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticator)
		r.Use(auth_middleware.RequireScope("evaluation.read"))
		r.Get("/api/secure/ping", handler.SecurePing)
	})

	// Настройка и запуск сервера
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("server is starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
