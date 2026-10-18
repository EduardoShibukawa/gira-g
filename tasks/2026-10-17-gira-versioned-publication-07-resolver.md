---
schema: task/1
tier: standard
status: todo
parent: 2026-10-17-gira-versioned-publication
spec: 2026-10-17-gira-versioned-publication@1
depends_on:
  - 2026-10-17-gira-versioned-publication-06-publish
---

# Gira-G: resolvedor fail-closed e compatibilidade de schema

## Objective
Resolver apenas uma versão compatível com manifesto e tool_id, com erros estáveis e sem tratar ausência/incompatibilidade como mapa vazio.

## In scope
- index_missing, index_tool_mismatch, index_schema_unsupported, index_corrupt, index_ambiguous.
- Schema 2 estrito; legado 1.0.0 só por --index explícito, com aviso, somente leitura.
- Nunca carregar legado automaticamente quando gira.yaml está presente; nunca downgrade/rewrite in-place.
- Remover fallback de eval ManifestPath = gira.yaml; retornar manifest_path_required quando overrides exigem manifesto real e ele não foi fornecido.
- Distinguir schema desconhecido de JSON corrompido.

## Acceptance criteria
- [ ] Cenários de ausência S1/S2 retornam index_missing, nunca [] como sucesso.
- [ ] Tool mismatch e ambiguidade falham fechado com erro estável.
- [ ] Tipo/schema desconhecido retorna index_schema_unsupported, distinto de index_corrupt.
- [ ] Legado só é aceito por --index explícito com aviso e sem reescrita.
- [ ] Eval exige manifesto real apenas quando precisa construir locators.
- [ ] Testes provam que reader antigo rejeita schema 2 e que nenhum caminho faz downgrade.

## Verification
| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Testes de compatibilidade schema/legacy | not run | | |
| AC2 | Testes resolver/eval e erros estáveis | not run | | |

## Dependencies
- Store versionado publicado.
