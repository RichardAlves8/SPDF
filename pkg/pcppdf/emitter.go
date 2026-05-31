package pcppdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type EmitRequest struct {
	Codigo        string
	Descricao     string
	Quantidade    float64
	DataFormatada string
	DataEmissao   string
	Numero        int
	Inclusao      bool
}

type Emitter struct {
	cfg      Config
	meta     map[string]ProdutoInfo
	renderer Renderer
}

func (e Emitter) Emit(ctx context.Context, req EmitRequest) (GeneratedFile, string, error) {
	cod := normalizeCode(req.Codigo)
	produto, ok := e.lookupProduct(cod)
	if !ok {
		return GeneratedFile{}, cod, nil
	}

	data := TemplateData{
		"CODIGO":     produto.Codigo,
		"PRODUTO":    req.Descricao,
		"EAN13":      produto.EAN13,
		"SECAO":      produto.Secao,
		"SUBGRUPO":   produto.Subgrupo,
		"DIA":        req.DataFormatada,
		"EMISSAO":    req.DataEmissao,
		"NUM":        emitNumber(req.Numero, req.Inclusao),
		"QUANTIDADE": req.Quantidade,
	}

	outputPath, err := e.outputPath(cod, req.DataFormatada, req.Inclusao)
	if err != nil {
		return GeneratedFile{}, "", err
	}
	if err := e.renderer.Render(ctx, outputPath, data); err != nil {
		return GeneratedFile{}, "", err
	}

	return GeneratedFile{
		Codigo:     produto.Codigo,
		Produto:    req.Descricao,
		Quantidade: req.Quantidade,
		Data:       req.DataFormatada,
		Path:       outputPath,
	}, "", nil
}

func (e Emitter) outputPath(cod string, date string, inclusao bool) (string, error) {
	base := e.cfg.FolderOPPCP
	if inclusao && e.cfg.FolderOPInclusao != "" {
		base = e.cfg.FolderOPInclusao
	}
	dataArquivo := numericDateForPath(date)
	// Use numeric date for the folder name to avoid unwanted characters like "Sexta - "
	dir := filepath.Join(base, dataArquivo)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%s [%s].pdf", cod, dataArquivo)
	return filepath.Join(dir, name), nil
}

func emitNumber(n int, inclusao bool) string {
	if inclusao {
		return "inclusao"
	}
	return fmt.Sprint(n)
}

func (e Emitter) lookupProduct(cod string) (ProdutoInfo, bool) {
	produto, ok := e.meta[cod]
	if ok {
		return produto, true
	}
	formatted := formatCodeWithDigit(cod)
	if formatted == cod {
		return ProdutoInfo{}, false
	}
	produto, ok = e.meta[formatted]
	return produto, ok
}

func formatCodeWithDigit(cod string) string {
	if len(cod) < 2 || containsHyphen(cod) {
		return cod
	}
	return cod[:len(cod)-1] + "-" + cod[len(cod)-1:]
}

func containsHyphen(s string) bool {
	for _, r := range s {
		if r == '-' {
			return true
		}
	}
	return false
}
