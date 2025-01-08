package marketing

import (
	"context"
	"project_sem/internal/models"
)

type IMarketingInfra interface {
	Prices(context.Context) ([]models.Price, error)
	SetPrices(context.Context, []models.Price) error
	AggPriceData(context.Context) (models.AggPriceData, error)
}
type Service struct {
	infra IMarketingInfra
}

func New(infra IMarketingInfra) *Service {
	return &Service{
		infra: infra,
	}
}
