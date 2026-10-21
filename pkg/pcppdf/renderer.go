package pcppdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/signintech/gopdf"
)

// Renderizador define o contrato dos motores que geram o documento final.
type Renderizador interface {
	// Renderizar transforma os dados já processados em um arquivo gravado em
	// caminhoSaida.
	Renderizar(ctx context.Context, caminhoSaida string, dados DadosDoTemplate) error
}

// RenderizadorGoPDF é o motor oficial de geração de PDF. Ele desenha o layout
// da ficha de produção linha a linha, o que garante precisão milimétrica e
// independência de drivers de impressão externos.
type RenderizadorGoPDF struct {
	PastaFontes string // pasta que contém as fontes arial.ttf e Code39AzaleaRegular3.ttf
}

// Renderizar desenha a ficha de produção em A4 paisagem e a grava em
// caminhoSaida. Ele devolve erro se alguma das duas fontes não for encontrada
// em PastaFontes.
//
// Os campos usados de dados são CODIGO, PRODUTO, EAN13, SECAO, SUBGRUPO, DIA,
// EMISSAO, NUM e QUANTIDADE.
func (r RenderizadorGoPDF) Renderizar(ctx context.Context, caminhoSaida string, dados DadosDoTemplate) error {
	pdf := gopdf.GoPdf{}
	// O A4 paisagem mede 841,89 x 595,28 pontos.
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4Landscape})
	pdf.AddPage()

	larguraPagina := gopdf.PageSizeA4Landscape.W
	margemX := 20.0
	larguraConteudo := larguraPagina - (margemX * 2)

	// As fontes são embarcadas no PDF para que o documento seja o mesmo em
	// qualquer computador (Windows, Linux ou Mac).
	caminhoArial := filepath.Join(r.PastaFontes, "arial.ttf")
	if _, err := os.Stat(caminhoArial); err != nil {
		return fmt.Errorf("Fonte (.ttf) não encontrada em: %s", r.PastaFontes)
	}
	if err := pdf.AddTTFFont("arial", caminhoArial); err != nil {
		return fmt.Errorf("Erro ao carregar fonte arial: %v", err)
	}

	// Fonte que transforma o texto em código de barras visual.
	caminhoFonteBarras := filepath.Join(r.PastaFontes, "Code39AzaleaRegular3.ttf")
	if _, err := os.Stat(caminhoFonteBarras); err != nil {
		return fmt.Errorf("Fonte de código de barras não encontrada em %s", r.PastaFontes)
	}
	if err := pdf.AddTTFFont("barcode", caminhoFonteBarras); err != nil {
		return fmt.Errorf("Erro ao carregar fonte de barcode: %v", err)
	}

	// valor devolve o campo dos dados como texto.
	valor := func(chave string) string {
		return valorComoTexto(dados[chave])
	}

	// desenharTextoCentralizado escreve o texto centralizado na largura da página.
	desenharTextoCentralizado := func(texto string, y float64, tamanho int, fonte string) {
		pdf.SetFont(fonte, "", tamanho)
		largura, _ := pdf.MeasureTextWidth(texto)
		x := (larguraPagina - largura) / 2
		pdf.SetXY(x, y)
		pdf.Cell(nil, texto)
	}

	// centralizarNaColuna escreve o texto centralizado dentro de uma coluna.
	centralizarNaColuna := func(texto string, colunaX, larguraColuna, y float64, tamanho int, fonte string) {
		pdf.SetFont(fonte, "", tamanho)
		larguraTexto, _ := pdf.MeasureTextWidth(texto)
		x := colunaX + (larguraColuna-larguraTexto)/2
		pdf.SetXY(x, y)
		pdf.Cell(nil, texto)
	}

	// Cabeçalho: título do setor, data de emissão e linha separadora.
	desenharTextoCentralizado("Departamento X — Setor Y", 28, 14, "arial")
	desenharTextoCentralizado(fmt.Sprintf("Ordem de Produção — Emissão: %s", valor("EMISSAO")), 46, 11, "arial")

	pdf.SetStrokeColor(204, 204, 204)
	pdf.SetLineWidth(0.5)
	pdf.Line(margemX, 62, margemX+larguraConteudo, 62)

	// Destaque do produto: faixa cinza-escura com código e descrição.
	destaqueY := 67.0
	destaqueAltura := 50.0
	pdf.SetFillColor(68, 68, 68)
	pdf.RectFromUpperLeftWithStyle(margemX, destaqueY, larguraConteudo, destaqueAltura, "F")
	pdf.SetTextColor(255, 255, 255)

	textoProduto := fmt.Sprintf("%s  |  %s", valor("CODIGO"), valor("PRODUTO"))
	pdf.SetFont("arial", "", 18)
	larguraProduto, _ := pdf.MeasureTextWidth(textoProduto)
	pdf.SetXY((larguraPagina-larguraProduto)/2, destaqueY+(destaqueAltura-18)/2)
	pdf.Cell(nil, textoProduto)

	// Volta para a cor preta.
	pdf.SetTextColor(0, 0, 0)

	// Informações principais, em três colunas: data, quantidade e classificação.
	blocoY := destaqueY + destaqueAltura + 5
	blocoAltura := 65.0
	larguraColuna := larguraConteudo / 3
	coluna1X := margemX
	coluna2X := margemX + larguraColuna
	coluna3X := margemX + larguraColuna*2

	pdf.SetFillColor(238, 238, 238)
	pdf.SetStrokeColor(204, 204, 204)
	pdf.SetLineWidth(0.5)
	pdf.RectFromUpperLeftWithStyle(margemX, blocoY, larguraConteudo, blocoAltura, "FD")

	// Linhas verticais que separam as colunas.
	pdf.Line(coluna2X, blocoY, coluna2X, blocoY+blocoAltura)
	pdf.Line(coluna3X, blocoY, coluna3X, blocoY+blocoAltura)

	// Coluna 1: data.
	pdf.SetFont("arial", "", 10)
	larguraRotuloData, _ := pdf.MeasureTextWidth("DATA")
	pdf.SetXY(coluna1X+(larguraColuna-larguraRotuloData)/2, blocoY+10)
	pdf.Cell(nil, "DATA")
	centralizarNaColuna(valor("DIA"), coluna1X, larguraColuna, blocoY+28, 18, "arial")

	// Coluna 2: quantidade.
	pdf.SetFont("arial", "", 10)
	larguraRotuloQuantidade, _ := pdf.MeasureTextWidth("QUANTIDADE NECESSÁRIA")
	pdf.SetXY(coluna2X+(larguraColuna-larguraRotuloQuantidade)/2, blocoY+10)
	pdf.Cell(nil, "QUANTIDADE NECESSÁRIA")
	centralizarNaColuna(valor("QUANTIDADE"), coluna2X, larguraColuna, blocoY+28, 24, "arial")

	// Coluna 3: classificação.
	pdf.SetFont("arial", "", 11)
	textoSecao := fmt.Sprintf("CLASS/EST: %s", valor("SECAO"))
	textoTipo := fmt.Sprintf("TIPO: %s", valor("SUBGRUPO"))
	larguraSecao, _ := pdf.MeasureTextWidth(textoSecao)
	larguraTipo, _ := pdf.MeasureTextWidth(textoTipo)
	pdf.SetXY(coluna3X+(larguraColuna-larguraSecao)/2, blocoY+20)
	pdf.Cell(nil, textoSecao)
	pdf.SetXY(coluna3X+(larguraColuna-larguraTipo)/2, blocoY+38)
	pdf.Cell(nil, textoTipo)

	// Campos de preenchimento manual: data de produção e responsável.
	preenchimentoY := blocoY + blocoAltura + 8
	pdf.SetFont("arial", "", 12)
	pdf.SetXY(margemX, preenchimentoY)
	pdf.Cell(nil, "DATA DE PRODUÇÃO: ________ / _________ / __________")
	pdf.SetXY(margemX, preenchimentoY+20)
	pdf.Cell(nil, "RESPONSÁVEL:")

	// Código de barras.
	barrasY := preenchimentoY + 48
	barrasAltura := 80.0
	barrasX := 200.0
	barrasLargura := 440.0

	pdf.SetFillColor(238, 238, 238)
	pdf.SetStrokeColor(204, 204, 204)
	pdf.SetLineWidth(0.5)
	pdf.RectFromUpperLeftWithStyle(barrasX, barrasY, barrasLargura, barrasAltura, "FD")

	// O Code39 exige asteriscos como caracteres de início e de fim, o que
	// também permite a leitura pelos coletores de dados.
	textoBarras := fmt.Sprintf("*%s*", strings.TrimSpace(valor("EAN13")))
	pdf.SetFont("barcode", "", 50)
	larguraBarrasTexto, _ := pdf.MeasureTextWidth(textoBarras)
	pdf.SetXY((larguraPagina-larguraBarrasTexto)/2, barrasY+8)
	pdf.Cell(nil, textoBarras)

	// Número do EAN13 legível, logo abaixo das barras.
	ean13 := valor("EAN13")
	ean13Espacado := ""
	for i, digito := range ean13 {
		if i > 0 {
			ean13Espacado += " "
		}
		ean13Espacado += string(digito)
	}
	pdf.SetFont("arial", "", 18)
	larguraEAN13, _ := pdf.MeasureTextWidth(ean13Espacado)
	pdf.SetXY((larguraPagina-larguraEAN13)/2, barrasY+barrasAltura+6)
	pdf.Cell(nil, ean13Espacado)

	// Assinaturas e conferência, em três colunas.
	assinaturasY := barrasY + barrasAltura + 28
	assinaturasAltura := 65.0

	pdf.SetFillColor(238, 238, 238)
	pdf.SetStrokeColor(204, 204, 204)
	pdf.RectFromUpperLeftWithStyle(margemX, assinaturasY, larguraConteudo, assinaturasAltura, "FD")

	pdf.Line(coluna2X, assinaturasY, coluna2X, assinaturasY+assinaturasAltura)
	pdf.Line(coluna3X, assinaturasY, coluna3X, assinaturasY+assinaturasAltura)

	// desenharColunaAssinatura escreve o rótulo e a linha de assinatura de uma coluna.
	desenharColunaAssinatura := func(rotulo, linha string, colunaX float64) {
		pdf.SetFont("arial", "", 10)
		larguraRotulo, _ := pdf.MeasureTextWidth(rotulo)
		pdf.SetXY(colunaX+(larguraColuna-larguraRotulo)/2, assinaturasY+10)
		pdf.Cell(nil, rotulo)

		pdf.SetFont("arial", "", 11)
		larguraLinha, _ := pdf.MeasureTextWidth(linha)
		pdf.SetXY(colunaX+(larguraColuna-larguraLinha)/2, assinaturasY+42)
		pdf.Cell(nil, linha)
	}

	desenharColunaAssinatura("QUANTIDADE PRODUZIDA", "_________________________", coluna1X)
	desenharColunaAssinatura("ACOMPANHADO POR", "_________________________", coluna2X)
	desenharColunaAssinatura("CONFERIDO POR", "_________________________", coluna3X)

	// Campo de observações.
	observacoesY := assinaturasY + assinaturasAltura + 8
	pdf.SetFont("arial", "", 12)
	pdf.SetXY(margemX, observacoesY)
	pdf.Cell(nil, "OBSERVAÇÕES:")
	pdf.RectFromUpperLeftWithStyle(margemX, observacoesY+14, larguraConteudo, 40, "D")

	// Rodapé: número da ordem, alinhado à direita.
	pdf.SetFont("arial", "", 7)
	pdf.SetTextColor(160, 160, 160)
	textoNumero := valor("NUM")
	larguraNumero, _ := pdf.MeasureTextWidth(textoNumero)
	pdf.SetXY(larguraPagina-margemX-larguraNumero, gopdf.PageSizeA4Landscape.H-12)
	pdf.Cell(nil, textoNumero)

	return pdf.WritePdf(caminhoSaida)
}

// RenderizadorArquivo é o motor alternativo que gera arquivos HTML ou de texto
// a partir de um arquivo-base, substituindo as variáveis no formato {{CHAVE}}.
type RenderizadorArquivo struct {
	CaminhoTemplate string // caminho do arquivo-base com os marcadores {{CHAVE}}
}

// Renderizar carrega o template, substitui os marcadores pelos valores de
// dados e grava o resultado em caminhoSaida, criando a pasta se necessário.
func (r RenderizadorArquivo) Renderizar(_ context.Context, caminhoSaida string, dados DadosDoTemplate) error {
	if err := os.MkdirAll(filepath.Dir(caminhoSaida), 0o755); err != nil {
		return err
	}
	conteudo, err := renderizarTemplateHTML(r.CaminhoTemplate, dados)
	if err != nil {
		return err
	}
	return os.WriteFile(caminhoSaida, []byte(conteudo), 0o644)
}

// renderizarTemplateHTML lê o template e substitui cada marcador {{CHAVE}} pelo
// valor correspondente em dados. Marcadores sem valor em dados permanecem no
// texto.
func renderizarTemplateHTML(caminhoTemplate string, dados DadosDoTemplate) (string, error) {
	bytesTemplate, err := os.ReadFile(caminhoTemplate)
	if err != nil {
		return "", err
	}
	conteudo := string(bytesTemplate)

	// As chaves são ordenadas para que a substituição seja sempre a mesma.
	chaves := make([]string, 0, len(dados))
	for chave := range dados {
		chaves = append(chaves, chave)
	}
	sort.Strings(chaves)

	for _, chave := range chaves {
		conteudo = strings.ReplaceAll(conteudo, "{{"+chave+"}}", valorComoTexto(dados[chave]))
	}
	return conteudo, nil
}
