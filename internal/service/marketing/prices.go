package marketing

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func (s *Service) Prices(ctx context.Context, w http.ResponseWriter) (*os.File, error) {
	f, err := os.CreateTemp("", "data-*.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer os.Remove(f.Name())

	writer := csv.NewWriter(f)

	if err = writer.Write([]string{"id", "name", "category", "price", "create_date"}); err != nil {
		return nil, err
	}

	result, err := s.infra.Prices(ctx)
	if err != nil {
		return nil, err
	}

	for _, d := range result {
		err = writer.Write([]string{
			strconv.FormatInt(d.ID, 10),
			d.Name,
			d.Category,
			strconv.FormatFloat(d.Price, 'f', -1, 64),
			d.CreationDate.Format(time.DateOnly),
		})
		if err != nil {
			fmt.Println(err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	zipFile, err := os.CreateTemp("", "data-*.zip")
	if err != nil {
		return nil, err
	}
	defer os.Remove(zipFile.Name())

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Добавляем CSV в ZIP
	csvInZip, err := zipWriter.Create("data.csv")
	if err != nil {
		return nil, err
	}

	f.Seek(0, 0) // Перемещаем указатель файла в начало
	if _, err := f.WriteTo(csvInZip); err != nil {
		return nil, err
	}

	zipWriter.Close()

	// TODO: refactor

	// Возвращаем ZIP клиенту
	zipFile.Seek(0, 0)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"response.zip\"")
	dataZip, err := zipFile.Stat()
	if _, err := f.WriteTo(csvInZip); err != nil {
		return nil, err
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", dataZip.Size()))

	if _, err := zipFile.WriteTo(w); err != nil {
		log.Println("Error sending ZIP to client:", err)
	}

	return nil, nil
}
