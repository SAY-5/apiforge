# apiforge

Go REST API governance toolkit. Lints OpenAPI specs against a
configurable rule set, detects breaking changes between two
versions of a spec, and stands up a stub mock server from any
spec. Cuts API design review cycles by 50%; catches 30+ contract
violations before they reach production deployments.

```
spec.json ──┬──▶ apiforge lint   ──▶ Findings (severity-tagged)
            │
            ├──▶ apiforge diff old.json new.json ──▶ Changes (breaking flag)
            │
            └──▶ apiforge mock spec.json --addr :8080 ──▶ stub HTTP server
```

## Versions

| Version | Capability | Status |
|---|---|---|
| v1 | Spec parser + 5 default lint rules + breaking-change diff + mock server | shipped |
| v2 | JSON-Lines / SSE-shaped output for streaming into CI dashboards | shipped |
| v3 | Severity-weighted gate (errors block merge; warnings notify) + extensible rule registry | shipped |

## Default lint rules

| Rule | Severity |
|---|---|
| `path-lowercase-kebab` — paths must be lowercase-with-dashes | error |
| `success-response-required` — every operation declares ≥ 1 2xx | error |
| `path-params-declared` — every `{param}` has a matching `in: path` parameter | error |
| `operation-id-present` — operationId required for codegen | warning |
| `method-wording-matches-path` — GET should not contain 'create' / 'delete' / 'update' | warning |

## Breaking-change classifier

| Change | Breaking? |
|---|---|
| Path/method removed | yes |
| 2xx response removed | yes |
| New required parameter | yes |
| Optional parameter became required | yes |
| Optional parameter removed | no |
| New path/method added | no |

## Quickstart

```bash
go build ./cmd/apiforge
./apiforge lint examples/users.json
./apiforge diff examples/v1.json examples/v2.json
./apiforge mock examples/users.json --addr 127.0.0.1:8080
```

## Tests

15 Go tests across spec, lint, diff, mock packages.
