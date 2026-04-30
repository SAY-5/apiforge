# Architecture

## Packages

```
spec/   parsed OpenAPI subset (paths, methods, params, responses)
lint/   rule engine + 5 default rules
diff/   breaking-change classifier
mock/   stub HTTP server that responds to declared operations
cmd/    CLI entry: lint / diff / mock subcommands
```

## Why a minimal OpenAPI subset

The full OpenAPI 3.x surface is huge (security schemes, callbacks,
links, encoding, examples). Most contract violations occur on a
small core: path naming, method/path coherence, parameter
declaration, response shape. We model that core and walk away
from the rest.

If you need full OpenAPI fidelity, swap in `getkin/kin-openapi`
or `pb33f/libopenapi`. The lint + diff engines consume the same
shape; the parse step is the only thing that changes.

## Lint rule contract

A `Rule` is `func(*spec.Spec) []Finding`. That's intentionally
simple — adding a custom rule for your team is one function. The
default set (5 rules) is the rule set most teams agree on; the
engine doesn't pretend to know your house style.

Severity choice:
- **error** for bugs that would let bad code merge (path conflicts,
  missing 2xx, undeclared path parameter)
- **warning** for style / future-pain (missing operationId,
  method-wording mismatch)

The CI gate exits non-zero on any error finding; warnings are
informational.

## Breaking-change classifier

The classifier walks both specs in parallel (`indexOps` + diff per
operation). The asymmetry is intentional:

- Removing **anything** a client currently relies on is breaking.
- Adding **anything optional** isn't — clients ignore unknown
  fields.
- Tightening (required-flag added, response code removed) is
  always breaking.

Production users wrap this with a "breaking-change allowed via
explicit semver bump" policy: a major-version bump can include
breaking changes; a minor cannot.

## Mock server

`mock.NewServer(spec)` returns a `*http.ServeMux` with one handler
per operation. Each handler responds with a canned JSON body
matching the lowest-numbered 2xx response declared on that
operation.

Go 1.22's path-pattern syntax (`"GET /users/{id}"`) makes this
trivial — no regex, no fancy router. Mock servers exist so the
frontend team can integrate before the backend is ready; they're
not meant to be production-faithful.

## What's deliberately not here

- **Schema validation** of request bodies. That's a runtime concern;
  lint catches the schema *being declared*, not *being correct*.
- **Mock server response shaping** that satisfies declared schemas.
  Returns a placeholder; production tools (Prism, Stoplight) do
  this fully.
- **Spectral compatibility**. Spectral uses JSONPath rules; we use
  Go functions. Easier to debug, less expressive. If you have a
  Spectral ruleset, port it to Go functions or call Spectral via
  its CLI from your CI.
