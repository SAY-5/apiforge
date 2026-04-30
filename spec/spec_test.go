package spec

import "testing"

const sample = `{
  "title": "demo", "version": "1.0",
  "paths": {
    "/users/{id}": {
      "get": {"operationId": "getUser", "responses": {"200": {"description": "ok"}}}
    },
    "/users": {
      "post": {"operationId": "createUser", "responses": {"201": {"description": "created"}}}
    }
  }
}`

func TestParseExtractsPathsAndOps(t *testing.T) {
	s, err := Parse([]byte(sample))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(s.Paths) != 2 {
		t.Errorf("paths=%d want 2", len(s.Paths))
	}
	ops := s.Operations()
	if len(ops) != 2 {
		t.Errorf("ops=%d want 2", len(ops))
	}
	// Deterministic order: /users POST before /users/{id} GET (path-then-method).
	if ops[0].Path != "/users" || ops[0].Method != "POST" {
		t.Errorf("first op: %+v", ops[0])
	}
}

func TestParseEmptyJSONReturnsEmptySpec(t *testing.T) {
	s, err := Parse([]byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Paths) != 0 {
		t.Errorf("expected empty paths, got %d", len(s.Paths))
	}
}
