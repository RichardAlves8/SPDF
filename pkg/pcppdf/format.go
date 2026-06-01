package pcppdf

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// weekdaysPTBR mapeia os dias da semana para o português brasileiro.
var weekdaysPTBR = map[time.Weekday]string{
	time.Sunday:    "Domingo",
	time.Monday:    "Segunda",
	time.Tuesday:   "Terça",
	time.Wednesday: "Quarta",
	time.Thursday:  "Quinta",
	time.Friday:    "Sexta",
	time.Saturday:  "Sábado",
}

// normalizeCode remove espaços em branco e o sufixo decimal ".0" que o Excel
// costuma adicionar a códigos numéricos.
func normalizeCode(v string) string {
	v = strings.TrimSpace(v)
	return strings.TrimSuffix(v, ".0")
}

// parseFloat converte strings numéricas para float64, tratando o padrão
// brasileiro de vírgula como separador decimal.
func parseFloat(v string) float64 {
	v = strings.TrimSpace(v)
	if v == "" || v == "0" {
		return 0
	}
	// Normaliza para o padrão do Go (ponto como separador).
	v = strings.ReplaceAll(v, ",", ".")
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return f
}

// formatDatePTBR transforma uma string de data (ex: 29/05/2026) em uma string
// amigável com o dia da semana (ex: Sexta — 29/05/2026).
func formatDatePTBR(raw string) string {
	// Algumas planilhas podem usar ponto como separador (29.05.2026).
	raw = strings.TrimSpace(strings.ReplaceAll(raw, ".", "/"))
	if raw == "" {
		return raw
	}
	t, err := time.Parse("02/01/2006", raw)
	if err != nil {
		return raw
	}
	return fmt.Sprintf("%s — %s", weekdaysPTBR[t.Weekday()], t.Format("02/01/2006"))
}

// numericDateForPath limpa a string de data para ser usada em nomes de pastas e arquivos,
// removendo textos e garantindo que o separador seja o ponto (padrão do projeto: 29.05.2026).
func numericDateForPath(date string) string {
	var b strings.Builder
	for _, r := range date {
		if (r >= '0' && r <= '9') || r == '.' || r == '/' {
			if r == '/' {
				b.WriteRune('.')
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// dateFolderForPath é um helper simples para normalizar caminhos de pasta baseados em data.
func dateFolderForPath(date string) string {
	return strings.ReplaceAll(strings.TrimSpace(date), "/", ".")
}

// valueString converte qualquer valor para string, tratando nil como string vazia.
// Útil para garantir que a injeção de dados no template não falhe.
func valueString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
