package helpers

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// ProcessZip обрабатывает zip-архив
func ProcessZip(file io.ReaderAt, size int64) ([][]string, error) {
	const op = "ProcessZip"

	reader, err := zip.NewReader(file, size)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", op, err.Error())
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

// readCSV читает CSV-данные из io.Reader
func readCSV(r io.Reader) ([][]string, error) {
	reader := csv.NewReader(r)
	return reader.ReadAll()
}

// WriteToZip Записывает строки в архив и возвращает его
func WriteToZip(data [][]string) (*os.File, int64, error) {
	const op = "WriteToZip"

	f, err := os.CreateTemp("", "data-*.csv")
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}
	defer os.Remove(f.Name())

	writer := csv.NewWriter(f)

	if err = writer.Write([]string{"id", "name", "category", "price", "create_date"}); err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	for _, d := range data {
		if err = writer.Write(d); err != nil {
			return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
		}
	}

	writer.Flush()
	if err = writer.Error(); err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	if _, err = f.Seek(0, 0); err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	zipFile, err := os.CreateTemp("", "data-*.zip")
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	zipWriter := zip.NewWriter(zipFile)

	csvInZip, err := zipWriter.Create("data.csv")
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	if _, err = io.Copy(csvInZip, f); err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	if err = zipWriter.Close(); err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	if _, err = zipFile.Seek(0, 0); err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	dataZip, err := zipFile.Stat()
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	return zipFile, dataZip.Size(), nil
}
