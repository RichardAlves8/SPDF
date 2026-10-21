package pcppdf

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCarregarCSVMetadados verifica se o cabeçalho é descartado e se os campos
// do produto são lidos nas colunas certas.
func TestCarregarCSVMetadados(t *testing.T) {
	pasta := t.TempDir()
	caminho := filepath.Join(pasta, "dados.csv")
	conteudo := "CODIGO;PRODUTO;EAN13;SECAO;SUBGRUPO\n123-4;Produto Teste;789;17;Pratos\n"
	if err := os.WriteFile(caminho, []byte(conteudo), 0o644); err != nil {
		t.Fatal(err)
	}

	metadados, err := CarregarCSVMetadados(caminho)
	if err != nil {
		t.Fatal(err)
	}
	obtido, encontrado := metadados["123-4"]
	if !encontrado {
		t.Fatal("produto nao carregado")
	}
	if obtido.Nome != "Produto Teste" || obtido.EAN13 != "789" {
		t.Fatalf("produto inesperado: %#v", obtido)
	}
}
