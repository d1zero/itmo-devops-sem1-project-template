package config

import (
	"os"
	"project_sem/pkg/db"
	"project_sem/pkg/http"
)

type Config struct {
	Http http.Config
	Db   db.Config
}

func New() *Config {
	return &Config{
		Http: http.Config{
			Addr: os.Getenv("API_HOST"),
		},
		Db: db.Config{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			Username: os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DbName:   os.Getenv("POSTGRES_DB"),
		},
	}
}
