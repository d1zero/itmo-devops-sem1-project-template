package marketing

import (
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const priceTable = "prices"

type Infra struct {
	db *pgxpool.Pool
	qb squirrel.StatementBuilderType
}

func New(pool *pgxpool.Pool) *Infra {
	return &Infra{
		db: pool,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
