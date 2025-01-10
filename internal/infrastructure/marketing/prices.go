package marketing

import (
	"context"
	"fmt"
	"project_sem/internal/models"
)

func (i *Infra) Prices(ctx context.Context) ([]models.Price, error) {
	const op = "Infra.Prices"

	q := i.db.Builder().
		Select("id", "create_date", "name", "category", "price").
		From(priceTable)

	var result []models.Price

	if err := i.db.Select(ctx, &result, q); err != nil {
		return nil, fmt.Errorf("%s: %s", op, err.Error())
	}

	return result, nil
}
