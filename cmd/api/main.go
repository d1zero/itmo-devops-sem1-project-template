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

	db, err := db.New(ctx, cfg.Db)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err = db.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	app.Run(cfg, db)
}
