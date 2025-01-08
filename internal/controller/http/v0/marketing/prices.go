package marketing

import (
	"net/http"
)

func (c *Controller) Prices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := c.marketingService.Prices(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}
}
