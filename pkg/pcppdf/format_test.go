package pcppdf

import "testing"

// TestFormatarDataPTBR verifica se a data ganha o dia da semana em português.
func TestFormatarDataPTBR(t *testing.T) {
	obtido := formatarDataPTBR("29/05/2026")
	esperado := "Sexta — 29/05/2026"
	if obtido != esperado {
		t.Fatalf("obtido %q, esperado %q", obtido, esperado)
	}
}

// TestPastaDeDataParaCaminho verifica se as barras da data viram pontos e o
// dia da semana é preservado.
func TestPastaDeDataParaCaminho(t *testing.T) {
	obtido := pastaDeDataParaCaminho("Sexta — 29/05/2026")
	esperado := "Sexta — 29.05.2026"
	if obtido != esperado {
		t.Fatalf("obtido %q, esperado %q", obtido, esperado)
	}
}

// TestNormalizarCodigo verifica se o sufixo ".0" do Excel é removido e o hífen
// do dígito verificador é mantido.
func TestNormalizarCodigo(t *testing.T) {
	obtido := normalizarCodigo("123-4.0")
	esperado := "123-4"
	if obtido != esperado {
		t.Fatalf("obtido %q, esperado %q", obtido, esperado)
	}
}
