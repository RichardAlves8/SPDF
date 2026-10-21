package pcppdf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestProcessadorProcessarArquivo percorre o fluxo completo com uma planilha
// de duas linhas: um produto cadastrado e outro ausente do CSV. Usa o
// RenderizadorArquivo para que o conteúdo gerado possa ser lido como texto.
func TestProcessadorProcessarArquivo(t *testing.T) {
	pasta := t.TempDir()
	caminhoMetadados := filepath.Join(pasta, "dados.csv")
	if err := os.WriteFile(caminhoMetadados, []byte("CODIGO;PRODUTO;EAN13;SECAO;SUBGRUPO\n123-4;Produto Meta;789;17;Pratos\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	caminhoTemplate := filepath.Join(pasta, "template.html")
	if err := os.WriteFile(caminhoTemplate, []byte("{{CODIGO}} {{PRODUTO}}"), 0o644); err != nil {
		t.Fatal(err)
	}

	caminhoPlanilha := filepath.Join(pasta, "pcp.xlsx")
	planilha := excelize.NewFile()
	const aba = "PCP"
	planilha.SetSheetName("Sheet1", aba)
	definirCelulaOuFalhar(t, planilha, aba, "C1", "29/05/2026")
	definirCelulaOuFalhar(t, planilha, aba, "A2", "123-4")
	definirCelulaOuFalhar(t, planilha, aba, "B2", "Produto Planilha")
	definirCelulaOuFalhar(t, planilha, aba, "C2", 10)
	definirCelulaOuFalhar(t, planilha, aba, "A3", "999-9")
	definirCelulaOuFalhar(t, planilha, aba, "B3", "Produto Ausente")
	definirCelulaOuFalhar(t, planilha, aba, "C3", 5)
	if err := planilha.SaveAs(caminhoPlanilha); err != nil {
		t.Fatal(err)
	}

	configuracao := Configuracao{
		CaminhoTemplateHTML:  caminhoTemplate,
		CaminhoCSVMetadados:  caminhoMetadados,
		NomeAbaPCP:           aba,
		ColunasDeData:        []int{2},
		ColunaCodigo:         0,
		ColunaDescricao:      1,
		LinhaInicialDosDados: 1,
		PastaOrdensPCP:       filepath.Join(pasta, "out"),
	}
	processador, err := NovoProcessador(configuracao, ComRenderizador(RenderizadorArquivo{CaminhoTemplate: caminhoTemplate}))
	if err != nil {
		t.Fatal(err)
	}

	resultado, err := processador.ProcessarArquivo(context.Background(), caminhoPlanilha)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.FichasGeradas != 1 {
		t.Fatalf("FichasGeradas = %d, esperado 1", resultado.FichasGeradas)
	}
	if resultado.LinhasComProducao != 2 {
		t.Fatalf("LinhasComProducao = %d, esperado 2", resultado.LinhasComProducao)
	}
	if len(resultado.ProdutosAusentes) != 1 || resultado.ProdutosAusentes[0] != "999-9" {
		t.Fatalf("ProdutosAusentes inesperado: %#v", resultado.ProdutosAusentes)
	}
	if len(resultado.ArquivosGerados) != 1 {
		t.Fatalf("ArquivosGerados = %d, esperado 1", len(resultado.ArquivosGerados))
	}
	if _, err := os.Stat(resultado.ArquivosGerados[0].Caminho); err != nil {
		t.Fatalf("arquivo gerado nao encontrado: %v", err)
	}
	caminhoEsperado := filepath.Join(configuracao.PastaOrdensPCP, "29.05.2026", "123-4 [29.05.2026].pdf")
	if resultado.ArquivosGerados[0].Caminho != caminhoEsperado {
		t.Fatalf("Caminho = %q, esperado %q", resultado.ArquivosGerados[0].Caminho, caminhoEsperado)
	}
	conteudo, err := os.ReadFile(resultado.ArquivosGerados[0].Caminho)
	if err != nil {
		t.Fatal(err)
	}
	if string(conteudo) != "123-4 Produto Planilha" {
		t.Fatalf("conteudo renderizado inesperado: %q", string(conteudo))
	}
}

// definirCelulaOuFalhar grava um valor na célula e interrompe o teste se a
// gravação falhar.
func definirCelulaOuFalhar(t *testing.T, planilha *excelize.File, aba string, celula string, valor any) {
	t.Helper()
	if err := planilha.SetCellValue(aba, celula, valor); err != nil {
		t.Fatal(err)
	}
}
