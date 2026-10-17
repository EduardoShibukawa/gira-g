# Gira-G (`gira`)

A tool that builds and maintains an evidence-backed architectural map of any software system — from a single repository to many — and serves selected slices of it to humans and AI coding agents.

> **Name origin:** *Gira-G* (*Gira* + *G* [a nota Sol na cifra musical] = *Girassol*). No terminal, executado diretamente como **`gira`**.

## Instalação

### Via `mise` (backend nativo `github:`)

```bash
mise use -g github:EduardoShibukawa/gira-g@0.1.0
```

Ou no `mise.toml` do projeto:

```toml
[tools]
"github:EduardoShibukawa/gira-g" = "0.1.0"
```

### Via `install.sh` (`curl`)

```bash
curl -fsSL https://raw.githubusercontent.com/EduardoShibukawa/gira-g/main/install.sh | sh
```

Detecta SO/arquitetura, baixa o tarball do release, confere o SHA-256 contra
o `checksums.txt` publicado e instala o binário `gira` (`/usr/local/bin`, ou
`~/.local/bin` sem root). Versão específica: `sh -s -- --version v0.1.0`.
Com cosign instalado, `--cosign` verifica o bundle Sigstore keyless antes.

> **Nota sobre `GITHUB_TOKEN` (opcional):** Downloads públicos não exigem token. Exporte `GITHUB_TOKEN="$(gh auth token)"` se desejar limites de rate autenticados na API do GitHub ou ao instalar a partir de forks privados.

## Quickstart Reproduzível e Resolução de Índice

O `gira` suporta dois modos de entrada:
1. **Zero-config (repositório único):** `gira init [dir]` gera `gira.yaml` fixado no `HEAD` local e indexa o repositório imediatamente.
2. **Multi-repositório (orientado por manifesto):** `gira index [manifest]` lê `gira.yaml` declarando múltiplos repositórios e commits.

**Onde o índice fica salvo (`--output` vs cache padrão):**
- Sem `--output`, o `gira` grava o Macro Map em `~/.cache/gira/indexes/<system>/macro-map.json` (ou `.gira/indexes/<system>/macro-map.json` como fallback).
- Quando `--index` é omitido nos comandos de consulta, o `gira` resolve automaticamente na ordem: `.gira/macro-map.json`, `macro-map.json` e o índice em `~/.cache/gira/indexes/`.
- Para fixar um caminho explícito no projeto (ex.: `.gira/macro-map.json` ou `.gira/index.json`), passe `--output .gira/macro-map.json` na indexação e `--index .gira/macro-map.json` nas consultas.

### Exemplo executável (zero-config ou fixture do repositório)

```bash
# Opção A: Zero-config em qualquer repositório Git (ex.: este próprio checkout)
gira init . --output .gira/macro-map.json
gira tour --format md --index .gira/macro-map.json
gira locate Order --format md --index .gira/macro-map.json

# Opção B: Inicializar um repositório local a partir da fixture testdata/dogfood/services
cp -r testdata/dogfood/services /tmp/gira-demo && git -C /tmp/gira-demo init -q && \
  git -C /tmp/gira-demo add -A && git -C /tmp/gira-demo -c user.name=demo -c user.email=demo@local commit -q -m "init"
gira init /tmp/gira-demo --output /tmp/gira-demo/.gira/macro-map.json
gira impact service:order --format md --index /tmp/gira-demo/.gira/macro-map.json
gira doc service:order --format md --index /tmp/gira-demo/.gira/macro-map.json
```

## Referência da CLI

```bash
# Indexação e Onboarding Zero-Config
gira init [dir] [--force] [--output <path>] [--offline] [--cache-dir <dir>]
gira index [manifest] [--output <path>] [--offline] [--cache-dir <dir>]

# Consultas de Arquitetura, Impacto e Evidência
gira impact <target_id> [--depth <int>] [--budget <tokens>] [--index <path>] [--format json|md]
gira impact --from-diff <base>..<head> [--repo <dir>] [--repo-id <id>] [--index <path>] [--format json|md]
gira slice <node_id> [--tier summary|contracts|deep] [--view full|architecture|governance] [--budget <tokens>] [--index <path>]
gira explain [--node <id>] [--edge <id>] [--repo <dir>] [--source|--no-source] [--index <path>] [--format json|md]
gira tour [--focus <node>] [--max-steps <n>] [--index <path>] [--format json|md]
gira flows <entry> [--index <path>] [--format json|md]
gira doc <service-id> --format md|json|html [--index <path>] [--checkout <dir>] [--budget <tokens>]
gira diff <index-a> <index-b> [--format json|md]
gira locate <query> [--index <path>] [--format json|md]

# Exportação, Visualização e Integração com Agentes (MCP & Skills)
gira export [--format md|html] [--output <path>] [--index <path>]
gira view [--output <dir>] [--index <path>] [--no-open]
gira skills install|manifest|check [--host claude|cursor|agents] [--dir <project>] [--index <path>] [--hook]
gira serve mcp [--index <path>]

# Governança, Verificação e Avaliação (F1–F5)
gira eval [--check <c>] [--expect <e>] [--cwd <d>] [--taxonomy <t>] [--control-total] [--token-report measured]
gira verify --check <c> --expect <e> [--cwd <d>] [--fingerprint <hex>] [--taxonomy <t>] [--exit <code>]
gira review [--findings-register <path>] [--fix-check --finding <id> --expect-content <s>] [--gc [--keep-resolved <n>]]
gira gate --status | --reverify [--state <path>] [--digest <hex>]
gira gate-lint [--strict] [--fail-on-ci] [--state <path>] [--digest <hex>]
```

## Documentação

- [Docs Site](https://eduardoshibukawa.github.io/gira-g/) — documentação completa, guias e referência da CLI
- [**CONTEXT.md**](CONTEXT.md): Glossário e definições de domínio do projeto.
- [**docs/adr/**](docs/adr/): Architectural Decision Records (ADRs).
- [**docs/prior-art/**](docs/prior-art/): Análises de referências técnicas (Graphify, Understand Anything).
- [**Mapa e Rastreamento de Decisões**](docs/decisions.md): Issues e roadmap no issue tracker local.
