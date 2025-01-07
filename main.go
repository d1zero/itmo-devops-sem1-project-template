package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
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
	TotalItems      int     `json:"total_items"`
	TotalCategories int     `json:"total_categories"`
	TotalPrice      float64 `json:"total_price"`
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

			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, r.Body); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			var rows [][]string
			if archiveType == "zip" {
				// Обработка zip-архива
				rows, err = processZip(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
			} else if archiveType == "tar" {
				// Обработка tar-архива
				rows, err = processTar(bytes.NewReader(buf.Bytes()))
			} else {
				http.Error(w, "Неподдерживаемый тип архива", http.StatusBadRequest)
				return
			}

			if err != nil {
				http.Error(w, "Ошибка обработки архива", http.StatusInternalServerError)
				log.Println("Archive processing error:", err)
				return
			}

			insQ := queryBuilder.Insert(priceTable).
				Columns("id", "name", "category", "price", "creation_date")

			price := .0
			cats := map[string]struct{}{}
			for _, row := range rows[1:] {
				insQ = insQ.Values(row[0], row[1], row[2], row[3], row[4])

				pr, _ := strconv.ParseFloat(row[3], 8)
				price += pr
				cats[row[2]] = struct{}{}
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

			response := Response{
				TotalItems:      len(rows) - 1,
				TotalCategories: len(cats),
				TotalPrice:      price,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			q := queryBuilder.
				Select("id", "name", "category", "price", "creation_date").
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

	http.ListenAndServe(":8000", r)
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

	return allRecords, nil
}

// processTar обрабатывает tar-архив
func processTar(file io.Reader) ([][]string, error) {
	reader := tar.NewReader(file)

	var allRecords [][]string
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(header.Name, ".csv") {
			records, err := readCSV(reader)
			if err != nil {
				return nil, err
			}
			allRecords = append(allRecords, records...)
		}
	}

	return allRecords, nil
}

// readCSV читает CSV-данные из io.Reader
func readCSV(r io.Reader) ([][]string, error) {
	reader := csv.NewReader(r)
	return reader.ReadAll()
}
