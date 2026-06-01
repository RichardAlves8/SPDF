package pcppdf

// Option define uma função que modifica as propriedades internas do Processor.
// Este padrão (Functional Options) permite uma inicialização flexível e extensível.
type Option func(*Processor)

// WithRenderer injeta um motor de renderização customizado no Processor.
// Útil para trocar o gerador de PDF por um gerador de HTML ou mock em testes.
func WithRenderer(r Renderer) Option {
	return func(p *Processor) {
		p.renderer = r
	}
}

// WithMetadata injeta um mapa de metadados pré-carregado no Processor.
// Evita que o Processor tente carregar o arquivo CSV por conta própria.
func WithMetadata(meta map[string]ProdutoInfo) Option {
	return func(p *Processor) {
		p.meta = meta
	}
}
