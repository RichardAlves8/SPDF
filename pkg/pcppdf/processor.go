package pcppdf

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

// Processor é o motor principal responsável por ler a planilha de programação,
// extrair os dados de produção e orquestrar a geração das fichas em PDF.
type Processor struct {
	cfg      Config
	meta     map[string]ProdutoInfo
	renderer Renderer
}

// NewProcessor inicializa um novo processador com as configurações fornecidas.
// Ele valida a configuração e garante que os metadados (CSV) sejam carregados.
func NewProcessor(cfg Config, opts ...Option) (*Processor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	p := &Processor{
		cfg:      cfg,
		renderer: FileRenderer{}, // Renderer padrão (HTML) caso não seja especificado outro.
	}
	
	// Aplica opções funcionais (WithRenderer, WithMetadata, etc).
	for _, opt := range opts {
		opt(p)
	}
	
	if p.renderer == nil {
		return nil, ErrNilRenderer
	}
	
	// Carrega metadados do CSV se eles ainda não foram injetados.
	if p.meta == nil {
		meta, err := LoadMetadataCSV(cfg.MetaCSVFile)
		if err != nil {
			return nil, err
		}
		p.meta = meta
	}
	return p, nil
}

// SetRenderer permite trocar o motor de renderização em tempo de execução.
func (p *Processor) SetRenderer(r Renderer) {
	p.renderer = r
}

// ProcessFile é o ponto central de execução. Ele abre a planilha, itera sobre as linhas
// usando streaming (para economizar memória) e envia cada item válido para o Emitter.
func (p *Processor) ProcessFile(ctx context.Context, planningPath string) (*ProcessResult, error) {
	// Verifica se o arquivo existe antes de tentar abrir.
	if _, err := os.Stat(planningPath); err != nil {
		return nil, err
	}

	// Abre o workbook (arquivo .xlsx / .xlsm).
	workbook, err := excelize.OpenFile(planningPath)
	if err != nil {
		return nil, err
	}
	defer workbook.Close()

	// Valida se a aba informada existe.
	if idx, err := workbook.GetSheetIndex(p.cfg.PCPTabName); err != nil || idx < 0 {
		return nil, fmt.Errorf("aba PCP nao encontrada: %s", p.cfg.PCPTabName)
	}

	// Utiliza o iterador de linhas (Rows) em vez de carregar tudo na memória (GetRows).
	// Essencial para planilhas grandes e ambientes com pouca RAM.
	rows, err := workbook.Rows(p.cfg.PCPTabName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 1. Fase de Leitura: Extrai cabeçalhos de data e os itens que têm produção.
	dateRaw, dateFormatted, items, err := p.readPlanRows(rows)
	if err != nil {
		return nil, err
	}

	result := &ProcessResult{
		LinesWithProduction: len(items),
	}

	// 2. Fase de Emissão: Para cada item e cada data com quantidade > 0, gera uma Ordem de Produção.
	emitter := Emitter{
		cfg:      p.cfg,
		meta:     p.meta,
		renderer: p.renderer,
	}

	nowDate := time.Now().Format("02/01/2006")
	for _, item := range items {
		for _, col := range p.cfg.ListColumnData {
			qty := item.Quantities[col]
			if qty <= 0 {
				continue
			}
			
			result.OrdersProcessed++
			date := dateFormatted[col]
			if date == "" {
				date = dateRaw[col]
			}
			
			// Solicita ao Emitter a geração do documento (enriquecimento + renderização).
			generated, missing, err := emitter.Emit(ctx, EmitRequest{
				Codigo:        item.Codigo,
				Descricao:     item.Descricao,
				Quantidade:    qty,
				DataFormatada: date,
				DataEmissao:   nowDate,
				Numero:        result.FormsGenerated,
				Inclusao:      false,
			})
			if err != nil {
				return result, err
			}
			
			// Se o código do produto não estiver no meta-dados, logamos como item faltante.
			if missing != "" {
				result.MissingProducts = append(result.MissingProducts, missing)
				result.MissingItems = append(result.MissingItems, MissingItem{
					Codigo:     missing,
					Produto:    item.Descricao,
					Quantidade: qty,
					Data:       date,
				})
				continue
			}
			
			result.GeneratedFiles = append(result.GeneratedFiles, generated)
			result.FormsGenerated++
		}
	}

	return result, nil
}

// readPlanRows percorre a planilha uma única vez, identificando as datas e os produtos.
func (p *Processor) readPlanRows(rows *excelize.Rows) (map[int]string, map[int]string, []PlannedItem, error) {
	raw := make(map[int]string)
	formatted := make(map[int]string)
	var items []PlannedItem

	rowIdx := 0
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return nil, nil, nil, err
		}

		// Captura o cabeçalho de datas na linha configurada.
		if rowIdx == p.cfg.IndexLineDate {
			for _, col := range p.cfg.ListColumnData {
				v := rowValue(row, col)
				raw[col] = v
				formatted[col] = formatDatePTBR(v)
			}
		}

		// Processa os dados dos produtos a partir da linha de início.
		if rowIdx >= p.cfg.IndexStartLine {
			item, ok := p.plannedItemFromRow(row)
			if ok {
				items = append(items, item)
			}
		}

		rowIdx++
	}
	if err := rows.Error(); err != nil {
		return nil, nil, nil, err
	}
	return raw, formatted, items, nil
}

// plannedItemFromRow converte uma linha da planilha em um PlannedItem se houver produção nela.
func (p *Processor) plannedItemFromRow(row []string) (PlannedItem, bool) {
	quantities := make(map[int]float64)
	hasProduction := false
	
	// Extrai quantidades apenas das colunas de datas configuradas.
	for _, col := range p.cfg.ListColumnData {
		qty := parseFloat(rowValue(row, col))
		quantities[col] = qty
		if qty > 0 {
			hasProduction = true
		}
	}
	
	// Se a linha não tem nenhuma quantidade > 0, nós a ignoramos.
	if !hasProduction {
		return PlannedItem{}, false
	}
	
	return PlannedItem{
		Codigo:     normalizeCode(rowValue(row, p.cfg.IndexColumnCod)),
		Descricao:  rowValue(row, p.cfg.IndexColItemDesc),
		Quantities: quantities,
	}, true
}

// rowValue é um helper seguro para obter o valor de uma célula em um slice de strings.
func rowValue(row []string, colIdx int) string {
	if colIdx < 0 || colIdx >= len(row) {
		return ""
	}
	return row[colIdx]
}
