package pcppdf

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var weekdaysPTBR = map[time.Weekday]string{
	time.Sunday:    "Domingo",
	time.Monday:    "Segunda",
	time.Tuesday:   "Terça",
	time.Wednesday: "Quarta",
	time.Thursday:  "Quinta",
	time.Friday:    "Sexta",
	time.Saturday:  "Sábado",
}

func normalizeCode(v string) string {
	v = strings.TrimSpace(v)
	return strings.TrimSuffix(v, ".0")
}

func parseFloat(v string) float64 {
	v = strings.TrimSpace(v)
	if v == "" || v == "0" {
		return 0
	}
	v = strings.ReplaceAll(v, ",", ".")
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return f
}

func formatDatePTBR(raw string) string {
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

func dateFolderForPath(date string) string {
	return strings.ReplaceAll(strings.TrimSpace(date), "/", ".")
}

func valueString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
