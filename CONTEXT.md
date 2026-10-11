# Gira-G (CLI: `gira`)

A tool that builds and maintains an evidence-backed architectural map of any software system — from a single repository to many — and serves selected slices of it to humans and AI coding agents.

## Language

### The map

**Target System**:
Any software system the harness maps, whether it lives in one repository or spans many.
_Avoid_: the codebase, the project, the estate

**Component**:
The unit of the Macro Map: a deployable Service, or a module/package boundary declared in a manifest or workspace. Never an arbitrary directory, class, or function.
_Avoid_: module, node, unit; not a C4 Component (an in-process part) — a Service maps to a C4 Container

**Service**:
A Component that is independently deployable.
_Avoid_: app, microservice, deployable

**Library**:
A Component that represents a declared module or package boundary in a manifest or workspace, consumed in-process.

**Infrastructure**:
A shared stateful datastore, message broker, or middleware (e.g. PostgreSQL, Kafka, Valkey) used by Components.
_Avoid_: resource, backend

**Contract**:
A boundary specification representing an interface, endpoint, RPC method, or message schema (e.g. gRPC service/method, HTTP route, Kafka topic definition).
_Avoid_: endpoint, spec, interface (when unqualified)

**Evidence Record**:
A structured record of how a fact was discovered, including its derivation origin (declared, extracted, inferred), locator (repo, commit, path, line range), deployment profile, and staleness status.
_Avoid_: source, citation, annotation

**Macro Map**:
The default view of a Target System: its Components, their interfaces, and the relationships between them.
_Avoid_: graph, code graph, architecture diagram

**Extractor**:
The part of the harness that reads one kind of artifact — a data format, a tool's files (a CI workflow, a deployment manifest), or a programming language — and emits facts with evidence for the Macro Map. Anything no Extractor covers is reported as unsupported, never guessed.
_Avoid_: parser, plugin, scanner, analyzer

### Evaluation corpora

**Reference Corpus**:
A public multi-service Target System, pinned to an exact commit in each of its repos, used for repeatable, shareable benchmarks that never require sending private code anywhere. It is evaluated both as published and as a Perturbed Variant.
_Avoid_: test repo, demo app, sample system

**Perturbed Variant**:
A Reference Corpus with names and a few relationships systematically altered, so that an answer from a model's memory of the public original is detectably wrong.
_Avoid_: fork, mutated corpus, obfuscated copy

### Change impact

**Change-maker**:
A developer about to change something in a part of a Target System they don't know well; the first person the harness is built for.
_Avoid_: user, developer (unqualified), engineer

**Impact Briefing**:
The step, before any edit, in which a person or coding agent works out what a proposed change touches.
_Avoid_: pre-change context, analysis, planning

**Impact Set**:
The outcome of an Impact Briefing: the affected Components, contracts and known consumers, each with evidence, plus what is explicitly unknown.
_Avoid_: blast radius, affected list, dependencies
