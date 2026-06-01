package pcppdf

// ProdutoInfo representa os metadados enriquecidos de um produto,
// vindos geralmente do arquivo CSV de metadados.
type ProdutoInfo struct {
	Codigo   string // Código identificador do produto
	Nome     string // Nome completo/Descrição oficial
	EAN13    string // Código de barras para o scanner da fábrica
	Secao    string // Seção/Departamento (ex: 17)
	Subgrupo string // Subgrupo/Tipo de produto (ex: Pratos)
}

// PlannedItem representa uma linha da planilha de programação que possui produção.
type PlannedItem struct {
	Codigo     string           // Código extraído da planilha
	Descricao  string           // Descrição extraída da planilha
	Quantities map[int]float64 // Mapeamento de Coluna -> Quantidade de produção
}

// GeneratedFile armazena o resultado de um PDF gerado com sucesso.
type GeneratedFile struct {
	Codigo     string  // Código do produto processado
	Produto    string  // Descrição utilizada
	Quantidade float64 // Quantidade solicitada na ordem
	Data       string  // Data da produção
	Path       string  // Caminho físico do arquivo .pdf gerado
}

// MissingItem registra informações de um item que estava na planilha
// mas não foi encontrado no arquivo de metadados (CSV).
type MissingItem struct {
	Codigo     string
	Produto    string
	Quantidade float64
	Data       string
}

// ProcessResult consolida as estatísticas e resultados de uma execução completa do processador.
type ProcessResult struct {
	GeneratedFiles      []GeneratedFile // Lista de todos os PDFs gerados
	MissingProducts     []string        // Apenas os códigos dos produtos faltantes
	MissingItems        []MissingItem   // Detalhes dos itens faltantes (para logs detalhados)
	OrdersProcessed     int             // Total de ordens individuais encontradas
	FormsGenerated      int             // Total de fichas PDF geradas com sucesso
	LinesWithProduction int             // Total de linhas da planilha que continham alguma produção
}

// TemplateData é um mapa de chave-valor usado para injetar dados nos templates HTML ou PDF.
type TemplateData map[string]any
