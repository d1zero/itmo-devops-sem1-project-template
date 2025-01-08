package marketing

import (
	"context"
	"project_sem/internal/models"
)

func (i *Infra) SetPrices(ctx context.Context, prices []models.Price) error {
	insQ := i.qb.Insert(priceTable).
		Columns("id", "name", "category", "price", "create_date")

	for _, price := range prices {
		insQ = insQ.Values(price.ID, price.Name, price.Category, price.Price, price.CreationDate)
	}

	query, args, err := insQ.ToSql()
	if err != nil {
		return err
	}

	if _, err = i.db.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}
