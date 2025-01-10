package marketing

import (
	"context"
	"fmt"
	"mime/multipart"
	"project_sem/internal/models"
	"project_sem/pkg/helpers"
	"time"

	"github.com/shopspring/decimal"
)

func (s *Service) SetPrices(ctx context.Context, f multipart.File, fSize int64) (models.AggPriceData, error) {
	const op = "Service.SetPrices"

	rows, err := helpers.ProcessZip(f, fSize)
	if err != nil {
		return models.AggPriceData{}, fmt.Errorf("%s: %s", op, err.Error())
	}

	var result models.AggPriceData
	catsMap := make(map[string]struct{})

	if err = s.tx.InTx(ctx, func(ctx context.Context) error {
		for _, row := range rows {
			t, innerErr := time.Parse(time.DateOnly, row[4])
			if innerErr != nil {
				continue
			}

			price, innerErr := decimal.NewFromString(row[3])
			if innerErr != nil {
				continue
			}

			if innerErr = s.infra.SetPrice(ctx, models.Price{
				CreationDate: t,
				Name:         row[1],
				Category:     row[2],
				Price:        price,
			}); innerErr != nil {
				return fmt.Errorf("%s: %s", op, innerErr.Error())
			}

			result.TotalItems += 1
			result.TotalPrice = result.TotalPrice.Add(price)
			catsMap[row[2]] = struct{}{}
		}

		return nil
	}); err != nil {
		return models.AggPriceData{}, fmt.Errorf("%s: %s", op, err.Error())
	}

	result.TotalCategories = len(catsMap)

	return result, nil
}
