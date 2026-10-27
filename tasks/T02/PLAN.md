---
mode: bugfix
goal: gira outline lists Go funcs, methods (receiver), types, top-level consts/vars with lines; impact error names accepted ID format
done_when: mise exec -- go test ./cmd/gira/ ./internal/slices/ -run 'TestOutlineGoSymbols|TestImpactUnknownTargetHint' -count=1
touch:
  - cmd/gira/main.go
  - cmd/gira/outline_go_test.go
  - internal/slices/engine.go
  - internal/slices/impact_hint_test.go
tests_frozen:
  - cmd/gira/outline_go_test.go
  - internal/slices/impact_hint_test.go
---
# T02 — outline Go symbols

Root cause: runOutline renders only extractor nodes (Contract/Library);
Layer3GoAST emits architectural nodes without per-symbol metadata, so .go
files print only the header. Fix: for .go files, list symbols via go/parser.
Impact: unknown target error must state the accepted node-ID format.
