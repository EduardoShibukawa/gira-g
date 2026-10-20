# INTENT 0002: Análise de Impacto Determinística por Git Diff e Verificação de Contratos

## 1. Contexto e Motivação
Em pipelines de SDLC autônomos (como o Pi Harness em `duh-homelab/pi-harness`), agentes trabalham em worktrees Git efêmeros. Antes de fazer o merge no trunk (`main`), o harness precisa saber exatamente:
- Quais componentes e contratos foram tocados pelo diff?
- Quais consumidores downstream podem quebrar?
- Se existem suítes de testes que precisam ser disparadas além do Fast Test local.

O Gira-G deve prover a verificação de impacto determinística e apoiada em evidências diretamente a partir do `git diff`.

---

## 2. Objetivos Principais

### 1. Injeção de Contexto por Diff (`gira impact --from-diff`)
- Ler o diff entre dois commits (`<base>..<head>`) ou o working tree corrente.
- Identificar as entidades do Macro Map afetadas e os contratos modificados.
- Retornar o `Impact Set` em tempo constante (< 100ms) sem rescaneamento completo do repositório.

### 2. Validador de Quebra de Contratos (Breaking Changes Gate)
- Comparar contratos antes e depois do diff:
  - Renomeações de endpoints ou campos de request/response.
  - Exclusão de métodos ou alterações de assinatura de tipos.
- Emitir aviso em formato estruturado alertando o agente sobre potenciais quebras de API.

---

## 3. Critérios de Aceite
- [ ] `gira impact --from-diff` processa diffs de 50+ arquivos em < 100ms.
- [ ] Retorna lista clara de componentes afetados com evidências (`file:lines`).
- [ ] Saída compatível com `--format=json`, `--format=md` e `--format=caveman`.
- [ ] Cobertura de testes unitários em Go (`go test ./internal/...`).
