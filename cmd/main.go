// Comando spdf gera as fichas de ordem de produção em PDF a partir de uma
// planilha de programação (.xlsx ou .xlsm) e de um CSV de metadados dos
// produtos.
//
// Com a flag -inspect, o comando não gera fichas: ele apenas lista as
// primeiras linhas e colunas da planilha, para descobrir os índices que as
// demais flags esperam.
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

// main lê as flags, valida a entrada e executa o modo de inspeção ou o
// processamento da planilha.
func main() {
	// As flags permitem adaptar o SPDF a diferentes layouts de planilha.
	// Seus nomes são a interface externa do comando e devem ser mantidos.
	var (
		caminhoPlanilha  = flag.String("xls", "", "Caminho do arquivo .xlsx/.xlsm (Obrigatório)")
		caminhoMetadados = flag.String("meta", "../resources/meta/dados.CSV", "Caminho do CSV de Metadados")
		caminhoTemplate  = flag.String("template", "../resources/templates/template.html", "Caminho do Template base")
		pastaSaida       = flag.String("out", "out/", "Pasta de saída para os PDFs")
		nomeAba          = flag.String("sheet", "RELATÓRIO PP GERAL", "Nome da aba da planilha")
		colunasDeData    = flag.String("date-cols", "12,15,18,21,24,27,30,33,35,37,39", "Índices das colunas de datas (separados por vírgula)")
		colunaCodigo     = flag.Int("code-col", 2, "Índice da coluna do código do produto")
		colunaDescricao  = flag.Int("desc-col", 3, "Índice da coluna da descrição")
		linhaInicial     = flag.Int("start-line", 5, "Linha onde começam os dados (ignora cabeçalho)")
		linhaDasDatas    = flag.Int("date-line", 0, "Linha onde estão os títulos das datas")
		pastaInclusao    = flag.String("inclusion-out", "", "Pasta de saída para ordens de inclusão")
		inspecionar      = flag.Bool("inspect", false, "Modo Inspeção: apenas lista as colunas da planilha no terminal")
		linhasInspecao   = flag.Int("inspect-rows", 8, "Qtd de linhas para inspecionar")
		colunasInspecao  = flag.Int("inspect-cols", 20, "Qtd de colunas para inspecionar")
		detalhado        = flag.Bool("verbose", false, "Exibe detalhes do processamento no terminal")
		gravarLog        = flag.Bool("log", false, "Habilita a gravação de logs em arquivo")
		pastaLog         = flag.String("log-dir", "logs", "Diretório para salvar os arquivos de log")
	)
	flag.Parse()

	if *caminhoPlanilha == "" {
		sairComErro("Erro: Caminho do arquivo não informado. Use a flag -xls")
	}

	// O modo de inspeção ajuda a descobrir os índices das colunas sem abrir o Excel.
	if *inspecionar {
		if err := inspecionarPlanilha(*caminhoPlanilha, *nomeAba, *linhasInspecao, *colunasInspecao); err != nil {
			sairComErro("Erro ao inspecionar: %v", err)
		}
		return
	}

	saida, fecharLog, err := montarSaida(*detalhado, *gravarLog, *pastaLog)
	if err != nil {
		sairComErro("Erro ao configurar log: %v", err)
	}
	defer fecharLog()

	colunas, err := lerColunas(*colunasDeData)
	if err != nil {
		sairComErro("Erro ao processar date-cols: %v", err)
	}
	if len(colunas) == 0 {
		sairComErro("Erro: Nenhuma coluna de data informada. Use -date-cols c1,c2...")
	}

	configuracao := pcppdf.Configuracao{
		CaminhoTemplateHTML:  *caminhoTemplate,
		CaminhoCSVMetadados:  *caminhoMetadados,
		NomeAbaPCP:           *nomeAba,
		ColunasDeData:        colunas,
		ColunaCodigo:         *colunaCodigo,
		ColunaDescricao:      *colunaDescricao,
		LinhaInicialDosDados: *linhaInicial,
		LinhaDasDatas:        *linhaDasDatas,
		PastaOrdensPCP:       *pastaSaida,
		PastaOrdensInclusao:  *pastaInclusao,
	}

	// As fontes ficam na mesma pasta do template.
	processador, err := pcppdf.NovoProcessador(configuracao, pcppdf.ComRenderizador(pcppdf.RenderizadorGoPDF{
		PastaFontes: filepath.Dir(*caminhoTemplate),
	}))
	if err != nil {
		sairComErro("Não foi possível iniciar o processador: %v", err)
	}

	resultado, err := processador.ProcessarArquivo(context.Background(), *caminhoPlanilha)
	if err != nil {
		sairComErro("Erro durante o processamento: %v", err)
	}

	if *detalhado || *gravarLog {
		fmt.Fprintln(saida, "\n--- Detalhes do Processamento ---")
		if len(resultado.ArquivosGerados) == 0 && len(resultado.ItensAusentes) == 0 {
			fmt.Fprintln(saida, "Nenhum dado produtivo encontrado.")
		}
		for _, arquivo := range resultado.ArquivosGerados {
			fmt.Fprintf(saida, "[OK]  | %s - %s - %.2f - %s\n", arquivo.Codigo, arquivo.Produto, arquivo.Quantidade, arquivo.Data)
		}
		for _, item := range resultado.ItensAusentes {
			fmt.Fprintf(saida, "[ERR] | %s - %s - %.2f - %s | Metadados não encontrados\n", item.Codigo, item.Produto, item.Quantidade, item.Data)
		}
	}

	// O resumo final sempre vai ao terminal e, com o log ativo, também ao arquivo.
	destinos := []io.Writer{os.Stdout}
	if *gravarLog && saida != os.Stdout {
		destinos = append(destinos, saida)
	}
	resumo := io.MultiWriter(destinos...)

	fmt.Fprintln(resumo, "\n--- Resumo Final ---")
	fmt.Fprintf(resumo, "Linhas com produção identificadas: %d\n", resultado.LinhasComProducao)
	fmt.Fprintf(resumo, "Total de ordens individuais: %d\n", resultado.OrdensProcessadas)
	fmt.Fprintf(resumo, "Fichas PDF geradas com sucesso: %d\n", resultado.FichasGeradas)
	fmt.Fprintf(resumo, "Produtos ausentes no cadastro (CSV): %d\n", len(resultado.ProdutosAusentes))
}

// montarSaida decide para onde vai a saída detalhada do programa e devolve,
// além do destino, a função que deve ser chamada ao final para fechá-lo.
//
// Sem verbose e sem log, a saída é descartada. Só com verbose, vai ao terminal.
// Com log, vai a um arquivo com a data e a hora no nome, dentro de pastaLog,
// e também ao terminal se verbose estiver ativo.
func montarSaida(detalhado, gravarLog bool, pastaLog string) (io.Writer, func(), error) {
	semAcao := func() {}

	if !detalhado && !gravarLog {
		return io.Discard, semAcao, nil
	}
	if detalhado && !gravarLog {
		return os.Stdout, semAcao, nil
	}

	if err := os.MkdirAll(pastaLog, 0o755); err != nil {
		return nil, semAcao, fmt.Errorf("erro ao criar pasta de log: %w", err)
	}
	instante := time.Now().Format("20060102150405")
	caminhoLog := filepath.Join(pastaLog, instante+".log")

	arquivo, err := os.Create(caminhoLog)
	if err != nil {
		return nil, semAcao, fmt.Errorf("erro ao criar arquivo de log: %w", err)
	}
	fechar := func() { arquivo.Close() }

	if detalhado {
		return io.MultiWriter(os.Stdout, arquivo), fechar, nil
	}
	return arquivo, fechar, nil
}

// lerColunas transforma um texto como "1,2,3" em []int{1, 2, 3}. Texto vazio
// resulta em lista vazia, sem erro.
func lerColunas(texto string) ([]int, error) {
	texto = strings.TrimSpace(texto)
	if texto == "" {
		return nil, nil
	}
	partes := strings.Split(texto, ",")
	colunas := make([]int, 0, len(partes))
	for _, parte := range partes {
		numero, err := strconv.Atoi(strings.TrimSpace(parte))
		if err != nil {
			return nil, err
		}
		colunas = append(colunas, numero)
	}
	return colunas, nil
}

// sairComErro escreve a mensagem formatada na saída de erro e encerra o
// programa com o código 1.
func sairComErro(formato string, argumentos ...any) {
	fmt.Fprintf(os.Stderr, formato+"\n", argumentos...)
	os.Exit(1)
}

// inspecionarPlanilha abre a aba indicada e imprime as primeiras colunas das
// primeiras linhas, sem as células vazias. Os índices impressos são os mesmos
// que as flags -code-col, -desc-col e -date-cols esperam.
func inspecionarPlanilha(caminho string, aba string, maximoLinhas int, maximoColunas int) error {
	arquivoExcel, err := excelize.OpenFile(caminho)
	if err != nil {
		return err
	}
	defer arquivoExcel.Close()

	linhas, err := arquivoExcel.Rows(aba)
	if err != nil {
		return err
	}
	defer linhas.Close()

	fmt.Printf("Inspecionando Aba: %s (Índices base zero)\n", aba)

	indiceLinha := 0
	for linhas.Next() {
		if indiceLinha >= maximoLinhas {
			break
		}
		linha, err := linhas.Columns()
		if err != nil {
			return err
		}

		fmt.Printf("\n[Linha %d]:\n", indiceLinha)
		for indiceColuna := 0; indiceColuna < len(linha) && indiceColuna < maximoColunas; indiceColuna++ {
			valor := strings.TrimSpace(linha[indiceColuna])
			if valor == "" {
				continue
			}
			fmt.Printf("  col %02d -> %s\n", indiceColuna, valor)
		}
		indiceLinha++
	}

	return linhas.Error()
}
