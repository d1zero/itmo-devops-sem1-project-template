package main

import (
	"context"
	"log"
	"project_sem/internal/app"
	"project_sem/internal/config"
	"project_sem/pkg/db"
)

func main() {
	ctx := context.Background()

	cfg := config.New()

	pool, err := db.NewPool(ctx, cfg.Db)
	if err != nil {
		log.Fatal(err)
	}

	if err = pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	// TODO: infra structure
	app.Run(cfg, pool)
}
