package pcppdf

import (
	"encoding/csv"
	"io"
	"os"
	"strings"
)

func LoadMetadataCSV(path string) (map[string]ProdutoInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1

	meta := make(map[string]ProdutoInfo)
	first := true
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if first {
			first = false
			continue
		}
		if len(record) < 5 {
			continue
		}
		info := ProdutoInfo{
			Codigo:   strings.TrimSpace(record[0]),
			Nome:     strings.TrimSpace(record[1]),
			EAN13:    strings.TrimSpace(record[2]),
			Secao:    strings.TrimSpace(record[3]),
			Subgrupo: strings.TrimSpace(record[4]),
		}
		meta[info.Codigo] = info
	}
	return meta, nil
}
