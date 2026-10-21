package pcppdf

// InformacaoProduto reúne os metadados de um produto, lidos do CSV de
// metadados.
type InformacaoProduto struct {
	Codigo   string // código identificador do produto
	Nome     string // nome completo ou descrição oficial
	EAN13    string // código de barras lido pelo coletor de dados da fábrica
	Secao    string // seção ou departamento, como "17"
	Subgrupo string // subgrupo ou tipo do produto, como "Pratos"
}

// ItemPlanejado representa uma linha da planilha de programação que tem
// produção em pelo menos uma das colunas de data.
type ItemPlanejado struct {
	Codigo      string          // código extraído da planilha
	Descricao   string          // descrição extraída da planilha
	Quantidades map[int]float64 // quantidade a produzir, indexada pela coluna
}

// ArquivoGerado descreve uma ficha em PDF gravada com sucesso.
type ArquivoGerado struct {
	Codigo     string  // código do produto
	Produto    string  // descrição usada na ficha
	Quantidade float64 // quantidade solicitada na ordem
	Data       string  // data da produção
	Caminho    string  // caminho do arquivo gravado
}

// ItemAusente registra um item que estava na planilha, mas não foi encontrado
// no CSV de metadados.
type ItemAusente struct {
	Codigo     string  // código lido na planilha
	Produto    string  // descrição lida na planilha
	Quantidade float64 // quantidade da ordem que deixou de ser emitida
	Data       string  // data da produção
}

// ResultadoProcessamento consolida as estatísticas de uma execução completa
// do Processador.
type ResultadoProcessamento struct {
	ArquivosGerados   []ArquivoGerado // todas as fichas gravadas
	ProdutosAusentes  []string        // códigos sem cadastro no CSV, um por ordem não emitida
	ItensAusentes     []ItemAusente   // detalhes dos itens sem cadastro, para o log
	OrdensProcessadas int             // ordens individuais encontradas, com ou sem cadastro
	FichasGeradas     int             // fichas gravadas com sucesso
	LinhasComProducao int             // linhas da planilha com alguma quantidade programada
}

// DadosDoTemplate associa o nome de cada campo ao valor que o Renderizador
// deve inserir na ficha.
type DadosDoTemplate map[string]any
