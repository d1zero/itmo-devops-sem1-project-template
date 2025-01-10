package marketing

import (
	"context"
	"fmt"
	"project_sem/internal/models"
)

func (i *Infra) SetPrice(ctx context.Context, price models.Price) error {
	const op = "Infra.SetPrice"

	insQ := i.db.Builder().Insert(priceTable).
		Columns("name", "category", "price", "create_date").
		Values(price.Name, price.Category, price.Price, price.CreationDate)

	if _, err := i.db.Insert(ctx, insQ); err != nil {
		return fmt.Errorf("%s: %s", op, err.Error())
	}

	return nil
}
