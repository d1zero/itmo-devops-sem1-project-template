package helpers

import (
	"archive/zip"
	"encoding/csv"
	"io"
	"strings"
)

// ProcessZip обрабатывает zip-архив
func ProcessZip(file io.ReaderAt, size int64) ([][]string, error) {
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

// readCSV читает CSV-данные из io.Reader
func readCSV(r io.Reader) ([][]string, error) {
	reader := csv.NewReader(r)
	return reader.ReadAll()
}
