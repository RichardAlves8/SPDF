package pcppdf

import (
	"encoding/csv"
	"io"
	"os"
	"strings"
)

// LoadMetadataCSV lê um arquivo CSV contendo informações detalhadas dos produtos.
// Este arquivo é essencial para o "enrichment" (enriquecimento), adicionando dados
// como EAN13, Seção e Subgrupo que não existem na planilha de programação original.
func LoadMetadataCSV(path string) (map[string]ProdutoInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Configura o leitor CSV para usar ponto-e-vírgula como delimitador,
	// padrão comum em exportações de ERPs brasileiros.
	reader := csv.NewReader(f)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1 // Permite número variável de campos se necessário.

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
		
		// Pula a primeira linha (cabeçalho: CODIGO;PRODUTO;EAN13;SECAO;SUBGRUPO).
		if first {
			first = false
			continue
		}
		
		// Garante que a linha possui os campos mínimos necessários.
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
		
		// Armazena no mapa usando o código como chave para busca rápida (O(1)).
		meta[info.Codigo] = info
	}
	return meta, nil
}
