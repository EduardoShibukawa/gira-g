---
schema: task/1
tier: standard
status: todo
parent: 2026-10-17-gira-versioned-publication
spec: 2026-10-17-gira-versioned-publication@1
depends_on:
  - 2026-10-17-gira-versioned-publication-01-adr-0014
  - 2026-10-17-gira-versioned-publication-02-gitcache-env
  - 2026-10-17-gira-versioned-publication-03-locator
---

# Gira-G: source_key, tool_id e source_identity

## Objective
Implementar identidade determinística de fonte e ferramenta, e propagar source_identity estritamente pelo modelo e todos os caminhos de cópia/serialização.

## In scope
- source_key a partir dos bytes crus exatos do manifesto que também são parseados.
- tool_id segue opção C: main.version, main.commit, GoVersion e dependências como campos explícitos; excluir -ldflags inteiro; hash do executável em builds de desenvolvimento não confiáveis.
- version_id = H(source_key, tool_id) com codificação injetiva.
- source_identity schema 2 embutido, manifest_sha256, versões de extractors e todos os sources ordenados deterministicamente, com url_sha256 e sem URLs cruas.
- Atualizar CanonicalJSON, clone/cópias e exportação para não perder nenhum campo.
- Decodificação estrita e teste de campos atuais e novos, incluindo UnresolvedRefs e Conflicts.

## Out of scope
- Locks e protocolo de publicação em disco.

## Acceptance criteria
- [ ] Contraexemplos da spec alteram a chave; teste/fuzz prova codificação injetiva.
- [ ] main.date isolado não altera tool_id de release; version/commit alteram tool_id.
- [ ] Builds dev não confiáveis incluem hash do executável.
- [ ] source_identity sobrevive CanonicalJSON, clone, store e snapshot.
- [ ] Schema 2 é inteiro e estrito; campos desconhecidos/inválidos não são descartados silenciosamente.
- [ ] Teste guard detecta divergência entre campos de MacroMap e campos tratados por clone/serialização.

## Verification
| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Testes e fuzz de chaves/tool_id | not run | | |
| AC2 | Matriz de propagação de campos | not run | | |
| AC3 | Testes schema 2 e legado | not run | | |

## Dependencies
- ADR 0014, Git/environment, locator.
