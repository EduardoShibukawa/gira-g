---
mode: padrao
goal: "gira index nao aborta quando um arquivo JSON/YAML e malformado: o pipeline pula o arquivo com aviso (UnsupportedBoundary + stderr) e indexa o resto."
done_when: "go test ./internal/extractors/ -run TestPipelineSkipsMalformedDataFile -count=1"
touch: ["internal/extractors/pipeline.go"]
tests_frozen: ["internal/extractors/pipeline_skip_malformed_test.go"]
---
## Passos
- [ ] RED: teste de pipeline com e1.json JSONL + valid.json
- [ ] GREEN: em ProcessFiles, erro de extractor LayerDataFormat vira UnsupportedBoundary + warning em stderr
## Fora
- extractor continua retornando erro (testes existentes intactos); sem flag nova
