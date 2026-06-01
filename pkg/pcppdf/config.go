package pcppdf

// Config centraliza todos os parâmetros necessários para orquestrar o pipeline do SPDF.
// Estes valores são geralmente preenchidos através das flags de CLI no cmd/main.go.
type Config struct {
	TemplateHTML     string // Caminho para o template base (HTML/PDF)
	FontFileName     string // Nome da fonte principal a ser usada
	MetaCSVFile      string // Caminho para o CSV de metadados dos produtos
	PCPTabName       string // Nome da aba (Planilha) que contém a programação de produção
	ListColumnData   []int  // Índices das colunas que representam as datas de produção
	IndexColumnCod   int    // Índice da coluna onde fica o código do produto
	IndexColItemDesc int    // Índice da coluna onde fica a descrição do item
	IndexStartLine   int    // Linha onde os dados dos produtos efetivamente começam
	IndexLineDate    int    // Linha onde estão localizados os cabeçalhos de data
	FolderOPPCP      string // Diretório de saída para as ordens de produção padrão
	FolderOPInclusao string // Diretório de saída opcional para ordens de inclusão manual
}

// Validate verifica se as configurações mínimas obrigatórias foram fornecidas.
// Impede a execução se parâmetros vitais como a planilha ou pastas de saída estiverem ausentes.
func (c Config) Validate() error {
	if c.TemplateHTML == "" {
		return ErrMissingTemplateHTML
	}
	if c.PCPTabName == "" {
		return ErrMissingPCPTabName
	}
	if len(c.ListColumnData) == 0 {
		return ErrMissingDateColumns
	}
	if c.FolderOPPCP == "" {
		return ErrMissingOutputFolder
	}
	if c.MetaCSVFile == "" {
		return ErrMissingMetadataCSV
	}
	return nil
}
