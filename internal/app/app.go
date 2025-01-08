package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"os/signal"
	"project_sem/internal/config"
	marketingControllerV0 "project_sem/internal/controller/http/v0/marketing"
	marketingInfrastructure "project_sem/internal/infrastructure/marketing"
	marketingUc "project_sem/internal/service/marketing"
	"project_sem/pkg/http"
	"syscall"
)

func Run(cfg *config.Config, pool *pgxpool.Pool) {
	marketingInfra := marketingInfrastructure.New(pool)

	marketingService := marketingUc.New(marketingInfra)

	marketingController := marketingControllerV0.New(marketingService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Group(func(r chi.Router) {
		r.Route("/api/v0", func(r chi.Router) {
			marketingController.RegisterRoutes(r)
		})
	})

	srv := http.New(cfg.Http, r)

	exit := make(chan os.Signal, 2)

	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	srv.Start(exit)
}
