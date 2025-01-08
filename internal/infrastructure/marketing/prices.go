package marketing

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"project_sem/internal/models"
)

func (i *Infra) Prices(ctx context.Context) ([]models.Price, error) {
	q := i.qb.
		Select("id", "create_date", "name", "category", "price").
		From(priceTable)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	var result []models.Price

	if err = pgxscan.Select(ctx, i.db, &result, query, args...); err != nil {
		return nil, err
	}

	return result, nil
}
