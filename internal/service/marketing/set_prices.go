package marketing

import (
	"context"
	"mime/multipart"
	"project_sem/internal/models"
	"project_sem/pkg/helpers"
	"strconv"
	"time"
)

func (s *Service) SetPrices(ctx context.Context, f multipart.File, fSize int64) (models.AggPriceData, error) {
	rows, err := helpers.ProcessZip(f, fSize)
	if err != nil {
		return models.AggPriceData{}, err
	}

	var data []models.Price

	for _, row := range rows {
		id, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			continue
		}
		t, err := time.Parse(time.DateOnly, row[4])
		if err != nil {
			continue
		}
		price, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			continue
		}

		data = append(data, models.Price{
			ID:           id,
			CreationDate: t,
			Name:         row[1],
			Category:     row[2],
			Price:        price,
		})
	}

	if err = s.infra.SetPrices(ctx, data); err != nil {
		return models.AggPriceData{}, err
	}

	aggData, err := s.infra.AggPriceData(ctx)
	if err != nil {
		return models.AggPriceData{}, err
	}

	return aggData, nil
}
