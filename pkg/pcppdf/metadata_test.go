package pcppdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMetadataCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dados.csv")
	content := "CODIGO;PRODUTO;EAN13;SECAO;SUBGRUPO\n123-4;Produto Teste;789;17;Pratos\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	meta, err := LoadMetadataCSV(path)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := meta["123-4"]
	if !ok {
		t.Fatal("produto nao carregado")
	}
	if got.Nome != "Produto Teste" || got.EAN13 != "789" {
		t.Fatalf("produto inesperado: %#v", got)
	}
}
