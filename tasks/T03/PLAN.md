---
mode: mvp
goal: "mise run install compila o gira com ldflags de versão (git describe), commit e data e grava em ~/.local/bin ou $GIRA_INSTALL_DIR"
done_when: "bash tools/test/install_test.sh"
touch: ["mise.toml", "README.md"]
tests_frozen: ["tools/test/install_test.sh"]
---
# T01: instalação local versionada

`version = "0.1.0-dev"` em cmd/gira/main.go só é sobrescrito pelo goreleaser.
Nova task `install` no mise.toml faz `go build -ldflags "-X main.version=...
-X main.commit=... -X main.date=..."` e grava em `${GIRA_INSTALL_DIR:-~/.local/bin}/gira`.
O teste aponta GIRA_INSTALL_DIR para um tmpdir (nunca toca o binário real).
README ganha uma linha documentando.
