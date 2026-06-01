FILE ?= resources/metalurgica.bicicletas/dFato.xlsm
SHEET ?= RELATORIO PP GERAL
META ?= resources/metalurgica.bicicletas/dim_Items.CSV
TEMPLATE ?= resources/templates/template.html
OUT ?= ../SPDF (out)/

DATE_COLS ?= 12,15,18,21,24,27,30
CODE_COL ?= 2
DESC_COL ?= 3
START_LINE ?= 5
DATE_LINE ?= 0
VERBOSE ?= false
LOG ?= true
LOGDIR ?= Logs/

INSPECT_ROWS ?= 5
INSPECT_COLS ?= 72

.PHONY: pcp inspect test

pdfs:
	go run ./cmd \
		-xls "$(FILE)" \
		-sheet "$(SHEET)" \
		-date-cols "$(DATE_COLS)" \
		-code-col "$(CODE_COL)" \
		-desc-col "$(DESC_COL)" \
		-start-line "$(START_LINE)" \
		-date-line "$(DATE_LINE)" \
		-meta "$(META)" \
		-template "$(TEMPLATE)" \
		-out "$(OUT)" \
		-verbose=$(VERBOSE) \
		-log=$(LOG) \
		-log-dir "$(LOGDIR)"

inspect:
	go run ./cmd \
		-xls "$(PCP)" \
		-sheet "$(SHEET)" \
		-inspect \
		-inspect-rows "$(INSPECT_ROWS)" \
		-inspect-cols "$(INSPECT_COLS)"

test:
	go test ./...
