---
mode: mvp
goal: "gira capabilities --json lista as capacidades do protocolo de contexto que o gira atende"
done_when: "mise exec -- go test ./cmd/gira/ -run Capabilities"
touch:
  - cmd/gira/main.go
  - cmd/gira/capabilities.go
  - cmd/gira/capabilities_test.go
  - tasks/T05/PLAN.md
tests_frozen:
  - cmd/gira/capabilities_test.go
---
# T05 — gira capabilities --json

- Comando novo no switch de Run (cmd/gira/main.go): `case "capabilities"` →
  cmd/gira/capabilities.go imprime exatamente
  `{"schema_version":1,"capabilities":["outline","slice","impact","read"],"extras":["gira:tour","gira:docs"]}`.
- Só capacidades que já existem no main.go (outline/slice/impact/read).
- Fora: não mudar a saída dos outros comandos; nada de flags extras.
