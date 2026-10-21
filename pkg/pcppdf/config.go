package pcppdf

// Configuracao reúne os parâmetros que descrevem o layout da planilha, os
// arquivos de apoio e as pastas de saída. Em geral, o comando spdf a preenche
// a partir das flags de linha de comando.
//
// Todos os índices de coluna e de linha começam em zero.
type Configuracao struct {
	CaminhoTemplateHTML  string // caminho do template base (HTML/PDF)
	NomeFontePrincipal   string // nome da fonte principal; reservado, ainda não é usado
	CaminhoCSVMetadados  string // caminho do CSV de metadados dos produtos
	NomeAbaPCP           string // aba que contém a programação de produção
	ColunasDeData        []int  // colunas com as quantidades de cada data de produção
	ColunaCodigo         int    // coluna do código do produto
	ColunaDescricao      int    // coluna da descrição do item
	LinhaInicialDosDados int    // primeira linha com produtos; as anteriores são cabeçalho
	LinhaDasDatas        int    // linha que contém os títulos de data
	PastaOrdensPCP       string // pasta de saída das ordens de produção padrão
	PastaOrdensInclusao  string // pasta de saída das ordens de inclusão manual (opcional)
}

// Validar verifica se os parâmetros obrigatórios foram informados. Ele devolve
// o primeiro erro encontrado, na ordem em que os campos aparecem no código.
func (c Configuracao) Validar() error {
	if c.CaminhoTemplateHTML == "" {
		return ErrTemplateHTMLAusente
	}
	if c.NomeAbaPCP == "" {
		return ErrNomeAbaPCPAusente
	}
	if len(c.ColunasDeData) == 0 {
		return ErrColunasDeDataAusentes
	}
	if c.PastaOrdensPCP == "" {
		return ErrPastaDeSaidaAusente
	}
	if c.CaminhoCSVMetadados == "" {
		return ErrCSVMetadadosAusente
	}
	return nil
}
