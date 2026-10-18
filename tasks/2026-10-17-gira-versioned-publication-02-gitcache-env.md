---
schema: task/1
tier: standard
status: todo
parent: 2026-10-17-gira-versioned-publication
spec: 2026-10-17-gira-versioned-publication@1
depends_on: []
---

# Gira-G: leituras Git determinísticas e segurança de caminhos

## Routing and approvals
- Child autorizado pelo dono em 2026-10-18 sob o parent heavy.
- PR #28 contém correções adjacentes já propostas; não duplicar trabalho e confirmar que os commits estão integrados antes de iniciar alterações dependentes.

## Objective
Tornar leituras locais Git determinísticas e garantir que paths impossíveis de materializar sejam tratados como falha, não como sucesso parcial.

## In scope
- Integrar/revalidar os fixes de ListTree NUL-delimited, ExportTree reapando tar e sanitização do ambiente de probes locais.
- Sanear variáveis de redirecionamento em todos os probes locais e em init, sem remover indiscriminadamente config/credenciais de clone/fetch.
- Cobrir core.quotePath, refs/replace, não ASCII, tab/newline quando o filesystem permite, GIT_DIR, GIT_WORK_TREE, GIT_COMMON_DIR, GIT_INDEX_FILE, GIT_OBJECT_DIRECTORY, GIT_ALTERNATE_OBJECT_DIRECTORIES, GIT_NAMESPACE, GIT_PREFIX e GIT_CEILING_DIRECTORIES.
- No Windows, se git archive não puder materializar um path, retornar erro contextual e falhar fechado; nunca aceitar extração parcial nem descartar arquivo silenciosamente.
- Cobrir erros de clone/fetch sem retirar configuração de usuário necessária a esses comandos.

## Out of scope
- Locator do manifesto, schema 2, locks e store versionado.

## Acceptance criteria
- [ ] Parsing git ls-tree -z preserva os bytes do pathname e falha com erro explícito em registro malformado.
- [ ] Falha de git archive sempre reap/processo consumidor e não deixa processo filho.
- [ ] Probes locais não são redirecionados por variáveis Git herdadas; clone/fetch mantêm comportamento compatível.
- [ ] Matriz quotePath/refs/replace não altera os bytes indexados.
- [ ] Windows: caminho não materializável resulta em falha explícita e nenhuma saída parcial é tratada como válida.
- [ ] Testes e CI passam nos três SOs; skips de symlink têm motivo registrado.

## Verification
| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Testes gitcache de caminhos e ambiente | not run | | |
| AC2 | Teste Windows de extração parcial/tab | not run | | |
| AC3 | CI Ubuntu/macOS/Windows, validator e floor guard | not run | | |

## Dependencies
- PR #28 integrada em main.
