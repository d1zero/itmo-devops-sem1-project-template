package marketing

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Controller) SetPrices() http.HandlerFunc {
	const op = "Controller.SetPrices"
	return func(w http.ResponseWriter, r *http.Request) {
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			http.Error(w, fmt.Sprintf("%s: %s", op, err.Error()), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		result, err := c.marketingService.SetPrices(r.Context(), file, fileHeader.Size)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s: %s", op, err.Error()), http.StatusInternalServerError)
			return
		}

		res, err := json.Marshal(result)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s: %s", op, err.Error()), http.StatusInternalServerError)
			return
		}

		if _, err = w.Write(res); err != nil {
			http.Error(w, fmt.Sprintf("%s: %s", op, err.Error()), http.StatusInternalServerError)
			return
		}
	}
}
