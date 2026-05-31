package pcppdf

type Option func(*Processor)

func WithRenderer(r Renderer) Option {
	return func(p *Processor) {
		p.renderer = r
	}
}

func WithMetadata(meta map[string]ProdutoInfo) Option {
	return func(p *Processor) {
		p.meta = meta
	}
}
