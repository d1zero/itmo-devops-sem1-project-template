package marketing

import (
	"context"
	"project_sem/internal/models"
	"project_sem/pkg/db"
)

type IMarketingInfra interface {
	Prices(context.Context) ([]models.Price, error)
	SetPrice(context.Context, models.Price) error
}
type Service struct {
	infra IMarketingInfra
	tx    db.Transactor
}

func New(
	infra IMarketingInfra,
	tx db.Transactor,
) *Service {
	return &Service{
		infra: infra,
		tx:    tx,
	}
}
