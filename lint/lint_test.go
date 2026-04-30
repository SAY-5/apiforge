package lint

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

func TestRulePathLowercaseKebabCatchesUppercase(t *testing.T) {
	s := mkSpec(t, `{
		"paths": {
			"/Users": {"get": {"operationId": "x", "responses": {"200": {"description": "ok"}}}}
		}
	}`)
	findings := RulePathLowercaseKebab()(s)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != SeverityError {
		t.Errorf("expected error severity")
	}
}

func TestRuleSuccessResponseRequired(t *testing.T) {
	s := mkSpec(t, `{
		"paths": {
			"/x": {"get": {"operationId": "g", "responses": {"500": {"description": "err"}}}}
		}
	}`)
	findings := RuleHTTPSuccessResponseRequired()(s)
	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
}

func TestRulePathParamsDeclared(t *testing.T) {
	s := mkSpec(t, `{
		"paths": {
			"/users/{id}": {
				"get": {"operationId": "g", "responses": {"200": {"description": "ok"}}}
			}
		}
	}`)
	// {id} not declared → finding.
	findings := RulePathParamsDeclared()(s)
	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
}

func TestRuleMethodWordingMatchesPath(t *testing.T) {
	s := mkSpec(t, `{
		"paths": {
			"/x": {"get": {"operationId": "createX", "responses": {"200": {"description": "ok"}}}}
		}
	}`)
	findings := RuleHTTPMethodWordingMatchesPath()(s)
	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
}

func TestRunAggregatesAcrossRules(t *testing.T) {
	s := mkSpec(t, `{
		"paths": {
			"/Users/{id}": {
				"get": {"responses": {"500": {"description": "err"}}}
			}
		}
	}`)
	findings := Run(s, DefaultRules())
	// Expect: lowercase, success-response, operation-id, path-param.
	if len(findings) < 4 {
		t.Errorf("expected >=4 findings, got %d", len(findings))
	}
}
