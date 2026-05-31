package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pcppdf "spdf-pcppdf-go/pkg/pcppdf"

	"github.com/xuri/excelize/v2"
)

func main() {
	var (
		pcpPath       = flag.String("xls", "", "Caminho do arquivo .xlsx/.xlsm")
		metaCSV       = flag.String("meta", "../resources/meta/dados.CSV", "CSV de Meta dados")
		templateHTML  = flag.String("template", "../resources/templates/template.html", "Template HTML")
		outDir        = flag.String("out", "out/", "Pasta de saida")
		sheetName     = flag.String("sheet", "RELATÓRIO PP GERAL", "Nome da aba")
		dateCols      = flag.String("date-cols", "12,15,18,21,24,27,30,33,35,37,39", "Colunas de datas/quantidades (ex: 5,6,7)")
		codeCol       = flag.Int("code-col", 2, "Coluna do codigo")
		descCol       = flag.Int("desc-col", 3, "Coluna da descricao")
		startLine     = flag.Int("start-line", 5, "Fim do cabeçalho/Inicio dos dados")
		dateLine      = flag.Int("date-line", 0, "Linha de datas")
		inclusionPath = flag.String("inclusion-out", "", "Pasta de inclusao")
		inspect       = flag.Bool("inspect", false, "Printa (Inspeciona) as primeiras linhas e colunas passadas nas flags/args")
		inspectRows   = flag.Int("inspect-rows", 8, "Linhas para inspecionar")
		inspectCols   = flag.Int("inspect-cols", 20, "Colunas para inspecionar")
		verbose       = flag.Bool("verbose", false, "Exibe todo resultado de processamento")
		enableLog     = flag.Bool("log", false, "Loga estado de processamento")
		logDir        = flag.String("log-dir", "logs", "diretorio do arquivo")
	)
	flag.Parse()

	if *pcpPath == "" {
		exitf("Caminho do arquivo não informado <flag: -pcp>")
	}
	if *inspect {
		if err := inspectSheet(*pcpPath, *sheetName, *inspectRows, *inspectCols); err != nil {
			exitf("Erro ao inspecionar: %v", err)
		}
		return
	}

	// Monta o writer composto antes de qualquer saída de dados.
	w, closeLog, err := buildWriter(*verbose, *enableLog, *logDir)
	if err != nil {
		exitf("Erro ao configurar log: %v", err)
	}
	defer closeLog()

	cols, err := parseColumns(*dateCols)
	if err != nil {
		exitf("Erro ao processar date-cols: %v", err)
	}
	if len(cols) == 0 {
		exitf("Colunas de datas não informado <flag: -date-cols c1, c2, c3>")
	}

	cfg := pcppdf.Config{
		TemplateHTML:     *templateHTML,
		MetaCSVFile:      *metaCSV,
		PCPTabName:       *sheetName,
		ListColumnData:   cols,
		IndexColumnCod:   *codeCol,
		IndexColItemDesc: *descCol,
		IndexStartLine:   *startLine,
		IndexLineDate:    *dateLine,
		FolderOPPCP:      *outDir,
		FolderOPInclusao: *inclusionPath,
	}

	processor, err := pcppdf.NewProcessor(cfg, pcppdf.WithRenderer(pcppdf.GoPDFRenderer{
		FontDir: filepath.Dir(*templateHTML),
	}))
	if err != nil {
		exitf("Não foi possível inciar <processor> : %v", err)
	}

	result, err := processor.ProcessFile(context.Background(), *pcpPath)
	if err != nil {
		exitf("Não foi possivel processar o arquivo <xls/xlsm> : %v", err)
	}

	if *verbose || *enableLog {
		fmt.Fprintln(w, "\nDados processados:")
		if len(result.GeneratedFiles) == 0 && len(result.MissingItems) == 0 {
			fmt.Fprintln(w, " Dados não encontrado.")
		}
		for _, file := range result.GeneratedFiles {
			fmt.Fprintf(w, " OK  | %s - %s - %.2f - %s\n", file.Codigo, file.Produto, file.Quantidade, file.Data)
		}
		for _, item := range result.MissingItems {
			fmt.Fprintf(w, "[CSV] ERR | %s - %s - %.2f - %s | Item não encontrado\n", item.Codigo, item.Produto, item.Quantidade, item.Data)
		}

		if len(result.GeneratedFiles) > 0 {
			for number, file := range result.GeneratedFiles {
				fmt.Fprintf(w, "\n[PDF | %v]: %v ", number,file)
			}
		} else {
			fmt.Fprintln(w, "Não foi possivel encontrar nenhum arquivo gerado pós processamento!")
		}
	}

	// Resumo vai para o writer E para stdout sempre.
	writers := []io.Writer{os.Stdout}
	if *enableLog && w != os.Stdout {
		writers = append(writers, w)
	}
	summary := io.MultiWriter(writers...)

	fmt.Fprintln(summary, "\nResumo:")
	fmt.Fprintf(summary, "Calculados para processar: %d\n", result.LinesWithProduction)
	fmt.Fprintf(summary, "Total: %d\n", result.OrdersProcessed)
	fmt.Fprintf(summary, "PDF: %d\n", result.FormsGenerated)
	fmt.Fprintf(summary, "Não encontrado: %d\n", len(result.MissingProducts))
}

// buildWriter constrói um io.Writer de acordo com as flags -verbose e -log.
// Retorna o writer, uma função para fechar o arquivo (se houver) e um erro
func buildWriter(verbose, enableLog bool, logDir string) (io.Writer, func(), error) {
	noop := func() {}

	if !verbose && !enableLog {
		return io.Discard, noop, nil
	}
	if verbose && !enableLog {
		return os.Stdout, noop, nil
	}

	// Log está ativo: cria pasta e arquivo.
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, noop, fmt.Errorf("criar pasta %q: %w", logDir, err)
	}
	timestamp := time.Now().Format("20060102150405") // YYYYMMDDHHMMSS
	logPath := filepath.Join(logDir, timestamp+".log")

	f, err := os.Create(logPath)
	if err != nil {
		return nil, noop, fmt.Errorf("criar arquivo de log %q: %w", logPath, err)
	}
	close := func() { f.Close() }

	if verbose {
		return io.MultiWriter(os.Stdout, f), close, nil
	}
	return f, close, nil
}

// ----- helpers -----

func parseColumns(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	cols := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		cols = append(cols, n)
	}
	return cols, nil
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func inspectSheet(path string, sheet string, maxRows int, maxCols int) error {
	workbook, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer workbook.Close()

	rows, err := workbook.Rows(sheet)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("Aba: %s\n", sheet)
	fmt.Println("Com idx base zero")

	rowIdx := 0
	for rows.Next() {
		if rowIdx >= maxRows {
			break
		}
		row, err := rows.Columns()
		if err != nil {
			return err
		}

		fmt.Printf("\nLinha %d:\n", rowIdx)
		for colIdx := 0; colIdx < len(row) && colIdx < maxCols; colIdx++ {
			value := strings.TrimSpace(row[colIdx])
			if value == "" {
				continue
			}
			fmt.Printf("  [%02d] %s\n", colIdx, value)
		}
		rowIdx++
	}

	return rows.Error()
}