---
schema: task/1
tier: standard
status: active
parent: 2026-10-17-gira-versioned-publication
spec: 2026-10-17-gira-versioned-publication@1
depends_on: []
---

# Gira-G: integridade de clonagem e pré-requisitos Git/CLI

## Routing and approvals

- approved: escopo de correção e testes autorizado pelo dono em 2026-10-18, sob o parent heavy.
- status active: implementação está na PR #28; não marcar como concluída antes de CI verde no HEAD final e merge confirmado.

## Objective

Corrigir perdas de campos no clone do MacroMap e as falhas adjacentes de caminhos Git, processos e ordem de publicação do manifesto encontradas na auditoria.

## In scope

- Preservar UnresolvedRefs e Conflicts, incluindo cópias independentes das slices Evidence.
- Usar saída NUL-delimited para git ls-tree e preservar caminhos com tabs/newlines/backslashes onde o sistema de arquivos permite.
- Reap o processo tar em todos os caminhos de erro de git archive.
- Sanitizar variáveis de redirecionamento Git em todos os probes locais.
- Adiar criação/repin de gira.yaml até indexação bem-sucedida.
- Adicionar testes de regressão para cada comportamento acima.

## Out of scope

- Implementação de schema 2, source_identity, store versionado, locks e resolvedor. Esses itens pertencem aos children seguintes.

## Acceptance criteria

- [ ] Clone preserva todos os campos de diagnóstico e não compartilha slices Evidence com o original.
- [ ] ListTree preserva nomes de arquivo sem parsear newline/tab como delimitadores.
- [ ] ExportTree reap processos mesmo quando git archive falha.
- [ ] Probes Git locais ignoram variáveis de redirecionamento sem alterar env de clone/fetch.
- [ ] Falha de indexação em init não cria nem modifica gira.yaml, inclusive com --force.
- [ ] CI completa em Ubuntu/macOS/Windows, validator e docs build passam no HEAD final.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Testes cloneMacroMap e aliasing | not run | | |
| AC2 | Testes gitcache e CLI de falha | not run | | |
| AC3 | CI completa no HEAD final | not run | | |

## Checkpoint

- as of: 2026-10-18
- done: PR #28 aberta com implementação e testes
- next: aguardar CI no HEAD final; corrigir falhas e integrar PR #28
- open: nenhuma verificação do HEAD final foi confirmada nesta sessão
