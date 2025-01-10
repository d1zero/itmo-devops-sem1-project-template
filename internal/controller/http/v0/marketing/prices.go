package marketing

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func (c *Controller) Prices() http.HandlerFunc {
	const op = "Controller.Prices"

	return func(w http.ResponseWriter, r *http.Request) {
		file, fileSize, err := c.marketingService.Prices(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf("%s: %s", op, err.Error()), http.StatusInternalServerError)
			return
		}
		defer os.Remove(file.Name())

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\"response.zip\"")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))

		if _, err := file.WriteTo(w); err != nil {
			log.Println("Error sending ZIP to client:", err)
		}
	}
}
