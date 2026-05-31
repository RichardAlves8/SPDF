package pcppdf

type ProdutoInfo struct {
	Codigo   string
	Nome     string
	EAN13    string
	Secao    string
	Subgrupo string
}

type PlannedItem struct {
	Codigo     string
	Descricao  string
	Quantities map[int]float64
}

type GeneratedFile struct {
	Codigo     string
	Produto    string
	Quantidade float64
	Data       string
	Path       string
}

type MissingItem struct {
	Codigo     string
	Produto    string
	Quantidade float64
	Data       string
}

type ProcessResult struct {
	GeneratedFiles      []GeneratedFile
	MissingProducts     []string
	MissingItems        []MissingItem
	OrdersProcessed     int
	FormsGenerated      int
	LinesWithProduction int
}

type TemplateData map[string]any
