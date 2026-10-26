---
sidebar_position: 15
---

# `gira read`, `gira write`, and `gira add`

Deterministic symbol-level AST reading and mutation commands.

## Synopsis

```text
gira read <file> <symbol>
gira write <file> <symbol> [--body <body>]
gira add <file> <container> [--body <body>]
```

## Description

The symbol manipulation commands provide deterministic reading and surgical AST-level editing of individual functions, methods, and class members:

- `gira read`: Extracts only the lines corresponding to the requested function or method, minimizing context tokens.
- `gira write`: Surgically replaces an existing function or method AST node, preserving indentation and surrounding code without full-file rewrites.
- `gira add`: Inserts a new member or method inside a class or interface container with proper indentation and syntax validation.

## Examples

### Reading a Function

```bash
gira read src/engine.ts KanbanEngine.addColumn
```

### Writing a Function

```bash
gira write src/engine.ts KanbanEngine.addColumn --body '  addColumn(boardId: string, name: string): Column {
    return { id: "col_1", boardId, name, position: 0 };
  }'
```

### Adding a Member

```bash
gira add src/engine.ts KanbanEngine --body 'reset(): void {
  this.boards.clear();
}'
```
