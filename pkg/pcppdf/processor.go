package pcppdf

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

type Processor struct {
	cfg      Config
	meta     map[string]ProdutoInfo
	renderer Renderer
}

func NewProcessor(cfg Config, opts ...Option) (*Processor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	p := &Processor{
		cfg:      cfg,
		renderer: FileRenderer{},
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.renderer == nil {
		return nil, ErrNilRenderer
	}
	if p.meta == nil {
		meta, err := LoadMetadataCSV(cfg.MetaCSVFile)
		if err != nil {
			return nil, err
		}
		p.meta = meta
	}
	return p, nil
}

func (p *Processor) SetRenderer(r Renderer) {
	p.renderer = r
}

func (p *Processor) ProcessFile(ctx context.Context, planningPath string) (*ProcessResult, error) {
	if _, err := os.Stat(planningPath); err != nil {
		return nil, err
	}

	workbook, err := excelize.OpenFile(planningPath)
	if err != nil {
		return nil, err
	}
	defer workbook.Close()

	if idx, err := workbook.GetSheetIndex(p.cfg.PCPTabName); err != nil || idx < 0 {
		return nil, fmt.Errorf("aba PCP nao encontrada: %s", p.cfg.PCPTabName)
	}

	rows, err := workbook.Rows(p.cfg.PCPTabName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dateRaw, dateFormatted, items, err := p.readPlanRows(rows)
	if err != nil {
		return nil, err
	}

	result := &ProcessResult{
		LinesWithProduction: len(items),
	}

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

		if rowIdx == p.cfg.IndexLineDate {
			for _, col := range p.cfg.ListColumnData {
				v := rowValue(row, col)
				raw[col] = v
				formatted[col] = formatDatePTBR(v)
			}
		}

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

func (p *Processor) plannedItemFromRow(row []string) (PlannedItem, bool) {
	quantities := make(map[int]float64)
	hasProduction := false
	for _, col := range p.cfg.ListColumnData {
		qty := parseFloat(rowValue(row, col))
		quantities[col] = qty
		if qty > 0 {
			hasProduction = true
		}
	}
	if !hasProduction {
		return PlannedItem{}, false
	}
	return PlannedItem{
		Codigo:     normalizeCode(rowValue(row, p.cfg.IndexColumnCod)),
		Descricao:  rowValue(row, p.cfg.IndexColItemDesc),
		Quantities: quantities,
	}, true
}

func rowValue(row []string, colIdx int) string {
	if colIdx < 0 || colIdx >= len(row) {
		return ""
	}
	return row[colIdx]
}
