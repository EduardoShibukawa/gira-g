---
mode: padrao
goal: "O docsite do gira não tem nenhum caso especial por nome de projeto; tour, workflows e título vêm da config do repo-alvo"
done_when: "mise exec -- go test ./internal/docsite/... ./internal/manifest/... && ! rg -n -i 'p-inga|pinga' internal/docsite cmd/gira --glob '!*_test.go'"
touch:
  - internal/docsite/tour.go
  - internal/docsite/landing.go
  - internal/docsite/types.go
  - internal/docsite/builder_test.go
  - tasks/T04/PLAN.md
  - internal/manifest/manifest.go
  - cmd/gira/docs_cmd.go
  - internal/docsite/no_special_case_test.go
tests_frozen:
  - internal/docsite/no_special_case_test.go
---
# T04 — docsite sem caso especial por nome (A3)

- Seção `docsite:` do gira.yaml (title, tour[], workflows[], pillars[]) lida em
  internal/manifest/manifest.go e passada em SiteConfig (cmd/gira/docs_cmd.go).
- Remover os ramos `TargetName == "p-inga"|"pinga"` de tour.go e landing.go
  (tour, cenários, pilares, FormatProjectTitle) e os comandos `bun .pinga/...`
  fixos. Sem `docsite:` → caminho genérico que já existe para outros alvos.
- Conteúdo p-inga vai para o gira.yaml do p-inga (T de outro repo).
- Fora: HTML de alvos sem config inalterado; nada de menção a pinga fora de _test.go.
