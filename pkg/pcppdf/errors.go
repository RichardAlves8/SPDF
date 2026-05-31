package pcppdf

import "errors"

var (
	ErrMissingTemplateHTML = errors.New("template HTML nao informado")
	ErrMissingPCPTabName   = errors.New("nome da aba PCP nao informado")
	ErrMissingDateColumns  = errors.New("colunas de data/quantidade nao informadas")
	ErrMissingOutputFolder = errors.New("pasta de saida PCP nao informada")
	ErrMissingMetadataCSV  = errors.New("CSV de metadados nao informado")
	ErrNilRenderer         = errors.New("renderer nao informado")
	ErrUnsupportedWorkbook = errors.New("formato de planilha nao suportado diretamente; use .xlsx/.xlsm ou converta antes")
)
