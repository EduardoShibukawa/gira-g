---
schema: task/1
tier: standard
status: todo
parent: 2026-10-17-gira-versioned-publication
spec: 2026-10-17-gira-versioned-publication@1
depends_on:
  - 2026-10-17-gira-versioned-publication-05-locks
  - 2026-10-17-gira-versioned-publication-07-resolver
---

# Gira-G: integrar init/index e semântica --output

## Routing and approvals

- approved: implementação desta etapa autorizada pelo dono em 2026-10-18 sob o parent heavy; os critérios de aceitação continuam obrigatórios.

## Objective
Integrar o store versionado à CLI sem publicar manifesto ou snapshots em estado parcial.

## In scope
- gira index sem --output publica no store canônico.
- index --output P grava apenas snapshot destacado e avisa que não é índice ativo.
- init publica versão canônica; init --output P grava cópia byte a byte e manifesto por último.
- init com falha não cria/altera manifesto; init --force rejeitado preserva manifesto/head/P.
- Se manifesto mudar durante init, retornar conflito.
- Publicação parcial de P informa o estado canônico com clareza.
- Implementar manifest_path_required e mensagens de durabilidade conforme spec.

## Acceptance criteria
- [ ] Testes CLI confirmam cada semântica de index/init --output.
- [ ] Build/gate falhando não cria nem repina manifesto.
- [ ] init --output P produz bytes canônicos e P idênticos.
- [ ] index --output P não altera store/head/manifesto.
- [ ] Manifesto alterado concorrentemente gera conflito, sem estado misto.
- [ ] Testes confirmam saída/exit code de visible_unconfirmed em Unix e Windows.

## Verification
| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Testes CLI integração | not run | | |
| AC2 | Testes de falha/concorrência | not run | | |

## Dependencies
- Locks e resolver.
