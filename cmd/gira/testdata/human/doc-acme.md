# Service Doc: acme (`service:acme`, Service)
- Locator: [acme@c0ffee01:cmd/acme/main.go#L1-L1]
- Ports: (unknown: no ports declared in the extracted manifests)
- Provenance: profiles go; origins extracted; statuses current

## Entrypoints (1)
- `goentry:acme:cmd/acme` main (package main, dir `cmd/acme`) [acme@c0ffee01:cmd/acme/main.go#L19-L19] (inferred)

## Symbols (1 packages, 2 types, 1 funcs)
- `gosym:acme:package:cmd/acme` package main (dir `cmd/acme`) [acme@c0ffee01:cmd/acme/main.go#L1-L1] (declared)
- `gosym:acme:type:cmd/acme.billingWiring` type billingWiring (struct) [acme@c0ffee01:cmd/acme/main.go#L15-L15] (extracted)
- `gosym:acme:type:cmd/acme.ordersServer` type ordersServer (struct) [acme@c0ffee01:cmd/acme/main.go#L11-L11] (extracted)
- `gosym:acme:func:cmd/acme.main` unexported func main [acme@c0ffee01:cmd/acme/main.go#L19-L19] (extracted)

## Contracts Provided (1)
- `contract:Orders` Orders via `rel:acme-provides-orders` (provides) [acme@c0ffee01:cmd/acme/main.go#L12-L12, acme@c0ffee01:cmd/acme/main.go#L20-L20]

## Contracts Consumed (1)
- `contract:Billing` Billing via `rel:acme-consumes-billing` (consumes) [acme@c0ffee01:cmd/acme/main.go#L25-L25]

## Dependencies: Imports (4)

### Intra-module (2)
- `goimport:acme:acme:cmd/acme/main.go#L7` example.com/acme/genproto/orders as pb [acme@c0ffee01:cmd/acme/main.go#L7-L7] (extracted)
- `goimport:acme:acme:cmd/acme/main.go#L8` example.com/acme/internal/slices as slices [acme@c0ffee01:cmd/acme/main.go#L8-L8] (extracted)

### External (1)
- `goimport:acme:acme:cmd/acme/main.go#L6` github.com/external/dep as dep [acme@c0ffee01:cmd/acme/main.go#L6-L6] (extracted)

### Stdlib (1)
- `goimport:acme:acme:cmd/acme/main.go#L4` fmt as fmt [acme@c0ffee01:cmd/acme/main.go#L4-L4] (extracted)

## Calls (7)
- `gocall:acme:acme:cmd/acme/main.go#L20-1` `gosym:acme:func:cmd/acme.main` --RegisterOrdersServer--> (unresolved: no target) (unresolved) [acme@c0ffee01:cmd/acme/main.go#L20-L20]
- `gocall:acme:acme:cmd/acme/main.go#L21-1` `gosym:acme:func:cmd/acme.main` --Println--> stub fmt.Println (external) [acme@c0ffee01:cmd/acme/main.go#L21-L21]
- `gocall:acme:acme:cmd/acme/main.go#L21-2` `gosym:acme:func:cmd/acme.main` --Hello--> `gosym:internal:func:internal/slices.Hello` (resolved) [acme@c0ffee01:cmd/acme/main.go#L21-L21]
- `gocall:acme:acme:cmd/acme/main.go#L22-1` `gosym:acme:func:cmd/acme.main` --Println--> stub fmt.Println (external) [acme@c0ffee01:cmd/acme/main.go#L22-L22]
- `gocall:acme:acme:cmd/acme/main.go#L22-2` `gosym:acme:func:cmd/acme.main` --Hello--> `gosym:internal:func:internal/slices.Hello` (resolved) [acme@c0ffee01:cmd/acme/main.go#L22-L22]
- `gocall:acme:acme:cmd/acme/main.go#L23-1` `gosym:acme:func:cmd/acme.main` --Process--> stub github.com/external/dep.Process (external) [acme@c0ffee01:cmd/acme/main.go#L23-L23]
- `gocall:acme:acme:cmd/acme/main.go#L25-1` `gosym:acme:func:cmd/acme.main` --Submit--> (unresolved: no target) (unresolved) [acme@c0ffee01:cmd/acme/main.go#L25-L25]

## Run & Test (repo manifests) (1)
- `service:task:check` check: `go test -race ./...` [acme@c0ffee01:mise.toml#L4-L6]

## Where to Change (1)
- `cmd/acme`

## Guided Tour (1 step)
1. acme (Service) (`service:acme`, `contract:Billing`, `contract:Orders`)
   - Why: `service:acme` is the focus of this tour.
   - Why, evidence: acme@c0ffee01:cmd/acme/main.go#L1-L1
   - Read: acme@c0ffee01:cmd/acme/main.go#L1-L1
   - Read: acme@c0ffee01:cmd/acme/main.go#L12-L12
   - Read: acme@c0ffee01:cmd/acme/main.go#L25-L25

## Flows from service:acme (1)
1. service:acme -> contract:Billing
   - consumes `contract:Billing` via `rel:acme-consumes-billing` [acme@c0ffee01:cmd/acme/main.go#L25-L25]

## Known Unknowns (1)
- `cmd/acme/helper.py` - no extractor registered for file extension [acme@c0ffee01:cmd/acme/helper.py#L2-L2] (unknown joins: `contract:Billing`)

