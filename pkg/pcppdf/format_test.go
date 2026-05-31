package pcppdf

import "testing"

func TestFormatDatePTBR(t *testing.T) {
	got := formatDatePTBR("29/05/2026")
	want := "Sexta — 29/05/2026"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDateFolderForPath(t *testing.T) {
	got := dateFolderForPath("Sexta — 29/05/2026")
	want := "Sexta — 29.05.2026"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNormalizeCode(t *testing.T) {
	got := normalizeCode("123-4.0")
	want := "123-4"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
