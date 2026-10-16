# Gira-G (`gira`)

A tool that builds and maintains an evidence-backed architectural map of any software system — from a single repository to many — and serves selected slices of it to humans and AI coding agents.

> **Name origin:** *Gira-G* (*Gira* + *G* [a nota Sol na cifra musical] = *Girassol*). No terminal, executado diretamente como **`gira`**.

## Instalação

O repositório é privado: o install e os downloads exigem um token com
leitura (`gh auth login` feito, ou PAT fino com `contents: read`).

### Via `mise` (backend nativo `github:`)

```bash
export GITHUB_TOKEN="$(gh auth token)"
mise use -g github:EduardoShibukawa/gira-g@0.1.0
```

Ou no `mise.toml` do projeto:

```toml
[tools]
"github:EduardoShibukawa/gira-g" = "0.1.0"
```

### Via `install.sh` (`curl`)

```bash
export GITHUB_TOKEN="$(gh auth token)"
curl -fsSL -H "Authorization: Bearer $GITHUB_TOKEN" \
  https://raw.githubusercontent.com/EduardoShibukawa/gira-g/main/install.sh | sh
```

Detecta SO/arquitetura, baixa o tarball do release, confere o sha256 contra
o `checksums.txt` publicado e instala o binário `gira` (`/usr/local/bin`, ou
`~/.local/bin` sem root). Versão específica: `sh -s -- --version v0.1.0`.
Com cosign instalado, `--cosign` verifica o bundle Sigstore keyless antes.

## CLI

```bash
gira index                          # Indexação determinística (gera .gira/index.json)
gira impact <target_id>             # Briefing de impacto (Impact Set)
gira slice <node_id>                # Recorte de contexto direcionado
gira explain <subject>              # Cadeia de evidência (Evidence Chain)
gira serve                          # Servidor MCP stdio para agentes
```

## Documentação

- [**CONTEXT.md**](CONTEXT.md): Glossário e definições de domínio do projeto.
- [**docs/adr/**](docs/adr/): Architectural Decision Records (ADRs).
- [**docs/prior-art/**](docs/prior-art/): Análises de referências técnicas (Graphify, Understand Anything).
- [**Mapa e Rastreamento de Decisões**](docs/decisions.md): Issues e roadmap no issue tracker local.
