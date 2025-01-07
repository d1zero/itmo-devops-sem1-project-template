package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Data struct {
	ID           int64
	CreationDate time.Time
	Name         string
	Category     string
	Price        float64
}

type Response struct {
	TotalItems      int     `json:"total_items" db:"total_items"`
	TotalCategories int     `json:"total_categories" db:"total_categories"`
	TotalPrice      float64 `json:"total_price" db:"total_price"`
}

func main() {
	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	priceTable := "prices"

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		30,
		60*time.Second,
		60*time.Second,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err = pool.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/api/v0/prices", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			archiveType := r.URL.Query().Get("type")
			if archiveType == "" {
				archiveType = "zip" // Тип по умолчанию
			}

			file, fileHeader, err := r.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer r.Body.Close()

			var rows [][]string
			if archiveType == "tar" {
				rows, err = processTar(file)
			} else {
				rows, err = processZip(file, fileHeader.Size)
			}

			if err != nil {
				http.Error(w, "Ошибка обработки архива", http.StatusInternalServerError)
				log.Println("Archive processing error:", err)
				return
			}

			insQ := queryBuilder.Insert(priceTable).
				Columns("id", "name", "category", "price", "create_date")

			for _, row := range rows {
				insQ = insQ.Values(row[0], row[1], row[2], row[3], row[4])
			}

			query, args, err := insQ.ToSql()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := pool.Exec(r.Context(), query, args...); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Get data from db
			getQ := queryBuilder.
				Select("COUNT(*) as total_items", "COUNT(DISTINCT category) as total_categories", "SUM(price) as total_price").
				From(priceTable)

			query, args, err = getQ.ToSql()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var resp Response

			if err = pgxscan.Get(r.Context(), pool, &resp, query, args...); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			q := queryBuilder.
				Select("id", "name", "category", "price", "create_date").
				From(priceTable)

			query, args, err := q.ToSql()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var result []Data

			if err = pgxscan.Select(r.Context(), pool, &result, query, args...); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			f, err := os.CreateTemp("", "data-*.csv")
			if err != nil {
				fmt.Println(err)
			}
			defer os.Remove(f.Name())

			writer := csv.NewWriter(f)

			if err := writer.Write([]string{"id", "name", "category", "price", "create_date"}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			zipFile, err := os.CreateTemp("", "data-*.zip")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer os.Remove(zipFile.Name())

			zipWriter := zip.NewWriter(zipFile)
			defer zipWriter.Close()

			// Добавляем CSV в ZIP
			csvInZip, err := zipWriter.Create("data.csv")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			f.Seek(0, 0) // Перемещаем указатель файла в начало
			if _, err := f.WriteTo(csvInZip); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			zipWriter.Close()

			// Возвращаем ZIP клиенту
			zipFile.Seek(0, 0)
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", "attachment; filename=\"data.zip\"")
			dataZip, err := zipFile.Stat()
			if _, err := f.WriteTo(csvInZip); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", dataZip.Size()))

			if _, err := zipFile.WriteTo(w); err != nil {
				log.Println("Error sending ZIP to client:", err)
			}

			return
		})
	})

	http.ListenAndServe(":8080", r)
}

// processZip обрабатывает zip-архив
func processZip(file io.ReaderAt, size int64) ([][]string, error) {
	reader, err := zip.NewReader(file, size)
	if err != nil {
		return nil, err
	}

	var allRecords [][]string
	for _, f := range reader.File {
		if strings.HasSuffix(f.Name, ".csv") {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			records, err := readCSV(rc)
			if err != nil {
				return nil, err
			}
			allRecords = append(allRecords, records...)
		}
	}

	return allRecords[1:], nil
}

// processTar обрабатывает tar-архив
func processTar(file io.Reader) ([][]string, error) {
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания gzip-ридера: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	var allRecords [][]string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения tar-записи: %w", err)
		}

		if strings.HasSuffix(header.Name, ".csv") {
			records, err := readCSV(tarReader)
			if err != nil {
				return nil, fmt.Errorf("ошибка чтения CSV: %w", err)
			}
			allRecords = append(allRecords, records...)
		}
	}

	return allRecords[2:], nil
}

// readCSV читает CSV-данные из io.Reader
func readCSV(r io.Reader) ([][]string, error) {
	reader := csv.NewReader(r)
	return reader.ReadAll()
}
