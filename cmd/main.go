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
	// Definição das flags de linha de comando.
	// Estes parâmetros permitem adaptar o SPDF a diferentes layouts de planilhas.
	var (
		pcpPath       = flag.String("xls", "", "Caminho do arquivo .xlsx/.xlsm (Obrigatório)")
		metaCSV       = flag.String("meta", "../resources/meta/dados.CSV", "Caminho do CSV de Metadados")
		templateHTML  = flag.String("template", "../resources/templates/template.html", "Caminho do Template base")
		outDir        = flag.String("out", "out/", "Pasta de saída para os PDFs")
		sheetName     = flag.String("sheet", "RELATÓRIO PP GERAL", "Nome da aba da planilha")
		dateCols      = flag.String("date-cols", "12,15,18,21,24,27,30,33,35,37,39", "Índices das colunas de datas (separados por vírgula)")
		codeCol       = flag.Int("code-col", 2, "Índice da coluna do código do produto")
		descCol       = flag.Int("desc-col", 3, "Índice da coluna da descrição")
		startLine     = flag.Int("start-line", 5, "Linha onde começam os dados (ignora cabeçalho)")
		dateLine      = flag.Int("date-line", 0, "Linha onde estão os títulos das datas")
		inclusionPath = flag.String("inclusion-out", "", "Pasta de saída para ordens de inclusão")
		inspect       = flag.Bool("inspect", false, "Modo Inspeção: apenas lista as colunas da planilha no terminal")
		inspectRows   = flag.Int("inspect-rows", 8, "Qtd de linhas para inspecionar")
		inspectCols   = flag.Int("inspect-cols", 20, "Qtd de colunas para inspecionar")
		verbose       = flag.Bool("verbose", false, "Exibe detalhes do processamento no terminal")
		enableLog     = flag.Bool("log", false, "Habilita a gravação de logs em arquivo")
		logDir        = flag.String("log-dir", "logs", "Diretório para salvar os arquivos de log")
	)
	flag.Parse()

	// Validação básica de entrada.
	if *pcpPath == "" {
		exitf("Erro: Caminho do arquivo não informado. Use a flag -xls")
	}

	// Modo Inspeção: Útil para descobrir os índices das colunas sem abrir o Excel.
	if *inspect {
		if err := inspectSheet(*pcpPath, *sheetName, *inspectRows, *inspectCols); err != nil {
			exitf("Erro ao inspecionar: %v", err)
		}
		return
	}

	// Configura o sistema de saída (Terminal + Arquivo de Log).
	w, closeLog, err := buildWriter(*verbose, *enableLog, *logDir)
	if err != nil {
		exitf("Erro ao configurar log: %v", err)
	}
	defer closeLog()

	// Converte a string de colunas (ex: "12,15") em um slice de inteiros.
	cols, err := parseColumns(*dateCols)
	if err != nil {
		exitf("Erro ao processar date-cols: %v", err)
	}
	if len(cols) == 0 {
		exitf("Erro: Nenhuma coluna de data informada. Use -date-cols c1,c2...")
	}

	// Monta a struct de configuração para o pacote pcppdf.
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

	// Inicializa o Processor com o Renderer oficial (GoPDFRenderer).
	processor, err := pcppdf.NewProcessor(cfg, pcppdf.WithRenderer(pcppdf.GoPDFRenderer{
		FontDir: filepath.Dir(*templateHTML),
	}))
	if err != nil {
		exitf("Não foi possível iniciar o processador: %v", err)
	}

	// Executa o processamento principal do arquivo.
	result, err := processor.ProcessFile(context.Background(), *pcpPath)
	if err != nil {
		exitf("Erro durante o processamento: %v", err)
	}

	// Exibe detalhes se o modo verbose ou log estiverem ativos.
	if *verbose || *enableLog {
		fmt.Fprintln(w, "\n--- Detalhes do Processamento ---")
		if len(result.GeneratedFiles) == 0 && len(result.MissingItems) == 0 {
			fmt.Fprintln(w, "Nenhum dado produtivo encontrado.")
		}
		for _, file := range result.GeneratedFiles {
			fmt.Fprintf(w, "[OK]  | %s - %s - %.2f - %s\n", file.Codigo, file.Produto, file.Quantidade, file.Data)
		}
		for _, item := range result.MissingItems {
			fmt.Fprintf(w, "[ERR] | %s - %s - %.2f - %s | Metadados não encontrados\n", item.Codigo, item.Produto, item.Quantidade, item.Data)
		}
	}

	// Prepara o resumo final para o usuário.
	writers := []io.Writer{os.Stdout}
	if *enableLog && w != os.Stdout {
		writers = append(writers, w)
	}
	summary := io.MultiWriter(writers...)

	fmt.Fprintln(summary, "\n--- Resumo Final ---")
	fmt.Fprintf(summary, "Linhas com produção identificadas: %d\n", result.LinesWithProduction)
	fmt.Fprintf(summary, "Total de ordens individuais: %d\n", result.OrdersProcessed)
	fmt.Fprintf(summary, "Fichas PDF geradas com sucesso: %d\n", result.FormsGenerated)
	fmt.Fprintf(summary, "Produtos ausentes no cadastro (CSV): %d\n", len(result.MissingProducts))
}

// buildWriter decide para onde a saída do programa deve ir.
func buildWriter(verbose, enableLog bool, logDir string) (io.Writer, func(), error) {
	noop := func() {}

	// Se nada estiver ativo, descarta a saída.
	if !verbose && !enableLog {
		return io.Discard, noop, nil
	}
	// Apenas terminal.
	if verbose && !enableLog {
		return os.Stdout, noop, nil
	}

	// Se o log em arquivo estiver ativo, cria a pasta e o arquivo com timestamp.
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, noop, fmt.Errorf("erro ao criar pasta de log: %w", err)
	}
	timestamp := time.Now().Format("20060102150405")
	logPath := filepath.Join(logDir, timestamp+".log")

	f, err := os.Create(logPath)
	if err != nil {
		return nil, noop, fmt.Errorf("erro ao criar arquivo de log: %w", err)
	}
	close := func() { f.Close() }

	// Se verbose estiver ativo, duplica a saída (Terminal + Arquivo).
	if verbose {
		return io.MultiWriter(os.Stdout, f), close, nil
	}
	return f, close, nil
}

// parseColumns transforma "1,2,3" em []int{1, 2, 3}.
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

// exitf exibe um erro no stderr e encerra o programa.
func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// inspectSheet abre a planilha e imprime as primeiras colunas de cada linha.
// Fundamental para o usuário mapear os índices das colunas corretamente.
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

	fmt.Printf("Inspecionando Aba: %s (Índices base zero)\n", sheet)

	rowIdx := 0
	for rows.Next() {
		if rowIdx >= maxRows {
			break
		}
		row, err := rows.Columns()
		if err != nil {
			return err
		}

		fmt.Printf("\n[Linha %d]:\n", rowIdx)
		for colIdx := 0; colIdx < len(row) && colIdx < maxCols; colIdx++ {
			value := strings.TrimSpace(row[colIdx])
			if value == "" {
				continue
			}
			fmt.Printf("  col %02d -> %s\n", colIdx, value)
		}
		rowIdx++
	}

	return rows.Error()
}
