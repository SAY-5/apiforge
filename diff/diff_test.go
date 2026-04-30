package diff

import (
	"testing"

	"github.com/SAY-5/apiforge/spec"
)

func mkSpec(t *testing.T, raw string) *spec.Spec {
	t.Helper()
	s, err := spec.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return s
}

func TestRemovedMethodIsBreaking(t *testing.T) {
	old := mkSpec(t, `{"paths": {"/x": {"get": {"responses": {"200": {"description": "ok"}}}}}}`)
	new := mkSpec(t, `{"paths": {"/x": {}}}`)
	changes := Compare(old, new)
	if !HasBreaking(changes) {
		t.Errorf("expected breaking change, got %+v", changes)
	}
}

func TestAddedMethodIsNotBreaking(t *testing.T) {
	old := mkSpec(t, `{"paths": {"/x": {}}}`)
	new := mkSpec(t, `{"paths": {"/x": {"get": {"responses": {"200": {"description": "ok"}}}}}}`)
	changes := Compare(old, new)
	if HasBreaking(changes) {
		t.Errorf("added method should not be breaking, got %+v", changes)
	}
}

func TestRequiredParamAddedIsBreaking(t *testing.T) {
	old := mkSpec(t, `{
		"paths": {"/x": {"get": {"responses": {"200": {"description": "ok"}}}}}
	}`)
	new := mkSpec(t, `{
		"paths": {"/x": {"get": {
			"parameters": [{"name": "filter", "in": "query", "required": true, "type": "string"}],
			"responses": {"200": {"description": "ok"}}
		}}}
	}`)
	changes := Compare(old, new)
	if !HasBreaking(changes) {
		t.Errorf("required param add should be breaking, got %+v", changes)
	}
}

func TestRemovedOptionalParamIsNotBreaking(t *testing.T) {
	old := mkSpec(t, `{
		"paths": {"/x": {"get": {
			"parameters": [{"name": "filter", "in": "query", "required": false, "type": "string"}],
			"responses": {"200": {"description": "ok"}}
		}}}
	}`)
	new := mkSpec(t, `{
		"paths": {"/x": {"get": {"responses": {"200": {"description": "ok"}}}}}
	}`)
	changes := Compare(old, new)
	if HasBreaking(changes) {
		t.Errorf("removing optional param should not be breaking, got %+v", changes)
	}
}

func TestRemoved2xxResponseIsBreaking(t *testing.T) {
	old := mkSpec(t, `{
		"paths": {"/x": {"get": {"responses": {"200": {"description": "ok"}, "404": {"description": "no"}}}}}
	}`)
	new := mkSpec(t, `{
		"paths": {"/x": {"get": {"responses": {"404": {"description": "no"}}}}}
	}`)
	changes := Compare(old, new)
	if !HasBreaking(changes) {
		t.Errorf("removing 2xx response should be breaking, got %+v", changes)
	}
}
