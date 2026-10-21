package pcppdf

import (
	"encoding/csv"
	"io"
	"os"
	"strings"
)

// CarregarCSVMetadados lê o CSV com as informações detalhadas dos produtos e
// devolve um mapa indexado pelo código do produto.
//
// Os metadados completam a planilha de programação com dados que ela não tem:
// EAN13, seção e subgrupo. O arquivo deve usar ponto e vírgula como separador
// e ter o cabeçalho CODIGO;PRODUTO;EAN13;SECAO;SUBGRUPO. A primeira linha é
// sempre descartada, e as linhas com menos de cinco campos são ignoradas. Se
// um código aparecer mais de uma vez, vale a última ocorrência.
func CarregarCSVMetadados(caminho string) (map[string]InformacaoProduto, error) {
	arquivo, err := os.Open(caminho)
	if err != nil {
		return nil, err
	}
	defer arquivo.Close()

	// O ponto e vírgula é o separador comum nas exportações de ERPs brasileiros.
	leitor := csv.NewReader(arquivo)
	leitor.Comma = ';'
	// Aceita um número variável de campos por linha.
	leitor.FieldsPerRecord = -1

	metadados := make(map[string]InformacaoProduto)
	primeiraLinha := true
	for {
		registro, err := leitor.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// A primeira linha é o cabeçalho.
		if primeiraLinha {
			primeiraLinha = false
			continue
		}

		if len(registro) < 5 {
			continue
		}

		informacao := InformacaoProduto{
			Codigo:   strings.TrimSpace(registro[0]),
			Nome:     strings.TrimSpace(registro[1]),
			EAN13:    strings.TrimSpace(registro[2]),
			Secao:    strings.TrimSpace(registro[3]),
			Subgrupo: strings.TrimSpace(registro[4]),
		}

		// O código é a chave do mapa para que a busca seja direta, em tempo constante.
		metadados[informacao.Codigo] = informacao
	}
	return metadados, nil
}
