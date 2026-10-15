# Gira-G (`gira`)

A tool that builds and maintains an evidence-backed architectural map of any software system — from a single repository to many — and serves selected slices of it to humans and AI coding agents.

> **Name origin:** *Gira-G* (*Gira* + *G* [a nota Sol na cifra musical] = *Girassol*). No terminal, executado diretamente como **`gira`**.

## Instalação

```bash
curl -fsSL https://raw.githubusercontent.com/EduardoShibukawa/gira-g/main/install.sh | sh
```

Detecta SO/arquitetura, baixa o tarball do release, confere o sha256 contra
o `checksums.txt` publicado e instala o binário `gira` (`/usr/local/bin`, ou
`~/.local/bin` sem root). Versão específica: `sh -s -- --version v0.1.0`.

## CLI

```bash
gira impact <contract-or-service>   # Briefing de impacto (Impact Set)
gira map                            # Macro Map da arquitetura
gira slice <component>              # Recorte de contexto direcionado
gira index                          # Indexação determinística
gira serve                          # Servidor MCP stdio para agentes
```

## Documentação

- [**CONTEXT.md**](CONTEXT.md): Glossário e definições de domínio do projeto.
- [**docs/adr/**](docs/adr/): Architectural Decision Records (ADRs).
- [**docs/prior-art/**](docs/prior-art/): Análises de referências técnicas (Graphify, Understand Anything).
- [**Mapa e Rastreamento de Decisões**](docs/decisions.md): Issues e roadmap no issue tracker local.
