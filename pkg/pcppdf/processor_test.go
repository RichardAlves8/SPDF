package pcppdf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestProcessorProcessFile(t *testing.T) {
	dir := t.TempDir()
	metaPath := filepath.Join(dir, "dados.csv")
	if err := os.WriteFile(metaPath, []byte("CODIGO;PRODUTO;EAN13;SECAO;SUBGRUPO\n123-4;Produto Meta;789;17;Pratos\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	templatePath := filepath.Join(dir, "template.html")
	if err := os.WriteFile(templatePath, []byte("{{CODIGO}} {{PRODUTO}}"), 0o644); err != nil {
		t.Fatal(err)
	}

	xlsxPath := filepath.Join(dir, "pcp.xlsx")
	f := excelize.NewFile()
	const sheet = "PCP"
	f.SetSheetName("Sheet1", sheet)
	mustSetCell(t, f, sheet, "C1", "29/05/2026")
	mustSetCell(t, f, sheet, "A2", "123-4")
	mustSetCell(t, f, sheet, "B2", "Produto Planilha")
	mustSetCell(t, f, sheet, "C2", 10)
	mustSetCell(t, f, sheet, "A3", "999-9")
	mustSetCell(t, f, sheet, "B3", "Produto Ausente")
	mustSetCell(t, f, sheet, "C3", 5)
	if err := f.SaveAs(xlsxPath); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		TemplateHTML:     templatePath,
		MetaCSVFile:      metaPath,
		PCPTabName:       sheet,
		ListColumnData:   []int{2},
		IndexColumnCod:   0,
		IndexColItemDesc: 1,
		IndexStartLine:   1,
		FolderOPPCP:      filepath.Join(dir, "out"),
	}
	processor, err := NewProcessor(cfg, WithRenderer(FileRenderer{TemplatePath: templatePath}))
	if err != nil {
		t.Fatal(err)
	}

	result, err := processor.ProcessFile(context.Background(), xlsxPath)
	if err != nil {
		t.Fatal(err)
	}
	if result.FormsGenerated != 1 {
		t.Fatalf("FormsGenerated = %d, want 1", result.FormsGenerated)
	}
	if result.LinesWithProduction != 2 {
		t.Fatalf("LinesWithProduction = %d, want 2", result.LinesWithProduction)
	}
	if len(result.MissingProducts) != 1 || result.MissingProducts[0] != "999-9" {
		t.Fatalf("MissingProducts inesperado: %#v", result.MissingProducts)
	}
	if len(result.GeneratedFiles) != 1 {
		t.Fatalf("GeneratedFiles = %d, want 1", len(result.GeneratedFiles))
	}
	if _, err := os.Stat(result.GeneratedFiles[0].Path); err != nil {
		t.Fatalf("arquivo gerado nao encontrado: %v", err)
	}
	wantPath := filepath.Join(cfg.FolderOPPCP, "29.05.2026", "123-4 [29.05.2026].pdf")
	if result.GeneratedFiles[0].Path != wantPath {
		t.Fatalf("Path = %q, want %q", result.GeneratedFiles[0].Path, wantPath)
	}
	content, err := os.ReadFile(result.GeneratedFiles[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "123-4 Produto Planilha" {
		t.Fatalf("conteudo renderizado inesperado: %q", string(content))
	}
}

func mustSetCell(t *testing.T, f *excelize.File, sheet string, cell string, value any) {
	t.Helper()
	if err := f.SetCellValue(sheet, cell, value); err != nil {
		t.Fatal(err)
	}
}
