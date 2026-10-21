package pcppdf

// Opcao é uma função que altera propriedades internas do Processador. O
// padrão (functional options) permite inicializar o Processador de forma
// flexível, sem inflar a assinatura de NovoProcessador.
type Opcao func(*Processador)

// ComRenderizador define o Renderizador que o Processador usará. Serve para
// trocar o gerador de PDF por um gerador de HTML ou por um dublê em testes.
func ComRenderizador(r Renderizador) Opcao {
	return func(p *Processador) {
		p.renderizador = r
	}
}

// ComMetadados fornece um mapa de metadados já carregado, o que evita que o
// Processador leia o arquivo CSV por conta própria.
func ComMetadados(metadados map[string]InformacaoProduto) Opcao {
	return func(p *Processador) {
		p.metadados = metadados
	}
}
