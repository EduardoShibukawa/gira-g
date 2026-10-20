# INTENT 0001: Fatiamento Semântico para Agentes de Código e Micro-Extratores de Símbolos

## 1. Contexto e Motivação
O **Gira-G** foi concebido para construir e manter um mapa arquitetural de sistemas de software apoiado em evidências, servindo fatias selecionadas para humanos e agentes de codificação por IA.

Nos benchmarks empíricos com LLMs locais e de nuvem no `duh-homelab`, comprovamos que o uso de LSPs tradicionais (como Eclipse JDTLS e Pyright) gera o **LSP Overhead Paradox**:
- O modelo local (`Ornith-1.5-9B`) dobrou o número de turnos (7 para 14) e saltou de **70.606 para 175.014 tokens de prompt** (+148%) devido ao despejo de árvores AST em JSON e dezenas de warnings cosméticos de linters.
- O modo "cego" (apenas `grep` e `read_file`), por outro lado, sofre de cegueira arquitetural, gastando tokens ao ler arquivos inteiros para descobrir assinaturas de métodos.

O Gira-G deve atuar como o **Semantic Context Middleware** definitivo: entregando contexto arquitetural de alto sinal, com orçamento estrito de tokens e zero ruído de protocolo.

---

## 2. Objetivos Principais

### 1. Formato Compacto para LLMs (`--format=caveman` / `--format=llm`)
- Estender os comandos `gira slice`, `gira impact` e `gira locate` para suportar o formato `--format=caveman` (ou `--format=llm`).
- Converter saídas estruturadas em texto puro conciso de 1 linha por fato:
  ```text
  Component: <Type>[<ID>] @ <File>
  Contracts:
    - <Method> <Route> (L<Start>-L<End>) -> returns <Type>
  Dependencies:
    - consumes <TargetComponent> @ <File>:L<Line>
  Impact:
    - Consumed by: <Consumers>
  ```
- Respeitar rigorosamente o teto orçamentário `--budget <tokens>`, aplicando truncamento seguro sem omitir contratos críticos.

### 2. Micro-Extrator de Assinaturas (`gira symbols <file>`)
- Implementar o comando `gira symbols <file>` que extrai definições de classes, métodos, interfaces e propriedades de um arquivo alvo sem exigir um Language Server em execução.
- Suporte inicial a:
  - **Java**: classes, interfaces, métodos anotados (`@GetMapping`, `@PostMapping`, etc.).
  - **Go**: structs, interfaces, funções e métodos receptores.
  - **TypeScript/JavaScript**: types, interfaces, funções exportadas e classes.
- Custo por chamada: **< 50 tokens** (em vez de 1.500 tokens ao ler o arquivo inteiro).

### 3. Filtro de Diagnósticos Estrito (`--errors-only`)
- Permitir checagens de integridade arquitetural e de código que descartam 100% de `Warning`, `Information` e `Hint` de linters, retornando apenas erros críticos de compilação em 1 linha por erro:
  `[ERROR] <File>:<Line>: <Message>`

---

## 3. Critérios de Aceite
- [ ] Flag `--format=caveman` disponível em `gira slice`, `gira impact` e `gira locate`.
- [ ] Comando `gira symbols <file>` extrai assinaturas em < 20ms para arquivos de até 1.000 linhas.
- [ ] Saída formatada consome menos de 400 tokens para componentes médios.
- [ ] 100% de aprovação na suíte de testes em Go (`go test ./...`).
- [ ] Zero dependências CGO ou binários externos pesados.
