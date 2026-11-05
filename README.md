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

### Local versionado (from source)

`mise run install` compila com a versão do `git describe --tags` e instala em `~/.local/bin/gira` (ou `$GIRA_INSTALL_DIR`).

### Via `install.sh` (`curl`)

```bash
curl -fsSL https://raw.githubusercontent.com/EduardoShibukawa/gira-g/main/install.sh | sh
```

Detecta SO/arquitetura, baixa o tarball do release, confere o SHA-256 contra
o `checksums.txt` publicado e instala o binário `gira` (`/usr/local/bin`, ou
`~/.local/bin` sem root). Versão específica: `sh -s -- --version v0.1.0`.
Com cosign instalado, `--cosign` verifica o bundle Sigstore keyless antes.

> **Nota sobre `GITHUB_TOKEN` (opcional):** Downloads públicos não exigem token. Exporte `GITHUB_TOKEN="$(gh auth token)"` se desejar limites de rate autenticados na API do GitHub ou ao instalar a partir de forks privados.

## Quickstart Reproduzível e Publicação Versionada de Índice

O `gira` suporta dois modos de entrada determinísticos:
1. **Zero-config (repositório único):** `gira init [dir]` inspeciona o repositório Git local, constrói o Macro Map e publica canonicamente no store versionado. O manifesto `gira.yaml` fixado no `HEAD` local é gravado por **último** — se o build ou gates falharem, nenhum manifesto quebrado fica no disco.
2. **Multi-repositório (orientado por manifesto):** `gira index [manifest]` lê `gira.yaml` declarando repositórios e commits fixados, publicando a versão imutável correspondente no store.

### Semântica de `--output` e Snapshots Destacados (D1)

- **`gira index [manifest] --output P`:** Grava exclusivamente um **snapshot destacado** (detached snapshot) no caminho `P`. Não altera o store versionado, não move ponteiros `.head` e não modifica `gira.yaml`. A CLI emite aviso em `stderr` alertando que `P` é um snapshot destacado e não o índice ativo.
- **`gira init [dir] --output P`:** Executa a publicação canônica no store versionado, atualiza `.head` da linhagem e grava uma cópia byte a byte idêntica do snapshot em `P`. O manifesto `gira.yaml` é gravado apenas após a conclusão bem-sucedida da publicação. Se a publicação canônica for concluída mas a cópia para `P` falhar, é reportado erro de publicação parcial.
- **Sem `--output`:** Tanto `gira init` quanto `gira index` publicam canonicamente no store versionado do usuário.

### Layout do Store Versionado e Hierarquia D2

A resolução do diretório-raiz do store de índices segue a ordem de precedência:
1. `$XDG_CACHE_HOME/gira/indexes`
2. `~/.cache/gira/indexes`
3. `<raiz-do-projeto>/.gira/indexes` (fallback)

Cada alvo indexado reside em uma linhagem imutável com versionamento atômico:
```text
<store_root>/
  <lineage_id>/
    .head                # Ponteiro atômico para o version_id ativo
    versions/
      <version_id>/
        macro-map.json   # Macro Map canônico imutável (Schema 2)
        meta.json        # Metadados estritos de publicação
```
Cada versão é endereçada determinística e unicamente por `version_id = SHA-256(source_key || tool_id)[:16]`, onde `source_key` é derivada dos bytes crus exatos do manifesto e `tool_id` identifica o binário `gira` e suas ferramentas de extração.

### Formato Canônico Schema 2 e Proveniência Embutida (`source_identity`)

O formato do índice utiliza `schema_version: 2` (inteiro). Tanto índices canônicos publicados quanto snapshots destacados carregam metadados autocontidos de proveniência:
```json
"source_identity": {
  "schema_version": 2,
  "manifest_sha256": "<64-hex SHA-256 dos bytes crus do manifesto>",
  "extractors": [
    { "id": "<extractor-id>", "version": "<semver>" }
  ],
  "sources": [
    {
      "repo_id": "<repo-id>",
      "sha": "<40-hex commit SHA>",
      "path_scope": "<path>",
      "url_sha256": "<hash do repositório>"
    }
  ]
}
```
URLs cruas com credenciais ou caminhos absolutos locais nunca são serializados.

### Política Estrita R2 e Isolamento do Legado Schema 1 (`"1.0.0"`)

- Índices legados com string `"1.0.0"` são estritamente isolados: **nunca** são carregados implicitamente quando um `gira.yaml` existe no projeto.
- O formato legado é aceito exclusivamente em modo somente leitura via `--index <path>` explícito, emitindo aviso visível em `stderr` (`[WARN] Reading legacy index format...`).
- Índices legados nunca são modificados nem atualizados in-place, e downgrades de schema são estritamente rejeitados pelo publicador.
- Leitores aplicam allowlist estrita: schemas desconhecidos ou mal tipados falham fechado com `index_schema_unsupported` (distinto de arquivo corrompido `index_corrupt`).

### Travas entre Processos (M -> S -> P) e Durabilidade

- **Exclusão mútua entre processos:** Escritores adquirem travas obrigatórias na sequência estrita:
  $$\mathbf{M} \text{ (Manifest Lock)} \longrightarrow \mathbf{S} \text{ (Store Lock)} \longrightarrow \mathbf{P} \text{ (Snapshot Lock)}$$
- **Primitivas nativas:** Utiliza `flock` (Linux/macOS) e `LockFileEx` (Windows) indexadas pela identidade física do descritor de arquivo (`os.SameFile`), imune a evasão por symlinks, hardlinks ou grafias relativas/absolutas.
- **Timeout uniforme de 60 segundos:** Concorrência não resolvida dentro de 60s retorna `index_locked`. Arquivos em sistemas de arquivos de rede sem suporte a travas retornam `lock_unavailable`.
- **Garantias de durabilidade:**
  - **Linux / macOS:** Sincronização atômica de arquivos e do diretório-pai (`fsync`) com status `confirmed`. Falhas de sync reportam `visible_unconfirmed` com código diferente de zero e cancelam gravação de `gira.yaml`.
  - **Windows:** Sem garantia de flush de metadados de diretório pelo Win32, reporta `visible_unconfirmed` com exit 0 e aviso explicativo em `stderr`.

### Taxonomia de Erros e Troubleshooting

| Código de Erro | Causa e Resolução |
|---|---|
| `index_missing` | Versão ou linhagem do índice não encontrada no store. Execute `gira index` ou `gira init`. |
| `index_tool_mismatch` | O índice em cache foi construído por versão incompatível do binário. Reindexe com o binário atual. |
| `index_schema_unsupported` | Versão de schema inválida ou não suportada. Atualize o índice para Schema 2. |
| `index_corrupt` | Arquivo truncado, JSON malformado ou hash divergente. Recupere ou recrie o índice. |
| `index_ambiguous` | Descoberta sem manifesto encontrou múltiplos índices concorrentes. Forneça `--index` ou execute em pasta com `gira.yaml`. |
| `manifest_outside_git_root` | Manifesto contendo overrides está fora da raiz Git. Mova o manifesto para dentro da worktree Git. |
| `git_root_unavailable` | Sonda de repositório Git falhou ou ownership é ambíguo. Verifique status e permissões do Git. |
| `manifest_path_required` | Avaliação de overrides requer o caminho real do manifesto; fallback implícito não é permitido. |
| `nondeterministic_build` | Re-indexação do mesmo `version_id` produziu bytes diferentes. Falha fechada sem sobrescrita. |
| `index_locked` | Timeout de concorrência (60s) atingido ao aguardar locks M/S/P. |
| `lock_unavailable` | Sistema de arquivos não suporta travas de arquivo obrigatórias (ex.: NFS/SMB/FUSE). |

### Exemplo executável (zero-config ou fixture do repositório)

```bash
# Opção A: Zero-config em qualquer repositório Git (publica no store canônico)
gira init .
gira tour --format md
gira locate Order --format md

# Opção B: Snapshot destacado para inspeção isolada sem alterar store ativo
gira index --output .gira/snapshot.json
gira impact service:order --format md --index .gira/snapshot.json

# Opção C: Inicializar a partir da fixture testdata/dogfood/services com cópia de snapshot
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
gira doc <service-id> [--format md|json|html] [--index <path>] [--checkout <dir>] [--budget <tokens>]
gira diff <index-a> <index-b> [--format json|md]
gira locate <query> [--index <path>] [--format json|md]

# Manipulação Cirúrgica de AST e Símbolos
gira outline <file> [--format text|json]
gira read <file> <symbol>
gira write <file> <symbol> [--body <body>]
gira add <file> <container> [--body <body>]

# Arquitetura Viva, Exportação e Integração com Agentes (MCP & Skills)
gira architecture [--format md|html|json] [--output <path>] [--index <path>]
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
