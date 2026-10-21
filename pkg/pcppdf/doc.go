// Pacote pcppdf transforma uma planilha de programação de produção em fichas
// de ordem de produção em PDF.
//
// O fluxo é linear e cada etapa fica a cargo de um tipo:
//
//   - Processador lê a planilha em streaming e separa as linhas que têm
//     produção programada;
//   - Emissor completa cada ordem com os metadados do produto (EAN13, seção e
//     subgrupo) e decide onde o arquivo será gravado;
//   - Renderizador desenha o documento e o grava em disco.
//
// Os metadados vêm de um CSV com ponto e vírgula como separador e cabeçalho
// CODIGO;PRODUTO;EAN13;SECAO;SUBGRUPO. A planilha precisa estar em formato
// .xlsx ou .xlsm.
//
// Um uso mínimo, com o renderizador oficial de PDF:
//
//	configuracao := pcppdf.Configuracao{ /* ... */ }
//	processador, err := pcppdf.NovoProcessador(configuracao,
//		pcppdf.ComRenderizador(pcppdf.RenderizadorGoPDF{PastaFontes: "resources/templates"}),
//	)
//	if err != nil {
//		// trate o erro
//	}
//	resultado, err := processador.ProcessarArquivo(ctx, "programacao.xlsm")
package pcppdf
