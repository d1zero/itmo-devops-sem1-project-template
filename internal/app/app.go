package app

import (
	"os"
	"os/signal"
	"project_sem/pkg/db"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"project_sem/internal/config"
	marketingControllerV0 "project_sem/internal/controller/http/v0/marketing"
	marketingInfrastructure "project_sem/internal/infrastructure/marketing"
	marketingUc "project_sem/internal/service/marketing"
	"project_sem/pkg/http"
)

func Run(cfg *config.Config, db *db.DB) {
	marketingInfra := marketingInfrastructure.New(db)

	marketingService := marketingUc.New(marketingInfra, db)

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
