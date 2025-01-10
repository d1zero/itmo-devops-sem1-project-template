package marketing

import (
	"project_sem/pkg/db"
)

const priceTable = "prices"

type Infra struct {
	db *db.DB
}

func New(db *db.DB) *Infra {
	return &Infra{
		db: db,
	}
}
