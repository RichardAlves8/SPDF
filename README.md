

## Origem — Security PDF

Diariamente, o responsável pela emissão de ordens de produção separava entre uma e duas horas para realizar um processo inteiramente manual: copiar o código da planilha de programação, colar no template de ficha e imprimir — uma ficha por vez. Em períodos sazonais, essa carga podia ultrapassar as duas horas com facilidade podendo chegar de 200 até 500 ordens por dia.

### Resumo do processo manual

```mermaid
flowchart LR
    A[Planilha\nde programação] -->|Ctrl+C| B[Template\nde ficha]
    B -->|Ctrl+P| C[Ficha\nimpressa]
    C -.->|de uma em uma| A
```

Por se tratar de documentos destinados ao controle de produção, qualquer erro tinha impacto direto no processo fabril. No entanto, por ser algo manual, a tarefa era ao mesmo tempo repetitiva e suscetível a falhas operacionais.

E por isso precisava ser algo seguro para este processo, dali o nome SPDF.

O projeto nasceu da observação de que, apesar da sensibilidade dos documentos gerados, o fluxo era inteiramente automatizável — e aprimorável. Com a solução implementada, mais de 80% dos problemas operacionais foram eliminados. Falhas agora ocorrem apenas em situações de exceção muito raras, e a diferença em relação ao processo anterior é discrepante.
*Pode parecer algo simples, mas SPDF mudou a forma como trabalhavamos diariamente.*
## Agradecimentos

Ao time de Programação e Controle de Produção *(PCP)* pelo apoio durante o desenvolvimento —
os reais responsáveis por todo o processo. Foi um dos projetos mais divertidos
que já trabalhei em equipe, e eles mudaram a forma como enxergo colaboração
e desenvolvimento. Grande parceria. 


---

## Resumo e início do Projeto

```mermaid
flowchart TD
    cmd[cmd/main.go]
    cmd --> proc

    subgraph pkg/pcppdf
        in[Planilha] --> proc[processor.go]
        proc --> emit[emitter.go]
        emit --> rend[renderer.go]
    end

    rend --> pdf[PDF]
    res[resources/] --> rend
```

O `cmd/main.go` atua como ponto de entrada da CLI, orquestrando o pacote `pkg/pcppdf`. Dentro dele, o pipeline segue uma direção linear: o `processor.go` lê e normaliza os dados da planilha, o `emitter.go` converte esses dados em contexto de renderização, e o `renderer.go` gera o PDF final — consumindo os templates, estilos e fontes em `resources/`.

---

## Pré-requisitos

- **Go (Golang):** Versão 1.22 ou superior.
- **Bibliotecas:** `excelize/v2` para leitura de Excel e `gopdf` para geração de documentos. Via Go Modules.

Para instalar as dependências Go e garantir que o ambiente esteja pronto:
```bash
#Sincronizar dependências do projeto
go mod tidy

# Bibliotecas principais:
go get github.com/xuri/excelize/v2
go get github.com/signintech/gopdf
```

---

## Parâmetros principais

Configuráveis `Makefile` e flags CLI.

| Parâmetro | Tipo | Descrição |
|-----------|------|-----------|
| `XLS` | `string` | Caminho do arquivo `.xlsm` |
| `SHEET` | `string` | Nome da aba (ex: `"RELATÓRIO PP GERAL"`) |
| `DATE_COLS` | `[]int` | Índices das colunas com quantidades de produção (base zero) |
| `CODE_COL` | `int` | Índice da coluna do código do produto |
| `DESC_COL` | `int` | Índice da coluna da descrição |
| `START_LINE` | `int` | Linha onde começam os dados dos produtos |
| `DATE_LINE` | `int` | Linha onde estão os cabeçalhos de data |
| `LOG` | `bool` | Emite log como proxy de I/O |
| `LOG_DIR` | `string` | Caminho do arquivo de log a ser salvo |
| `VERBOSE` | `bool` | Exibe resultado detalhado no terminal após a execução |

---

## Estrutura do projeto

```
.
├── cmd/
│   └── main.go                  # Ponto de entrada da CLI; parse de flags; orquestra o pacote pcppdf
│
├── pkg/pcppdf/
│   ├── processor.go             # Leitura e iteração sobre a planilha; normalização de registros
│   ├── emitter.go               # Converte modelos em contexto de renderização para templates e PDF
│   ├── renderer.go              # Gera o PDF final; trata paginação e exportação
│   ├── metadata.go              # Extração e normalização de metadados (datas, códigos, descrições)
│   ├── format.go                # Formatação de datas, números e códigos de barras
│   ├── config.go                # Carregamento de configurações e caminhos (templates, fontes, render)
│   ├── options.go               # Flags e opções do módulo (diretório de saída, template a usar)
│   ├── types.go                 # Structs e tipos centrais do domínio (Produto, Lote, Ficha...)
│   ├── errors.go                # Erros customizados e helpers para propagação de falhas
│   └── *_test.go                # Testes unitários de parsing, formatação e integração
│
├── resources/
│   ├── templates/
│   │   ├── template.html        # Layout base do PDF
│   │   ├── style.css            # Estilos aplicados durante a renderização
│   │   ├── arial.ttf            # Fonte embarcada para render consistente
│   │   └── Code39AzaleaRegular3.ttf  # Fonte para geração de códigos de barras
│   └── meta/
│       ├── qsort.CSV            # Arquivo auxiliar para exemplos e testes
│       └── dados.CSV            # Arquivo auxiliar para exemplos e testes
│
├── go.mod
└── go.sum
```

---

## Exemplo de uso (Linux)

```bash
# Executar todos os testes
go test ./...

# Processar uma planilha de exemplo — Mais detalhes de parâmetros na tabela a cima
go run ./cmd \
    -xls "caminho da planilha" \
    -sheet "Descrição da aba em que se encontra os dados" \
    -date-cols "Coluna com dados de quantidade" \
    -code-col "Coluna de Código Item" \
    -desc-col "Coluna de Descrição" \
    -start-line "Linha de ínicio (para desconsiderar o cabeçalho)" \
    -date-line "Linha onde se encontram as datas" \
    -meta "Arquivo CSV de meta dados dos produtos (dim_Item)" \
    -template "template.html" \
    -out "Pasta de saída" \
    -verbose=true/false \
    -log=true/false \
    -log-dir "saída de log"
#Inspeçaõ de planilha, o uso desta função foi feito para garantir que
#Os parametros passados estavam de acordo com as necessidades (caso houvesse alguma alteração na planilha)
o run ./cmd \
    -xls "caminho da planilha" \
    -sheet "Descrição da aba" \
    -inspect \
    -inspect-rows "linhas para inspecionar" \
    -inspect-cols "colunas"
```

---

## Notas de desenvolvimento

> **Memória com Excelize:** utilize sempre `Rows()` com streaming. Nunca use `GetRows()` ou `GetCellValue()` dentro de loops grandes — ambos carregam a planilha inteira na memória e estouram a RAM em ambientes com recursos limitados.

### Observações Finais

Este repositório representa apenas uma parte do todo (no entanto, a mais importante, e que resolve o problema). O SPDF completo possui
interface gráfica e foi desenvolvido em Java com Apache POI, conexão JDBC para
metadados atualizados e arquivos de configuração *(strings de conexão e acesso a outros servidores)* criptografados em AES-GCM.

O Original também contem cache de arquivos metadados — necessário porque no
setor de produção é comum operadores abrirem o sistema repetidamente para emitir
inclusões (pedidos fora da programação semanal ou por encomenda).
