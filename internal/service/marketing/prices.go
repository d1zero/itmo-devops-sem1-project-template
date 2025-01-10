package marketing

import (
	"context"
	"fmt"
	"os"
	"project_sem/pkg/helpers"
	"strconv"
	"time"
)

func (s *Service) Prices(ctx context.Context) (*os.File, int64, error) {
	const op = "Service.Prices"

	result, err := s.infra.Prices(ctx)
	if err != nil {
		return nil, 0, err
	}

	csvData := make([][]string, 0, len(result))

	for _, d := range result {
		csvData = append(csvData, []string{
			strconv.FormatInt(d.ID, 10),
			d.Name,
			d.Category,
			d.Price.String(),
			d.CreationDate.Format(time.DateOnly),
		})
	}

	file, fSize, err := helpers.WriteToZip(csvData)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	return file, fSize, nil
}
