package pcppdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// PedidoDeEmissao reúne os dados necessários para gerar uma única ordem de
// produção. Ele é o contrato entre o Processador e o Emissor.
type PedidoDeEmissao struct {
	Codigo        string  // código do produto, como veio da planilha
	Descricao     string  // descrição lida da planilha
	Quantidade    float64 // quantidade a produzir
	DataFormatada string  // data legível, como "Sexta — 29/05/2026"
	DataEmissao   string  // data de geração do documento
	Numero        int     // sequencial da ordem
	Inclusao      bool    // indica uma ordem extra, emitida fora da programação
}

// Emissor completa os dados básicos de uma ordem com os metadados do produto
// (EAN13, seção e subgrupo) e coordena a gravação do arquivo por meio do
// Renderizador.
type Emissor struct {
	configuracao Configuracao
	metadados    map[string]InformacaoProduto
	renderizador Renderizador
}

// Emitir realiza o ciclo completo de emissão de uma ordem:
//
//  1. busca os metadados do produto;
//  2. monta os dados do template;
//  3. define o caminho de saída e chama o Renderizador.
//
// O segundo retorno só é preenchido quando o produto não existe nos
// metadados. Nesse caso, ele contém o código normalizado, nenhum arquivo é
// gravado e o erro é nulo, para que o Processador registre a ausência sem
// interromper a execução.
func (e Emissor) Emitir(ctx context.Context, pedido PedidoDeEmissao) (ArquivoGerado, string, error) {
	codigo := normalizarCodigo(pedido.Codigo)

	produto, encontrado := e.buscarProduto(codigo)
	if !encontrado {
		return ArquivoGerado{}, codigo, nil
	}

	dados := DadosDoTemplate{
		"CODIGO":     produto.Codigo,
		"PRODUTO":    pedido.Descricao,
		"EAN13":      produto.EAN13,
		"SECAO":      produto.Secao,
		"SUBGRUPO":   produto.Subgrupo,
		"DIA":        pedido.DataFormatada,
		"EMISSAO":    pedido.DataEmissao,
		"NUM":        numeroDaOrdem(pedido.Numero, pedido.Inclusao),
		"QUANTIDADE": pedido.Quantidade,
	}

	caminho, err := e.caminhoDeSaida(codigo, pedido.DataFormatada, pedido.Inclusao)
	if err != nil {
		return ArquivoGerado{}, "", err
	}

	if err := e.renderizador.Renderizar(ctx, caminho, dados); err != nil {
		return ArquivoGerado{}, "", err
	}

	return ArquivoGerado{
		Codigo:     produto.Codigo,
		Produto:    pedido.Descricao,
		Quantidade: pedido.Quantidade,
		Data:       pedido.DataFormatada,
		Caminho:    caminho,
	}, "", nil
}

// caminhoDeSaida calcula o caminho do arquivo e cria a pasta da data, se ela
// ainda não existir. O nome segue o padrão "CODIGO [DATA].pdf".
func (e Emissor) caminhoDeSaida(codigo string, data string, inclusao bool) (string, error) {
	base := e.configuracao.PastaOrdensPCP
	if inclusao && e.configuracao.PastaOrdensInclusao != "" {
		base = e.configuracao.PastaOrdensInclusao
	}

	// Converte "Sexta — 29/05/2026" em "29.05.2026" para uso em pasta e arquivo.
	//
	// TODO: converter a data sem perder desempenho; o custo por ficha ainda é alto.
	dataParaCaminho := dataNumericaParaCaminho(data)

	pasta := filepath.Join(base, dataParaCaminho)
	if err := os.MkdirAll(pasta, 0o755); err != nil {
		return "", err
	}

	nome := fmt.Sprintf("%s [%s].pdf", codigo, dataParaCaminho)
	return filepath.Join(pasta, nome), nil
}

// numeroDaOrdem formata o número impresso no rodapé da ficha. Para ordens de
// inclusão, ele é o texto "inclusao".
func numeroDaOrdem(numero int, inclusao bool) string {
	if inclusao {
		return "inclusao"
	}
	return fmt.Sprint(numero)
}

// buscarProduto procura o produto nos metadados. Se o código não existir como
// veio, tenta uma segunda vez com o dígito verificador separado por hífen,
// porque a planilha pode trazer "1234" enquanto o CSV guarda "123-4".
func (e Emissor) buscarProduto(codigo string) (InformacaoProduto, bool) {
	produto, encontrado := e.metadados[codigo]
	if encontrado {
		return produto, true
	}

	comDigito := formatarCodigoComDigito(codigo)
	if comDigito == codigo {
		return InformacaoProduto{}, false
	}
	produto, encontrado = e.metadados[comDigito]
	return produto, encontrado
}

// formatarCodigoComDigito transforma "1234" em "123-4". Códigos com menos de
// dois caracteres ou que já têm hífen são devolvidos sem alteração.
func formatarCodigoComDigito(codigo string) string {
	if len(codigo) < 2 || contemHifen(codigo) {
		return codigo
	}
	return codigo[:len(codigo)-1] + "-" + codigo[len(codigo)-1:]
}

// contemHifen informa se o texto tem pelo menos um hífen.
func contemHifen(texto string) bool {
	for _, caractere := range texto {
		if caractere == '-' {
			return true
		}
	}
	return false
}
