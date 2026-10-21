package pcppdf

import "errors"

// Erros devolvidos pelo pacote. Use errors.Is para compará-los.
var (
	ErrTemplateHTMLAusente   = errors.New("template HTML nao informado")
	ErrNomeAbaPCPAusente     = errors.New("nome da aba PCP nao informado")
	ErrColunasDeDataAusentes = errors.New("colunas de data/quantidade nao informadas")
	ErrPastaDeSaidaAusente   = errors.New("pasta de saida PCP nao informada")
	ErrCSVMetadadosAusente   = errors.New("CSV de metadados nao informado")
	ErrRenderizadorNulo      = errors.New("renderer nao informado")
	ErrPlanilhaNaoSuportada  = errors.New("formato de planilha nao suportado diretamente; use .xlsx/.xlsm ou converta antes")
)
