package http

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Addr string
}

type Server struct {
	srv *http.Server
}

func New(cfg Config, r chi.Router) *Server {
	return &Server{
		srv: &http.Server{
			Addr:    cfg.Addr,
			Handler: r,
		},
	}

}

func (s *Server) Start(sig chan os.Signal) {
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	go func() {
		<-sig

		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := s.srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	err := s.srv.ListenAndServe()
	if err != nil {
		log.Fatalf("error on server.ListenAndServe: %v", err)
	}

	<-serverCtx.Done()
}
