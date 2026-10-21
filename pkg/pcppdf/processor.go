package pcppdf

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

// Processador lê a planilha de programação, extrai os dados de produção e
// orquestra a geração das fichas em PDF.
type Processador struct {
	configuracao Configuracao
	metadados    map[string]InformacaoProduto
	renderizador Renderizador
}

// NovoProcessador cria um Processador com a configuração informada. Ele valida
// a configuração e, se as opções não trouxerem os metadados, carrega o CSV
// indicado em Configuracao.CaminhoCSVMetadados.
//
// Sem a opção ComRenderizador, o Processador usa um RenderizadorArquivo sem
// caminho de template, que não consegue gravar fichas. Informe sempre um
// Renderizador.
func NovoProcessador(configuracao Configuracao, opcoes ...Opcao) (*Processador, error) {
	if err := configuracao.Validar(); err != nil {
		return nil, err
	}

	processador := &Processador{
		configuracao: configuracao,
		renderizador: RenderizadorArquivo{},
	}

	for _, opcao := range opcoes {
		opcao(processador)
	}

	if processador.renderizador == nil {
		return nil, ErrRenderizadorNulo
	}

	if processador.metadados == nil {
		metadados, err := CarregarCSVMetadados(configuracao.CaminhoCSVMetadados)
		if err != nil {
			return nil, err
		}
		processador.metadados = metadados
	}
	return processador, nil
}

// DefinirRenderizador troca o Renderizador em uso. Não é seguro chamá-lo
// enquanto outro método do Processador estiver em execução.
func (p *Processador) DefinirRenderizador(r Renderizador) {
	p.renderizador = r
}

// ProcessarArquivo é o ponto central de execução. Ele abre a planilha, percorre
// as linhas em streaming para economizar memória e entrega cada ordem válida
// ao Emissor.
//
// Uma falha ao emitir interrompe o processamento e devolve o resultado parcial
// junto com o erro. Produtos sem cadastro no CSV não interrompem: eles são
// registrados no resultado e as demais ordens seguem.
func (p *Processador) ProcessarArquivo(ctx context.Context, caminhoPlanilha string) (*ResultadoProcessamento, error) {
	if _, err := os.Stat(caminhoPlanilha); err != nil {
		return nil, err
	}

	arquivoExcel, err := excelize.OpenFile(caminhoPlanilha)
	if err != nil {
		return nil, err
	}
	defer arquivoExcel.Close()

	if indice, err := arquivoExcel.GetSheetIndex(p.configuracao.NomeAbaPCP); err != nil || indice < 0 {
		return nil, fmt.Errorf("aba PCP nao encontrada: %s", p.configuracao.NomeAbaPCP)
	}

	// Rows lê a aba linha a linha. GetRows carregaria a aba inteira na memória,
	// o que estoura a RAM em planilhas grandes.
	linhas, err := arquivoExcel.Rows(p.configuracao.NomeAbaPCP)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	datasBrutas, datasFormatadas, itens, err := p.lerLinhasDaPlanilha(linhas)
	if err != nil {
		return nil, err
	}

	resultado := &ResultadoProcessamento{
		LinhasComProducao: len(itens),
	}

	emissor := Emissor{
		configuracao: p.configuracao,
		metadados:    p.metadados,
		renderizador: p.renderizador,
	}

	dataEmissao := time.Now().Format("02/01/2006")
	for _, item := range itens {
		for _, coluna := range p.configuracao.ColunasDeData {
			quantidade := item.Quantidades[coluna]
			if quantidade <= 0 {
				continue
			}

			resultado.OrdensProcessadas++

			// Se a data não estiver no formato esperado, a formatação devolve
			// o texto original, e o valor bruto da planilha é usado.
			data := datasFormatadas[coluna]
			if data == "" {
				data = datasBrutas[coluna]
			}

			gerado, ausente, err := emissor.Emitir(ctx, PedidoDeEmissao{
				Codigo:        item.Codigo,
				Descricao:     item.Descricao,
				Quantidade:    quantidade,
				DataFormatada: data,
				DataEmissao:   dataEmissao,
				Numero:        resultado.FichasGeradas,
				Inclusao:      false,
			})
			if err != nil {
				return resultado, err
			}

			// O Emissor devolve o código quando o produto não existe no CSV.
			if ausente != "" {
				resultado.ProdutosAusentes = append(resultado.ProdutosAusentes, ausente)
				resultado.ItensAusentes = append(resultado.ItensAusentes, ItemAusente{
					Codigo:     ausente,
					Produto:    item.Descricao,
					Quantidade: quantidade,
					Data:       data,
				})
				continue
			}

			resultado.ArquivosGerados = append(resultado.ArquivosGerados, gerado)
			resultado.FichasGeradas++
		}
	}

	return resultado, nil
}

// lerLinhasDaPlanilha percorre a planilha uma única vez e devolve os títulos de
// data (brutos e formatados, indexados pela coluna) e os itens com produção.
func (p *Processador) lerLinhasDaPlanilha(linhas *excelize.Rows) (map[int]string, map[int]string, []ItemPlanejado, error) {
	brutas := make(map[int]string)
	formatadas := make(map[int]string)
	var itens []ItemPlanejado

	indiceLinha := 0
	for linhas.Next() {
		linha, err := linhas.Columns()
		if err != nil {
			return nil, nil, nil, err
		}

		if indiceLinha == p.configuracao.LinhaDasDatas {
			for _, coluna := range p.configuracao.ColunasDeData {
				valor := valorDaColuna(linha, coluna)
				brutas[coluna] = valor
				formatadas[coluna] = formatarDataPTBR(valor)
			}
		}

		if indiceLinha >= p.configuracao.LinhaInicialDosDados {
			item, temProducao := p.itemPlanejadoDaLinha(linha)
			if temProducao {
				itens = append(itens, item)
			}
		}

		indiceLinha++
	}
	if err := linhas.Error(); err != nil {
		return nil, nil, nil, err
	}
	return brutas, formatadas, itens, nil
}

// itemPlanejadoDaLinha converte uma linha da planilha em ItemPlanejado. O
// segundo retorno é falso quando a linha não tem quantidade maior que zero em
// nenhuma coluna de data, e a linha deve ser ignorada.
func (p *Processador) itemPlanejadoDaLinha(linha []string) (ItemPlanejado, bool) {
	quantidades := make(map[int]float64)
	temProducao := false

	for _, coluna := range p.configuracao.ColunasDeData {
		quantidade := converterDecimal(valorDaColuna(linha, coluna))
		quantidades[coluna] = quantidade
		if quantidade > 0 {
			temProducao = true
		}
	}

	if !temProducao {
		return ItemPlanejado{}, false
	}

	return ItemPlanejado{
		Codigo:      normalizarCodigo(valorDaColuna(linha, p.configuracao.ColunaCodigo)),
		Descricao:   valorDaColuna(linha, p.configuracao.ColunaDescricao),
		Quantidades: quantidades,
	}, true
}

// valorDaColuna devolve o valor da coluna indicada. Se o índice estiver fora
// da linha, devolve texto vazio em vez de causar pânico, porque uma linha lida
// em streaming pode ter menos colunas do que o layout configurado prevê.
func valorDaColuna(linha []string, indiceColuna int) string {
	if indiceColuna < 0 || indiceColuna >= len(linha) {
		return ""
	}
	return linha[indiceColuna]
}
