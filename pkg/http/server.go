package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
)

type Config struct {
	Addr string
}

type Server struct {
	srv *http.Server
}

func New(cfg Config, r chi.Router) *Server {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:8080"
	}
	return &Server{
		srv: &http.Server{
			Addr:    cfg.Addr,
			Handler: r,
		},
	}

}

func (s *Server) Start(sig chan os.Signal) {
	// Канал для перехвата системных сигналов
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt) // Подписываемся на SIGINT (Ctrl+C)

	// Запуск сервера в отдельной горутине
	go func() {
		log.Println("Starting server on :8080")
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание системного сигнала
	<-stopChan
	log.Println("Shutdown signal received")

	// Контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Завершаем сервер
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
