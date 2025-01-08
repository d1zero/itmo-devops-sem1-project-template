package marketing

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"project_sem/internal/models"
)

func (i *Infra) AggPriceData(ctx context.Context) (models.AggPriceData, error) {
	q := i.qb.
		Select(
			"COUNT(*) as total_items",
			"COUNT(DISTINCT category) as total_categories",
			"SUM(price) as total_price",
		).
		From(priceTable)

	query, args, err := q.ToSql()
	if err != nil {
		return models.AggPriceData{}, err
	}

	resp := models.AggPriceData{}

	if err = pgxscan.Get(ctx, i.db, &resp, query, args...); err != nil {
		return models.AggPriceData{}, err
	}

	return resp, nil
}
