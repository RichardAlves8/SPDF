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

type Renderer interface {
	Render(ctx context.Context, outputPath string, data TemplateData) error
}

type GoPDFRenderer struct {
	FontDir string
}

func (r GoPDFRenderer) Render(ctx context.Context, outputPath string, data TemplateData) error {
	pdf := gopdf.GoPdf{}
	// A4 Landscape: 841.89 x 595.28 points
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4Landscape})
	pdf.AddPage()

	pageW := gopdf.PageSizeA4Landscape.W // 841.89
	marginX := 20.0
	contentW := pageW - (marginX * 2) // 801.89

	// ── Load Fonts ────────────────────────────────────────────────────────────
	arialPath := filepath.Join(r.FontDir, "arial.ttf")
	if _, err := os.Stat(arialPath); err != nil {
		return fmt.Errorf("Fonte (.ttf) não encontrada em: %s", r.FontDir)
	}
	if err := pdf.AddTTFFont("arial", arialPath); err != nil {
		return fmt.Errorf("Erro ao carregar fonte: %v", err)
	}

	barcodeFontPath := filepath.Join(r.FontDir, "Code39AzaleaRegular3.ttf")
	if _, err := os.Stat(barcodeFontPath); err != nil {
		return fmt.Errorf("Fonte (.ttf) não encontrada em %s", r.FontDir)
	}
	if err := pdf.AddTTFFont("barcode", barcodeFontPath); err != nil {
		return fmt.Errorf("Erro ao carregar fonte: %v", err)
	}

	// ── Helpers ───────────────────────────────────────────────────────────────
	val := func(key string) string {
		return valueString(data[key])
	}

	drawCenteredText := func(text string, y float64, size int, font string) {
		pdf.SetFont(font, "", size)
		width, _ := pdf.MeasureTextWidth(text)
		x := (pageW - width) / 2
		pdf.SetXY(x, y)
		pdf.Cell(nil, text)
	}

	// centerTextInCol draws text centered horizontally within a column region.
	centerTextInCol := func(text string, colX, colW, y float64, size int, font string) {
		pdf.SetFont(font, "", size)
		tw, _ := pdf.MeasureTextWidth(text)
		x := colX + (colW-tw)/2
		pdf.SetXY(x, y)
		pdf.Cell(nil, text)
	}

	// ── 1. Header ─────────────────────────────────────────────────────────────
	// Mirrors: <h2> and <h3> in <header> with border-bottom
	drawCenteredText("Setor de Produção — Pratos Prontos", 28, 14, "arial")
	drawCenteredText(fmt.Sprintf("Ordem de Produção — Emissão: %s", val("EMISSAO")), 46, 11, "arial")

	// Border-bottom da header (border-bottom: 1px solid #ccc)
	pdf.SetStrokeColor(204, 204, 204)
	pdf.SetLineWidth(0.5)
	pdf.Line(marginX, 62, marginX+contentW, 62)

	// ── 2. Destaque Cinza — .destaque-cinza (padding: 15px, margin-bottom: 5px)
	// padding 15px em cima e embaixo = altura ~50pt para o texto + padding
	destY := 67.0
	destH := 50.0
	pdf.SetFillColor(68, 68, 68)
	pdf.RectFromUpperLeftWithStyle(marginX, destY, contentW, destH, "F")
	pdf.SetTextColor(255, 255, 255)

	productText := fmt.Sprintf("%s  |  %s", val("CODIGO"), val("PRODUTO"))
	pdf.SetFont("arial", "", 18)
	pw, _ := pdf.MeasureTextWidth(productText)
	pdf.SetXY((pageW-pw)/2, destY+(destH-18)/2)
	pdf.Cell(nil, productText)

	pdf.SetTextColor(0, 0, 0)

	// ── 3. Highlight Row — .highlight-row (background: #eee, border, 3 colunas centradas)
	// font-size base 10px no container; cada coluna usa text-align:center + vertical-align:middle
	hlY := destY + destH + 5
	hlH := 65.0
	colW := contentW / 3
	col1X := marginX
	col2X := marginX + colW
	col3X := marginX + colW*2

	pdf.SetFillColor(238, 238, 238)
	pdf.SetStrokeColor(204, 204, 204)
	pdf.SetLineWidth(0.5)
	pdf.RectFromUpperLeftWithStyle(marginX, hlY, contentW, hlH, "FD")

	// Separadores verticais entre colunas (border-left: 1px solid #ccc no .highlight-col)
	pdf.Line(col2X, hlY, col2X, hlY+hlH)
	pdf.Line(col3X, hlY, col3X, hlY+hlH)

	// Col 1 — DATA (label + valor)
	// label: font-size ~10px bold uppercase; valor: font-size 1.5em ≈ 15pt
	pdf.SetFont("arial", "", 10)
	labelW, _ := pdf.MeasureTextWidth("DATA")
	pdf.SetXY(col1X+(colW-labelW)/2, hlY+10)
	pdf.Cell(nil, "DATA")

	pdf.SetFont("arial", "", 18) // 1.5em de 12px base ≈ 18pt
	centerTextInCol(val("DIA"), col1X, colW, hlY+28, 18, "arial")

	// Col 2 — QUANTIDADE NECESSÁRIA
	pdf.SetFont("arial", "", 10)
	lbl2W, _ := pdf.MeasureTextWidth("QUANTIDADE NECESSÁRIA")
	pdf.SetXY(col2X+(colW-lbl2W)/2, hlY+10)
	pdf.Cell(nil, "QUANTIDADE NECESSÁRIA")

	pdf.SetFont("arial", "", 24) // 2em de 12px base = 24pt
	centerTextInCol(val("QUANTIDADE"), col2X, colW, hlY+28, 24, "arial")

	// Col 3 — CLASS/EST e TIPO (font-size: 13px bold)
	pdf.SetFont("arial", "", 11)
	lbl3a := fmt.Sprintf("CLASS/EST: %s", val("SECAO"))
	lbl3b := fmt.Sprintf("TIPO: %s", val("SUBGRUPO"))
	la3W, _ := pdf.MeasureTextWidth(lbl3a)
	lb3W, _ := pdf.MeasureTextWidth(lbl3b)
	pdf.SetXY(col3X+(colW-la3W)/2, hlY+20)
	pdf.Cell(nil, lbl3a)
	pdf.SetXY(col3X+(colW-lb3W)/2, hlY+38)
	pdf.Cell(nil, lbl3b)

	// ── 4. Campos manuais — .divider.row ──────────────────────────────────────
	manualY := hlY + hlH + 8
	pdf.SetFont("arial", "", 12)
	pdf.SetXY(marginX, manualY)
	pdf.Cell(nil, "DATA DE PRODUÇÃO: ________ / _________ / __________")
	pdf.SetXY(marginX, manualY+20)
	pdf.Cell(nil, "RESPONSÁVEL:")

	// ── 5. Barcode — .barcode-box e .barcode-number ───────────────────────────
	// barcode-box: border, background #eee, text-align center, padding 5px
	// barcode-number: font Courier 18px, letter-spacing 2px, text-align center
	bcY := manualY + 48
	bcH := 80.0
	bcX := 200.0
	bcW := 440.0

	pdf.SetFillColor(238, 238, 238)
	pdf.SetStrokeColor(204, 204, 204)
	pdf.SetLineWidth(0.5)
	pdf.RectFromUpperLeftWithStyle(bcX, bcY, bcW, bcH, "FD")

	barcodeVal := fmt.Sprintf("*%s*", strings.TrimSpace(val("EAN13")))
	pdf.SetFont("barcode", "", 50)
	bw, _ := pdf.MeasureTextWidth(barcodeVal)
	pdf.SetXY((pageW-bw)/2, bcY+8)
	pdf.Cell(nil, barcodeVal)

	// Número EAN13: font-size 18px, letra espaçada, centralizado
	// gopdf não suporta letter-spacing nativo — adicionamos espaços entre chars como aproximação
	ean13 := val("EAN13")
	ean13Spaced := ""
	for i, ch := range ean13 {
		if i > 0 {
			ean13Spaced += " "
		}
		ean13Spaced += string(ch)
	}
	pdf.SetFont("arial", "", 18)
	nw, _ := pdf.MeasureTextWidth(ean13Spaced)
	pdf.SetXY((pageW-nw)/2, bcY+bcH+6)
	pdf.Cell(nil, ean13Spaced)

	// ── 6. Assinaturas — .highlight-row (mesma classe, fundo cinza + 3 colunas) ──
	sigY := bcY + bcH + 28
	sigH := 65.0

	pdf.SetFillColor(238, 238, 238)
	pdf.SetStrokeColor(204, 204, 204)
	pdf.SetLineWidth(0.5)
	pdf.RectFromUpperLeftWithStyle(marginX, sigY, contentW, sigH, "FD")

	pdf.Line(col2X, sigY, col2X, sigY+sigH)
	pdf.Line(col3X, sigY, col3X, sigY+sigH)

	drawSigCol := func(label, line string, colX float64) {
		pdf.SetFont("arial", "", 10)
		lw, _ := pdf.MeasureTextWidth(label)
		pdf.SetXY(colX+(colW-lw)/2, sigY+10)
		pdf.Cell(nil, label)

		// linha de assinatura
		pdf.SetFont("arial", "", 11)
		uw, _ := pdf.MeasureTextWidth(line)
		pdf.SetXY(colX+(colW-uw)/2, sigY+42)
		pdf.Cell(nil, line)
	}

	drawSigCol("QUANTIDADE PRODUZIDA", "_________________________", col1X)
	drawSigCol("ACOMPANHADO POR", "_________________________", col2X)
	drawSigCol("CONFERIDO POR", "_________________________", col3X)

	// ── 7. Observações — .divider + .obs-box ─────────────────────────────────
	// obs-box: border, height 40px
	obsY := sigY + sigH + 8
	pdf.SetFont("arial", "", 12)
	pdf.SetXY(marginX, obsY)
	pdf.Cell(nil, "OBSERVAÇÕES:")

	pdf.SetStrokeColor(204, 204, 204)
	pdf.SetLineWidth(0.5)
	pdf.RectFromUpperLeftWithStyle(marginX, obsY+14, contentW, 40, "D")

	// ── 8. Footer — font-size tiny, cor cinza claro, alinhado à direita ───────
	// Mirrors: footer { font-size: 0.3em; color: #a0a0a0; text-align: right }
	pdf.SetFont("arial", "", 7)
	pdf.SetTextColor(160, 160, 160)
	numText := val("NUM")
	ntw, _ := pdf.MeasureTextWidth(numText)
	pdf.SetXY(pageW-marginX-ntw, gopdf.PageSizeA4Landscape.H-12)
	pdf.Cell(nil, numText)
	pdf.SetTextColor(0, 0, 0)

	return pdf.WritePdf(outputPath)
}

type FileRenderer struct {
	TemplatePath string
}

func (f FileRenderer) Render(_ context.Context, outputPath string, data TemplateData) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	html, err := renderTemplateHTML(f.TemplatePath, data)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, []byte(html), 0o644)
}

func renderTemplateHTML(templatePath string, data TemplateData) (string, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}
	html := string(content)
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		html = strings.ReplaceAll(html, "{{"+key+"}}", valueString(data[key]))
	}
	return html, nil
}
