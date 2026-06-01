package pcppdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// EmitRequest contém os dados necessários para gerar uma única Ordem de Produção.
// Ele serve como o contrato entre o Processor e o Emitter.
type EmitRequest struct {
	Codigo        string
	Descricao     string
	Quantidade    float64
	DataFormatada string // Data legível (ex: Sexta — 29/05/2026)
	DataEmissao   string // Data de geração do documento
	Numero        int    // Sequencial da ordem
	Inclusao      bool   // Define se é uma ordem extra/manual
}

// Emitter é responsável por enriquecer os dados básicos com metadados (EAN13, Seção)
// e coordenar a gravação física do arquivo PDF através do Renderer.
type Emitter struct {
	cfg      Config
	meta     map[string]ProdutoInfo
	renderer Renderer
}

// Emitter realiza o ciclo completo de emissão de um documento:
// 1. Busca metadados do produto.
// 2. Prepara o mapa de dados (TemplateData).
// 3. Define o caminho de saída e invoca o Renderer.
func (e Emitter) Emit(ctx context.Context, req EmitRequest) (GeneratedFile, string, error) {
	cod := normalizeCode(req.Codigo)
	
	// Tenta localizar o produto no mapa de metadados carregado do CSV.
	produto, ok := e.lookupProduct(cod)
	if !ok {
		// Se não encontrar, retorna o código para que o Processor registre a falha.
		return GeneratedFile{}, cod, nil
	}

	// Monta o contexto que será injetado no PDF/Template.
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

	// local onde o PDF será salvo.
	outputPath, err := e.outputPath(cod, req.DataFormatada, req.Inclusao)
	if err != nil {
		return GeneratedFile{}, "", err
	}

	// renderização física do arquivo.
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

// outputPath calcula o caminho do arquivo e garante que a estrutura de pastas por data exista.
func (e Emitter) outputPath(cod string, date string, inclusao bool) (string, error) {
	base := e.cfg.FolderOPPCP
	if inclusao && e.cfg.FolderOPInclusao != "" {
		base = e.cfg.FolderOPInclusao
	}

	// Converte "Sexta — 29/05/2026" para "29.05.2026" para usar no nome da pasta e arquivo.
	//ainda há necessidade de traduzir as datas sem perca de processamento, o desafio do momento pois o custo em segundos está caro!
	dataArquivo := numericDateForPath(date)
	
	dir := filepath.Join(base, dataArquivo)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	// Nome do arquivo segue o padrão: CODIGO [DATA].pdf
	name := fmt.Sprintf("%s [%s].pdf", cod, dataArquivo)
	return filepath.Join(dir, name), nil
}

// emitNumber formata o número da ordem ou indica se é uma inclusão.
func emitNumber(n int, inclusao bool) string {
	if inclusao {
		return "inclusao"
	}
	return fmt.Sprint(n)
}

// lookupProduct busca o produto no mapa de metadados. 
// Possui uma lógica de fallback para tentar códigos com ou sem o dígito verificador separado por hífen.
func (e Emitter) lookupProduct(cod string) (ProdutoInfo, bool) {
	produto, ok := e.meta[cod]
	if ok {
		return produto, true
	}
	
	// Fallback: se o código veio "1234" mas no CSV está "123-4", tenta formatar.
	formatted := formatCodeWithDigit(cod)
	if formatted == cod {
		return ProdutoInfo{}, false
	}
	produto, ok = e.meta[formatted]
	return produto, ok
}

// formatCodeWithDigit tenta transformar "1234" em "123-4".
func formatCodeWithDigit(cod string) string {
	if len(cod) < 2 || containsHyphen(cod) {
		return cod
	}
	return cod[:len(cod)-1] + "-" + cod[len(cod)-1:]
}

// containsHyphen verifica se a string já possui um hífen.
func containsHyphen(s string) bool {
	for _, r := range s {
		if r == '-' {
			return true
		}
	}
	return false
}
