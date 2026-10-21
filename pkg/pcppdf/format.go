package pcppdf

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// diasDaSemanaPTBR traduz os dias da semana para o português do Brasil.
var diasDaSemanaPTBR = map[time.Weekday]string{
	time.Sunday:    "Domingo",
	time.Monday:    "Segunda",
	time.Tuesday:   "Terça",
	time.Wednesday: "Quarta",
	time.Thursday:  "Quinta",
	time.Friday:    "Sexta",
	time.Saturday:  "Sábado",
}

// normalizarCodigo remove os espaços das pontas e o sufixo decimal ".0" que o
// Excel costuma acrescentar a códigos numéricos.
func normalizarCodigo(valor string) string {
	valor = strings.TrimSpace(valor)
	return strings.TrimSuffix(valor, ".0")
}

// converterDecimal converte um texto numérico em float64, aceitando a vírgula
// como separador decimal, como no padrão brasileiro. Texto vazio, "0" ou
// inválido resulta em zero. Números com ponto como separador de milhar, como
// "1.500,50", não são reconhecidos e também resultam em zero.
func converterDecimal(valor string) float64 {
	valor = strings.TrimSpace(valor)
	if valor == "" || valor == "0" {
		return 0
	}
	// O Go espera o ponto como separador decimal.
	valor = strings.ReplaceAll(valor, ",", ".")
	numero, err := strconv.ParseFloat(valor, 64)
	if err != nil {
		return 0
	}
	return numero
}

// formatarDataPTBR transforma uma data como "29/05/2026" em um texto com o dia
// da semana, como "Sexta — 29/05/2026". O ponto também é aceito como
// separador. Se o texto não estiver no formato DD/MM/AAAA, ele é devolvido sem
// alteração, apenas sem os espaços das pontas.
func formatarDataPTBR(bruta string) string {
	bruta = strings.TrimSpace(strings.ReplaceAll(bruta, ".", "/"))
	if bruta == "" {
		return bruta
	}
	data, err := time.Parse("02/01/2006", bruta)
	if err != nil {
		return bruta
	}
	return fmt.Sprintf("%s — %s", diasDaSemanaPTBR[data.Weekday()], data.Format("02/01/2006"))
}

// dataNumericaParaCaminho mantém apenas os dígitos, os pontos e as barras da
// data, e troca as barras por pontos. O resultado serve para nomes de pasta e
// de arquivo, no padrão do projeto: "Sexta — 29/05/2026" vira "29.05.2026".
func dataNumericaParaCaminho(data string) string {
	var construtor strings.Builder
	for _, caractere := range data {
		if (caractere >= '0' && caractere <= '9') || caractere == '.' || caractere == '/' {
			if caractere == '/' {
				construtor.WriteRune('.')
			} else {
				construtor.WriteRune(caractere)
			}
		}
	}
	return construtor.String()
}

// pastaDeDataParaCaminho troca as barras da data por pontos, sem remover o
// restante do texto. Diferente de dataNumericaParaCaminho, ela preserva o dia
// da semana.
func pastaDeDataParaCaminho(data string) string {
	return strings.ReplaceAll(strings.TrimSpace(data), "/", ".")
}

// valorComoTexto converte qualquer valor em texto, tratando nil como texto
// vazio, para que a inserção de dados no template nunca falhe.
func valorComoTexto(valor any) string {
	if valor == nil {
		return ""
	}
	return fmt.Sprint(valor)
}
