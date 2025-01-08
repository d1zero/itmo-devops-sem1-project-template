package marketing

import (
	"context"
	"github.com/go-chi/chi/v5"
	"mime/multipart"
	"net/http"
	"os"
	"project_sem/internal/models"
)

type IMarketingService interface {
	Prices(context.Context, http.ResponseWriter) (*os.File, error)
	SetPrices(context.Context, multipart.File, int64) (models.AggPriceData, error)
}

type Controller struct {
	marketingService IMarketingService
}

func New(
	marketingService IMarketingService,
) *Controller {
	return &Controller{
		marketingService: marketingService,
	}
}

func (c *Controller) RegisterRoutes(r chi.Router) {
	r.Route("/prices", func(r chi.Router) {
		r.Get("/", c.Prices())
		r.Post("/", c.SetPrices())
	})
}
